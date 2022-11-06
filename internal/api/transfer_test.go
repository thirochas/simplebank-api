package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/thirochas/simplebank-golang-api/internal/api/fake"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	mockdb "github.com/thirochas/simplebank-golang-api/internal/repository/mock"
	"github.com/thirochas/simplebank-golang-api/internal/token"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransfer(t *testing.T) {

	transfer := repository.Transfer{
		ID:            1,
		FromAccountID: 53,
		ToAccountID:   54,
		Amount:        70000,
	}

	testCases := []struct {
		name          string
		body          createTransferRequest
		buildMocks    func(store *mockdb.MockStore, body createTransferRequest)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "successfully",
			body: createTransferRequest{
				FromAccountID: 53,
				ToAccountID:   54,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(1).
					Return(repository.Account{ID: 53, Owner: "user", Currency: "BRL"}, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(1).
					Return(repository.Account{ID: 54, Owner: "user-2", Currency: "BRL"}, nil)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(transfer, nil)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyTransferMatches(t, recorder.Body, transfer)
			},
		},
		{
			name: "invalid currency for from account id",
			body: createTransferRequest{
				FromAccountID: 53,
				ToAccountID:   54,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(1).
					Return(repository.Account{ID: 53, Owner: "user", Currency: "USD"}, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(0)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "invalid currency for to account id",
			body: createTransferRequest{
				FromAccountID: 53,
				ToAccountID:   54,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(1).
					Return(repository.Account{ID: 53, Owner: "user", Currency: "BRL"}, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(1).
					Return(repository.Account{ID: 54, Owner: "user-2", Currency: "USD"}, nil)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "invalid request",
			body: createTransferRequest{
				FromAccountID: 53,
				ToAccountID:   54,
				Currency:      "",
				Amount:        0,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(0)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "transfer error",
			body: createTransferRequest{
				FromAccountID: 53,
				ToAccountID:   54,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(1).
					Return(repository.Account{ID: 53, Owner: "user", Currency: "BRL"}, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(1).
					Return(repository.Account{ID: 54, Owner: "user-2", Currency: "BRL"}, nil)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(repository.Transfer{}, errors.New("error message"))
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "account not found error",
			body: createTransferRequest{
				FromAccountID: 51,
				ToAccountID:   54,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(1).
					Return(repository.Account{}, sql.ErrNoRows)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(0)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "get account error",
			body: createTransferRequest{
				FromAccountID: 51,
				ToAccountID:   54,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(1).
					Return(repository.Account{}, sql.ErrConnDone)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(0)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "authenticated owner account different from 'account from owner'",
			body: createTransferRequest{
				FromAccountID: 53,
				ToAccountID:   54,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(1).
					Return(repository.Account{ID: 53, Owner: "user-4", Currency: "BRL"}, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.ToAccountID)).
					Times(0)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "authenticated owner account cannot transfer to himself",
			body: createTransferRequest{
				FromAccountID: 53,
				ToAccountID:   53,
				Currency:      "BRL",
				Amount:        70000,
			},
			buildMocks: func(store *mockdb.MockStore, body createTransferRequest) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(body.FromAccountID)).
					Times(2).
					Return(repository.Account{ID: 53, Owner: "user", Currency: "BRL"}, nil)

				store.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildMocks(store, tc.body)

			server := NewServer(store, fake.Config())
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyTransferMatches(t *testing.T, body *bytes.Buffer, account repository.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTranfer repository.Transfer
	err = json.Unmarshal(data, &gotTranfer)
	require.NoError(t, err)
	require.Equal(t, account, gotTranfer)
}
