package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/thirochas/simplebank-golang-api/internal/api/fake"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	repositoryfake "github.com/thirochas/simplebank-golang-api/internal/repository/fake"
	mockdb "github.com/thirochas/simplebank-golang-api/internal/repository/mock"
	"github.com/thirochas/simplebank-golang-api/internal/util"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type eqCreateUserMatcher struct {
	arg      repository.CreateUserParams
	password string
}

func (e eqCreateUserMatcher) Matches(x interface{}) bool {
	arg, ok := x.(repository.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg.HashedPassword, e.password)
}

func EqCreateUserParams(arg repository.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserMatcher{arg, password}
}

func TestCreateUser(t *testing.T) {
	user, err := repositoryfake.User()
	require.NoError(t, err)

	testsCases := []struct {
		name             string
		body             createUserRequest
		expectedResponse createUserResponse
		buildStubs       func(store *mockdb.MockStore)
		checkResponse    func(recorder *httptest.ResponseRecorder, expectedResponse createUserResponse)
	}{
		{
			name: "successfully",
			body: createUserRequest{
				Username: user.Username,
				Password: "secret",
				FullName: user.FullName,
				Email:    user.Email,
			},
			expectedResponse: createUserResponse{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := repository.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, "secret")).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder, expectedResponse createUserResponse) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyUserMatches(t, recorder.Body, expectedResponse)
			},
		},
		{
			name: "invalid-request",
			body: createUserRequest{
				Username: "",
				Password: "secret",
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder, expectedResponse createUserResponse) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "internal-server-error",
			body: createUserRequest{
				Username: user.Username,
				Password: "secret",
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(repository.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder, expectedResponse createUserResponse) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "unique-violation",
			body: createUserRequest{
				Username: user.Username,
				Password: "secret",
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := repository.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, "secret")).
					Times(1).
					Return(repository.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder, expectedResponse createUserResponse) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for i := range testsCases {
		tc := testsCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store, fake.Config())
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder, tc.expectedResponse)
		})
	}
}

func requireBodyUserMatches(t *testing.T, body *bytes.Buffer, user createUserResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount createUserResponse
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, user, gotAccount)
}
