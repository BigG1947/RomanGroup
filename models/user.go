package models

import (
	"database/sql"
	"github.com/pkg/errors"
)

type User struct {
	Id       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *User) Auth(db *sql.DB, login string, password string) error {
	err := db.QueryRow("SELECT id, login, password FROM admin WHERE login = ? LIMIT 1", login).Scan(&u.Id, &u.Login, &u.Password)
	if err != nil {
		return err
	}
	if password != u.Password {
		return errors.New("Неверный пароль!")
	}
	return nil
}

func CheckUserPassword(db *sql.DB, idUser int64, passwordHash string) (bool, error) {
	var password string
	err := db.QueryRow("SELECT password FROM admin WHERE id = ?", idUser).Scan(&password)
	if err != nil {
		return false, err
	}

	if password == passwordHash {
		return true, nil
	}

	return false, nil
}

func UpdateUserPassword(db *sql.DB, idUser int64, newPassword string) error {
	_, err := db.Exec("UPDATE admin SET password = ? WHERE id = ?", newPassword, idUser)
	if err != nil {
		return err
	}
	return nil
}
