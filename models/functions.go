package models

import (
	"database/sql"
)

func CheckNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
