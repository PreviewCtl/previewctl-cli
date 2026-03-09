package database

import (
	"errors"

	sqlite3 "modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"
)

func isSQLUniqueConstraintError(original error) bool {
	var sqliteErr *sqlite3.Error
	if errors.As(original, &sqliteErr) {
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE ||
			sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_PRIMARYKEY
	}
	return false
}
