package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const threshold = 0.001

func TestTransaction(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount()
	account2 := createRandomAccount()

	n := 5
	amount := float64(10)

	errs := make(chan error)
	results := make(chan TransactionResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.Transaction(context.Background(), TransactionParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Tranfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.True(t, diff1-diff2 < threshold && diff1-diff2 > -threshold)
		require.True(t, diff1 > 0)

		k := int(diff1/amount + threshold)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.True(t, account1.Balance-float64(n)*amount-updatedAccount1.Balance < threshold && account1.Balance-float64(n)*amount-updatedAccount1.Balance > -threshold)
	require.True(t, account2.Balance+float64(n)*amount-updatedAccount2.Balance < threshold && account2.Balance+float64(n)*amount-updatedAccount2.Balance > -threshold)
}

func TestTransactionDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount()
	account2 := createRandomAccount()

	n := 10
	amount := float64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.Transaction(context.Background(), TransactionParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.True(t, account1.Balance-updatedAccount1.Balance < threshold && account1.Balance-updatedAccount1.Balance > -threshold)
	require.True(t, account2.Balance-updatedAccount2.Balance < threshold && account2.Balance-updatedAccount2.Balance > -threshold)
}
