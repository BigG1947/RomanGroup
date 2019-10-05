package models

import (
	"database/sql"
)

type NewsList struct {
	NewsList []News `json:"news_list"`
}

type News struct {
	Id          int64  `json:"id"`
	Title       string `json:"name"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Images      string `json:"images"`
	Date        string `json:"date"`
	CountSee    int    `json:"count_see"`
}

func (n *News) Get(db *sql.DB, id int64) error {
	row := db.QueryRow("SELECT id, title, description, text, images, date, count_see FROM news WHERE id = ?", id)
	err := row.Scan(&n.Id, &n.Title, &n.Description, &n.Text, &n.Images, &n.Date, &n.CountSee)
	if err != nil {
		return err
	}
	return nil
}

func (n *News) Add(db *sql.DB) error {
	res, err := db.Exec("INSERT INTO news(title, description, text, images, date, count_see) VALUES (?,?,?,?,?,?)",
		n.Title, n.Description, n.Text, n.Images, n.Date, n.CountSee)
	if err != nil {
		return err
	}
	n.Id, err = res.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func (n *News) Update(db *sql.DB) error {
	_, err := db.Exec("UPDATE news SET title = ?, description = ?, text = ?, images = ?, date = ?, count_see = ? WHERE id = ?",
		n.Title, n.Description, n.Text, n.Images, n.Date, n.CountSee, n.Id)
	if err != nil {
		return err
	}
	return nil
}

func (n *News) IncrementCounter(db *sql.DB) error {
	_, err := db.Exec("UPDATE news SET count_see = ? WHERE id = ?;", n.CountSee, n.Id)
	if err != nil {
		return err
	}
	return nil
}

func (n *News) Delete(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM news WHERE id = ?", n.Id)
	if err != nil {
		return err
	}
	return nil
}

func (nl *NewsList) GetAllNews(db *sql.DB, page int64, perPage int64) error {
	offsetStart := (page - 1) * perPage
	rows, err := db.Query("SELECT id, title, description, text, images, date, count_see FROM news ORDER BY date DESC, id DESC LIMIT ?,?", offsetStart, perPage)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var n News
		err = rows.Scan(&n.Id, &n.Title, &n.Description, &n.Text, &n.Images, &n.Date, &n.CountSee)
		if err != nil {
			return err
		}
		nl.NewsList = append(nl.NewsList, n)
	}

	return nil
}

func (nl *NewsList) GetLatestNews(db *sql.DB) error {
	rows, err := db.Query("SELECT id, title, description, text, images, date, count_see FROM news ORDER BY date DESC, id DESC LIMIT 3")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var n News
		err = rows.Scan(&n.Id, &n.Title, &n.Description, &n.Text, &n.Images, &n.Date, &n.CountSee)
		if err != nil {
			return err
		}
		nl.NewsList = append(nl.NewsList, n)
	}

	return nil
}

func GetCountNews(db *sql.DB) int64 {
	var count int64
	err := db.QueryRow("SELECT COUNT(id) FROM news").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
