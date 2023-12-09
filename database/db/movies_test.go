package db

import (
	"context"
	"main/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func createRandomMovie(t *testing.T) *Movie {
	arg := &CreateMovieParams{
		ID:          util.RandomString(8),
		Title:       util.RandomString(8),
		Description: util.RandomString(16),
		Director:    util.RandomString(8),
	}

	movie, err := testStore.CreateMovie(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, movie)

	require.Equal(t, arg.ID, movie.ID)
	require.Equal(t, arg.Title, movie.Title)
	require.Equal(t, arg.Description, movie.Description)
	require.Equal(t, arg.Director, movie.Director)

	return movie
}

func TestCreateMovie(t *testing.T) {
	createRandomMovie(t)
}

func TestGetMovie(t *testing.T) {
	movie1 := createRandomMovie(t)
	movie2, err := testStore.GetMovie(context.Background(), movie1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, movie2)

	require.Equal(t, movie1.ID, movie2.ID)
	require.Equal(t, movie1.Title, movie2.Title)
	require.Equal(t, movie1.Description, movie2.Description)
	require.Equal(t, movie1.Director, movie2.Director)
}
