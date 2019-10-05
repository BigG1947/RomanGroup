package models

import (
	"database/sql"
)

type FeedBack struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Theme   string `json:"theme"`
	Text    string `json:"text"`
	Date    string `json:"date"`
	IsCheck bool   `json:"is_check"`
}

type FeedBackList struct {
	FeedBackList []FeedBack
}

func (fb *FeedBack) Add(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO feedback(name, email, theme,  text, date) VALUES (?, ?, ?, ?, ?);", fb.Name, fb.Email, fb.Theme, fb.Text, fb.Date)
	if err != nil {
		return err
	}
	return nil
}

func (fb *FeedBack) Get(db *sql.DB, id int64) error {
	err := db.QueryRow("SELECT id, name, email, theme, text, date, is_check FROM feedback WHERE id = ?;", id).Scan(&fb.Id, &fb.Name, &fb.Email, &fb.Theme, &fb.Text, &fb.Date, &fb.IsCheck)
	if err != nil {
		return err
	}
	return nil
}

func (fb *FeedBack) Check(db *sql.DB) error {
	_, err := db.Exec("UPDATE feedback SET is_check = true WHERE id = ?", fb.Id)
	if err != nil {
		return err
	}
	return nil
}

func (fbl *FeedBackList) GetOldFedBack(db *sql.DB, page int64, perPage int64) error {
	offsetStart := (page - 1) * perPage
	rows, err := db.Query("SELECT id, name, email, theme, text, date, is_check FROM feedback WHERE is_check = true ORDER BY date DESC, id DESC LIMIT ?, ?", offsetStart, perPage)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var fb FeedBack
		err = rows.Scan(&fb.Id, &fb.Name, &fb.Email, &fb.Theme, &fb.Text, &fb.Date, &fb.IsCheck)
		if err != nil {
			return err
		}
		fbl.FeedBackList = append(fbl.FeedBackList, fb)
	}
	return nil
}

func (fbl *FeedBackList) GetNewFedBack(db *sql.DB, page int64, perPage int64) error {
	offsetStart := (page - 1) * perPage
	rows, err := db.Query("SELECT id, name, email, theme, text, date, is_check FROM feedback WHERE is_check = false ORDER BY date DESC, id DESC LIMIT ?, ?", offsetStart, perPage)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var fb FeedBack
		err = rows.Scan(&fb.Id, &fb.Name, &fb.Email, &fb.Theme, &fb.Text, &fb.Date, &fb.IsCheck)
		if err != nil {
			return err
		}
		fbl.FeedBackList = append(fbl.FeedBackList, fb)
	}
	return nil
}

func GetCountFeedBacks(db *sql.DB, checkStatus bool) int64 {
	var count int64
	err := db.QueryRow("SELECT COUNT(id) FROM feedback WHERE is_check = ?", checkStatus).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
