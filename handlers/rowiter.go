package handlers

import "database/sql"

// iterateRows owns rows: it invokes fn once per row, always closes rows, and
// returns rows.Err(). fn may return an error to stop iteration early.
func iterateRows(rows *sql.Rows, fn func() error) error {
	defer rows.Close()
	for rows.Next() {
		if err := fn(); err != nil {
			return err
		}
	}
	return rows.Err()
}
