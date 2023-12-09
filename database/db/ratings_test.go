package db

import (
	"context"
	"main/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func createRandomRating(t *testing.T, movieId, recordType string) *Rating {
	arg := &CreateRatingParams{
		MovieID:    movieId,
		RecordType: recordType,
		UserID:     util.RandomString(8),
		Value:      int32(util.RandomInt(0, 10)),
	}

	rating, err := testStore.CreateRating(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, rating)

	require.Equal(t, arg.MovieID, rating.MovieID)
	require.Equal(t, arg.RecordType, rating.RecordType)
	require.Equal(t, arg.UserID, rating.UserID)
	require.Equal(t, arg.Value, rating.Value)

	return rating
}

func TestCreateRating(t *testing.T) {
	movieId := util.RandomString(8)
	recordType := util.RandomString(8)
	createRandomRating(t, movieId, recordType)
}

func TestListRatings(t *testing.T) {
	movieId := util.RandomString(8)
	recordType := util.RandomString(8)

	var lastAccount *Rating
	for i := 0; i < 10; i++ {
		lastAccount = createRandomRating(t, movieId, recordType)
	}

	arg := &ListRatingsParams{
		MovieID:    movieId,
		RecordType: recordType,
	}

	ratings, err := testStore.ListRatings(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, ratings)

	for _, rating := range ratings {
		require.NotEmpty(t, rating)
		require.Equal(t, lastAccount.MovieID, rating.MovieID)
		require.Equal(t, lastAccount.RecordType, rating.RecordType)
	}
}
