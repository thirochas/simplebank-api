package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/thirochas/simplebank-golang-api/internal/api/fake"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	repositoryfake "github.com/thirochas/simplebank-golang-api/internal/repository/fake"
	mockdb "github.com/thirochas/simplebank-golang-api/internal/repository/mock"
	"github.com/thirochas/simplebank-golang-api/internal/util"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogin(t *testing.T) {
	testCases := []struct {
		name          string
		body          createTokenRequest
		config        util.Config
		buildMocks    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "successfully-jwt",
			body: createTokenRequest{
				Username: "user1",
				Password: "secret",
			},
			config: util.Config{
				DBDriver:      "DB_DRIVER",
				DBSource:      "DB_SOURCE",
				ServerAddress: "SERVER_ADDRESS",
				SecretKey:     "FAKE_SECRET_KEY_WITH_32_CHARS_11",
				TokenType:     "jwt",
			},
			buildMocks: func(store *mockdb.MockStore) {
				user, err := repositoryfake.User()
				require.NoError(t, err)

				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyTokenMatches(t, recorder.Body)
			},
		},
		{
			name: "successfully-paseto",
			body: createTokenRequest{
				Username: "user1",
				Password: "secret",
			},
			config: fake.Config(),
			buildMocks: func(store *mockdb.MockStore) {
				user, err := repositoryfake.User()
				require.NoError(t, err)

				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyTokenMatches(t, recorder.Body)
			},
		},
		{
			name: "invalid-body",
			body: createTokenRequest{
				Username: "",
				Password: "",
			},
			config: fake.Config(),
			buildMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "non-existent-user",
			body: createTokenRequest{
				Username: "user1",
				Password: "secret",
			},
			config: fake.Config(),
			buildMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(repository.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "error-getting-user",
			body: createTokenRequest{
				Username: "user1",
				Password: "secret",
			},
			config: fake.Config(),
			buildMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(repository.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "invalid-password",
			body: createTokenRequest{
				Username: "user1",
				Password: "invalid-secret",
			},
			config: fake.Config(),
			buildMocks: func(store *mockdb.MockStore) {
				user, err := repositoryfake.User()
				require.NoError(t, err)

				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildMocks(store)

			server := NewServer(store, tc.config)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyTokenMatches(t *testing.T, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var tokenResponse createTokenResponse
	err = json.Unmarshal(data, &tokenResponse)
	require.NoError(t, err)
	require.NotEmpty(t, tokenResponse.AccessToken)
	require.WithinDuration(t, time.Now().Add(TokenDuration), tokenResponse.ExpireAt, time.Second)
}
