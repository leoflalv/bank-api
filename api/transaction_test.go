package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

func randomEntry(accountId int64, amount int64) db.Entry {
	return db.Entry{
		ID:        int64(util.RandomNumber(1, 1000)),
		Amount:    amount,
		AccountID: accountId,
	}
}

func requireBodyMatchTransaction(t *testing.T, body *bytes.Buffer, transaction db.TransactionResult) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var transactionResult db.TransactionResult
	err = json.Unmarshal(data, &transactionResult)
	require.Equal(t, transaction, transactionResult)
}

func TestCreateTransactionAPI(t *testing.T) {
	user1, _ := randomUser()
	fromAccount := randomAccount(user1.Username)
	fromAccount.Currency = util.EUR

	user2, _ := randomUser()
	toAccount := randomAccount(user2.Username)
	toAccount.Currency = util.EUR

	transfer := db.Transfer{
		ID:            int64(util.RandomNumber(1, 1000)),
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        4,
	}

	fromAccountEntry := randomEntry(fromAccount.ID, -transfer.Amount)
	toAccountEntry := randomEntry(toAccount.ID, transfer.Amount)

	arg := db.TransactionParams{FromAccountID: fromAccount.ID, ToAccountID: toAccount.ID, Amount: transfer.Amount}

	modifiedFromAccount := db.Account{
		ID:       fromAccount.ID,
		Owner:    fromAccount.Owner,
		Balance:  fromAccount.Balance - transfer.Amount,
		Currency: fromAccount.Currency,
	}

	modifiedToAccount := db.Account{
		ID:       toAccount.ID,
		Owner:    toAccount.Owner,
		Balance:  toAccount.Balance + transfer.Amount,
		Currency: toAccount.Currency,
	}

	result := db.TransactionResult{Tranfer: transfer, ToAccount: modifiedToAccount, FromAccount: modifiedFromAccount, ToEntry: toAccountEntry, FromEntry: fromAccountEntry}

	testCases := []struct {
		name          string
		toAccountID   int64
		fromAccountID int64
		setupAuth     func(t *testing.T, request *http.Request, tokenManager token.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
		body          gin.H
	}{
		{
			name: "Ok",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)
				store.EXPECT().Transaction(gomock.Any(), gomock.Eq(arg)).Times(1).Return(result, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransaction(t, recorder.Body, result)
			},
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"currency":        fromAccount.Currency,
				"amount":          transfer.Amount,
			},
		},
		{
			name: "BadRequest",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(0)
				store.EXPECT().Transaction(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"currency":        "FAIL",
				"amount":          transfer.Amount,
			},
		},
		{
			name: "NotFoundAccount",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(0)
				store.EXPECT().Transaction(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"currency":        fromAccount.Currency,
				"amount":          transfer.Amount,
			},
		},
		{
			name: "IncorrectCurrency",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(0)
				store.EXPECT().Transaction(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"currency":        util.USD,
				"amount":          transfer.Amount,
			},
		},
		{
			name: "InternalSeverErrorAccount",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(0)
				store.EXPECT().Transaction(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"currency":        fromAccount.Currency,
				"amount":          transfer.Amount,
			},
		},
		{
			name: "InternalSeverErrorTransaction",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager token.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)
				store.EXPECT().Transaction(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.TransactionResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"currency":        fromAccount.Currency,
				"amount":          transfer.Amount,
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

			url := "/transaction"

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
