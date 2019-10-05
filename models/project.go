package models

import (
	"database/sql"
)

type Project struct {
	Id          int64          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Images      string         `json:"images"`
	Photos      []Photo        `json:"photos"`
	Video1      sql.NullString `json:"video_1"`
	Video2      sql.NullString `json:"video_1"`
	Video3      sql.NullString `json:"video_1"`
	IsFavorite  int64          `json:"is_favorite"`
	Date        string         `json:"date"`
}

type Photo struct {
	Id        int64  `json:"id"`
	Src       string `json:"src"`
	Date      string `json:"date"`
	IdProject int64  `json:"id_project"`
}

type ProjectList struct {
	ProjectList []Project `json:"project_list"`
}

// Project

func (pl *ProjectList) GetProjectList(db *sql.DB, page int64, perPage int64) error {
	offsetStart := (page - 1) * perPage
	rows, err := db.Query("SELECT id, name, description, images, is_favorite, date FROM project ORDER BY date DESC, id DESC LIMIT ?,?", offsetStart, perPage)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p Project
		err = rows.Scan(&p.Id, &p.Name, &p.Description, &p.Images, &p.IsFavorite, &p.Date)
		if err != nil {
			return err
		}
		pl.ProjectList = append(pl.ProjectList, p)
	}

	return nil
}

func (pl *ProjectList) GetFavoriteProjectList(db *sql.DB) error {
	rows, err := db.Query("SELECT id, name, description, images, is_favorite, date FROM project WHERE is_favorite != 0 ORDER BY is_favorite ASC LIMIT 3")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p Project
		err = rows.Scan(&p.Id, &p.Name, &p.Description, &p.Images, &p.IsFavorite, &p.Date)
		if err != nil {
			return err
		}
		pl.ProjectList = append(pl.ProjectList, p)
	}

	return nil
}

func (p *Project) Get(db *sql.DB, id int64) error {
	err := db.QueryRow("SELECT id, name, description, images, is_favorite, video1, video2, video3, date FROM project WHERE id = ?", id).Scan(&p.Id, &p.Name, &p.Description, &p.Images, &p.IsFavorite, &p.Video1, &p.Video2, &p.Video3, &p.Date)
	if err != nil {
		return err
	}

	var photos []Photo
	rows, err := db.Query("SELECT id, src, date, id_project FROM project_photo WHERE id_project = ?", p.Id)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p Photo
		err = rows.Scan(&p.Id, &p.Src, &p.Date, &p.IdProject)
		if err != nil {
			return err
		}
		photos = append(photos, p)
	}
	p.Photos = photos

	return nil
}

func (p *Project) Add(db *sql.DB) error {
	res, err := db.Exec("INSERT INTO project (name, description, images, is_favorite, video1, video2, video3, date) VALUES (?,?,?,?,?,?,?,?)", p.Name, p.Description, p.Images, p.IsFavorite, p.Video1.String, p.Video2.String, p.Video3.String, p.Date)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if p.IsFavorite > 0 {
		err = changeFavoriteProject(db, id, p.IsFavorite)
		if err != nil {
			return err
		}
	}

	if len(p.Photos) > 0 {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, photo := range p.Photos {
			_, err = tx.Exec("INSERT INTO project_photo(src, date, id_project) VALUES (?,?,?)", photo.Src, photo.Date, id)
			if err != nil {
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) Update(db *sql.DB) error {
	_, err := db.Exec("UPDATE project SET name = ?, description = ?, images = ?, is_favorite = ?, video1 = ?, video2 = ?, video3 = ? WHERE id = ?", p.Name, p.Description, p.Images, p.IsFavorite, p.Video1.String, p.Video2.String, p.Video3.String, p.Id)
	if err != nil {
		return err
	}
	if p.IsFavorite > 0 {
		err = changeFavoriteProject(db, p.Id, p.IsFavorite)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) Delete(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM project_photo WHERE id_project = ?", p.Id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM project WHERE id = ?", p.Id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil

}

// Photo

func (p *Photo) Get(db *sql.DB, id int64) error {
	err := db.QueryRow("SELECT id, src, date, id_project FROM project_photo WHERE id = ?", id).Scan(&p.Id, &p.Src, &p.Date, &p.IdProject)
	if err != nil {
		return err
	}
	return nil
}

func (p *Photo) Add(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO project_photo(src, date, id_project) VALUES (?, ?, ?)", p.Src, p.Date, p.IdProject)
	if err != nil {
		return err
	}
	return nil
}

func (p *Photo) Delete(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM project_photo WHERE id = ?", p.Id)
	if err != nil {
		return err
	}
	return nil
}

// Functions

func AddPhotoToProject(db *sql.DB, idProject int64, photos []Photo) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, photo := range photos {
		_, err = tx.Exec("INSERT INTO project_photo(src, date, id_project) VALUES (?,?,?)", photo.Src, photo.Date, idProject)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func GetCountProject(db *sql.DB) int64 {
	var count int64
	err := db.QueryRow("SELECT COUNT(id) FROM project").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func changeFavoriteProject(db *sql.DB, idProject int64, favoritePosition int64) error {
	_, err := db.Exec("UPDATE project SET is_favorite = 0 WHERE id != ? AND is_favorite = ?", idProject, favoritePosition)
	if err != nil {
		return err
	}
	return nil
}
