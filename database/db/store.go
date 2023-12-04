package db

import "github.com/jackc/pgx/v5/pgxpool"

type Store interface {
	Querier
}

// SqlStore provides all functions to execute db queries and transactions
type SqlStore struct {
	*Queries
	conn *pgxpool.Pool
}

// NewStore creates a new Store
func NewStore(conn *pgxpool.Pool) Store {
	return &SqlStore{
		conn:    conn,
		Queries: New(conn),
	}
}
