package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github/leoflalv/bank-api/db/mock"
	db "github/leoflalv/bank-api/db/sqlc"
	"github/leoflalv/bank-api/token"
	"github/leoflalv/bank-api/util"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       int64(util.RandomNumber(1, 1000)),
		Owner:    owner,
		Balance:  util.RandomNumber(1, 10000),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var goAccount db.Account
	err = json.Unmarshal(data, &goAccount)
	require.Equal(t, account, goAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var goAccounts []db.Account
	err = json.Unmarshal(data, &goAccounts)
	for i := range goAccounts {
		require.Equal(t, accounts[i], goAccounts[i])
	}
}

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		accountId     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenManager token.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Ok",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountId: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newMockServer(t, store)
			SetupRoutes(server)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/account/%d", tc.accountId)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenManager)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})

	}

}

func TestListAccountsAPI(t *testing.T) {
	user, _ := randomUser()
	accounts := []db.Account{
		randomAccount(user.Username),
		randomAccount(user.Username),
		randomAccount(user.Username),
		randomAccount(user.Username),
		randomAccount(user.Username),
	}

	params := listAccountsRequest{Count: 1, Size: 5}
	args := db.ListAccountsParams{Owner: user.Username, Limit: 5, Offset: 0}

	testCases := []struct {
		name          string
		accountId     int64
		count         int32
		size          int32
		setupAuth     func(t *testing.T, request *http.Request, tokenManager token.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Ok",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Eq(args)).Times(1).Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
			count: params.Count,
			size:  params.Size,
		},
		{
			name: "InternalError",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Eq(args)).Times(1).Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			count: params.Count,
			size:  params.Size,
		},
		{
			name: "BadRequest",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			count: 0,
			size:  params.Size,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newMockServer(t, store)
			SetupRoutes(server)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts?count=%d&size=%d", tc.count, tc.size)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenManager)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})

	}

}

func TestDeleteAccountAPI(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		accountId     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenManager token.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Ok",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountId: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newMockServer(t, store)
			SetupRoutes(server)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/account/%d", tc.accountId)

			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenManager)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)
	arg := db.CreateAccountParams{Owner: account.Owner, Balance: 0, Currency: account.Currency}

	testCases := []struct {
		name          string
		accountId     int64
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, request *http.Request, tokenManager token.Manager)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
		body          gin.H
	}{
		{
			name:      "Ok",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
			body: gin.H{
				"balance":  account.Balance,
				"currency": account.Currency,
			},
		},
		{
			name:      "InternalServerError",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			body: gin.H{
				"balance":  account.Balance,
				"currency": account.Currency,
			},
		},
		{
			name:      "BadRequest",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			body: gin.H{
				"balance": account.Balance,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newMockServer(t, store)
			SetupRoutes(server)
			recorder := httptest.NewRecorder()

			url := "/account"

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenManager)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateAccountAPI(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)
	updatedAccount := db.Account{
		ID:        account.ID,
		Owner:     account.Owner,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt,
		Balance:   account.Balance + 1,
	}
	arg := db.UpdateAccountsParams{ID: account.ID, Balance: updatedAccount.Balance}

	testCases := []struct {
		name          string
		accountId     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenManager token.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
		body          gin.H
	}{
		{
			name:      "Ok",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccounts(gomock.Any(), gomock.Eq(arg)).Times(1).Return(updatedAccount, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, updatedAccount)
			},
			body: gin.H{
				"id":      account.ID,
				"balance": updatedAccount.Balance,
			},
		},
		{
			name:      "InternalServerError",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccounts(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			body: gin.H{
				"id":      account.ID,
				"balance": updatedAccount.Balance,
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccounts(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			body: gin.H{
				"id":      account.ID,
				"balance": updatedAccount.Balance,
			},
		},
		{
			name:      "BadRequest",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccounts(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			body: gin.H{
				"id":      0,
				"balance": updatedAccount.Balance,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newMockServer(t, store)
			SetupRoutes(server)
			recorder := httptest.NewRecorder()

			url := "/account"

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenManager)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}
