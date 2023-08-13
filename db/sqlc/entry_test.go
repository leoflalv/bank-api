package db

import (
	"context"
	"github/leoflalv/bank-api/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomEntry(account Account) Entry {
	args := CreateEntryParams{
		Amount:    util.RandomNumber(1, 1000),
		AccountID: account.ID,
	}

	entry, _ := testQueries.CreateEntry(context.Background(), args)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount()

	args := CreateEntryParams{
		Amount:    util.RandomNumber(1, 1000),
		AccountID: account.ID,
	}

	entry, err := testQueries.CreateEntry(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, args.Amount, entry.Amount)
	require.Equal(t, args.AccountID, entry.AccountID)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount()
	entry1 := createRandomEntry(account)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount()
	for i := 0; i < 10; i++ {
		createRandomEntry(account)
	}

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.Equal(t, arg.AccountID, entry.AccountID)
	}
}
