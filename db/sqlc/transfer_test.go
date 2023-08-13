package db

import (
	"context"
	"github/leoflalv/bank-api/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomTransfer(fromAccount Account, toAccount Account) Transfer {
	args := CreateTransferParams{
		Amount:        util.RandomNumber(1, 1000),
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
	}

	entry, _ := testQueries.CreateTransfer(context.Background(), args)

	return entry
}

func TestCreateTransaction(t *testing.T) {
	fromAccount := createRandomAccount()
	toAccount := createRandomAccount()

	args := CreateTransferParams{
		Amount:        util.RandomNumber(1, 1000),
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, args.Amount, transfer.Amount)
	require.Equal(t, args.FromAccountID, transfer.FromAccountID)
	require.Equal(t, args.ToAccountID, transfer.ToAccountID)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
}

func TestGetTransaction(t *testing.T) {
	fromAccount := createRandomAccount()
	toAccount := createRandomAccount()

	transfer1 := createRandomTransfer(fromAccount, toAccount)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}

func TestListTransactions(t *testing.T) {
	fromAccount := createRandomAccount()
	toAccount := createRandomAccount()

	for i := 0; i < 10; i++ {
		createRandomTransfer(fromAccount, toAccount)
	}

	arg1 := ListTransfersParams{
		FromAccountID: fromAccount.ID,
		Limit:         5,
		Offset:        5,
	}

	transfers1, err := testQueries.ListTransfers(context.Background(), arg1)
	require.NoError(t, err)
	require.Len(t, transfers1, 5)

	for _, transfer := range transfers1 {
		require.NotEmpty(t, transfer)
		require.Equal(t, arg1.FromAccountID, transfer.FromAccountID)
	}

	arg2 := ListTransfersParams{
		ToAccountID: toAccount.ID,
		Limit:       5,
		Offset:      5,
	}

	transfers2, err := testQueries.ListTransfers(context.Background(), arg2)
	require.NoError(t, err)
	require.Len(t, transfers2, 5)

	for _, transfer := range transfers1 {
		require.NotEmpty(t, transfer)
		require.Equal(t, arg2.ToAccountID, transfer.ToAccountID)
	}

}
