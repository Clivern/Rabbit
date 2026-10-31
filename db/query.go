// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
)

// querier is the query surface shared by *sql.DB and *sql.Tx.
type querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// QueryOne runs query and scans the single row into a new T.
// A missing row returns nil, nil.
func QueryOne[T any](db querier, query string, scan func(*T) []any, args ...any) (*T, error) {
	item := new(T)
	err := db.QueryRow(query, args...).Scan(scan(item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return item, nil
}

// QueryMany runs query and scans every row into a new T.
func QueryMany[T any](db querier, query string, scan func(*T) []any, args ...any) ([]*T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*T
	for rows.Next() {
		item := new(T)
		if err := rows.Scan(scan(item)...); err != nil {
			return nil, err
		}

		list = append(list, item)
	}

	return list, rows.Err()
}
