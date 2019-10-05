package models

import (
	"database/sql"
)

type Social struct {
	Facebook string
	Viber    string
	Telegram string
	Youtube  string
}

func (s *Social) Get(db *sql.DB) error {
	rows, err := db.Query("SELECT name, url FROM social")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var url sql.NullString
		err = rows.Scan(&name, &url)
		if err != nil {
			return err
		}

		switch name {
		case "viber":
			if url.Valid {
				s.Viber = url.String
			}
		case "facebook":
			if url.Valid {
				s.Facebook = url.String
			}
		case "youtube":
			if url.Valid {
				s.Youtube = url.String
			}
		case "telegram":
			if url.Valid {
				s.Telegram = url.String
			}
		}
	}
	return nil
}

func (s *Social) Update(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE social SET url = ? WHERE name = \"facebook\"", CheckNullString(s.Facebook))
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE social SET url = ? WHERE name = \"viber\"", CheckNullString(s.Viber))
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE social SET url = ? WHERE name = \"telegram\"", CheckNullString(s.Telegram))
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE social SET url = ? WHERE name = \"youtube\"", CheckNullString(s.Youtube))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
