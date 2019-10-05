package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"roman_group/models"
	"strconv"
	"time"
)

func isAuth(session *sessions.Session) bool {
	if _, ok := session.Values["admin"]; ok {
		return true
	}
	return false
}

// Admin Router

// Login
func logout(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	session.Values["admin"] = ""
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	http.Redirect(w, r, "/", 302)
	return
}
func login(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if isAuth(session) {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	if r.Method == "POST" {
		login := r.FormValue("login")
		password := r.FormValue("password")
		passwordHash := sha256.Sum256([]byte(password))
		var u models.User
		err = u.Auth(db, login, hex.EncodeToString(passwordHash[:]))
		if err != nil {
			log.Printf("%s\n", err)
			session.AddFlash("Неверный логин или пароль!")
			err = session.Save(r, w)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}
			http.Redirect(w, r, "/admin/login", 302)
			return
		} else {
			session.Values["admin"] = u.Id
			session.Options.MaxAge = 0
			err = session.Save(r, w)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}
			http.Redirect(w, r, "/admin", 302)
			return
		}
	}

	flashes := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/admin/login.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"csrfField": csrf.TemplateField(r),
		"flashes":   flashes,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

// News
func adminNews(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	var page int64 = 1
	pageParam, ok := mux.Vars(r)["page"]
	if ok {
		page, err = strconv.ParseInt(pageParam, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
	}

	pagination := initNewsPaginator(page, 10)

	var nl models.NewsList
	err = nl.GetAllNews(db, page, 10)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin/adminNews.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"nl":         nl,
		"pagination": pagination,
		"csrfField":  csrf.TemplateField(r),
	})

	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func newsDelete(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	idValue := r.FormValue("id")
	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var n models.News
	err = n.Get(db, id)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	err = n.Delete(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	err = DeleteImages(n.Images)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		_ = n.Add(db)
		return
	}
	http.Redirect(w, r, "/admin", 302)
	return
}
func newsEdit(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	idParams := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idParams, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var news models.News
	err = news.Get(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("News Edit: %s\n", err)
			http.Redirect(w, r, "/404", 302)
			return
		}
		w.WriteHeader(500)
		log.Printf("%s\n", err)
		return
	}

	if r.Method == "POST" {

		title := r.FormValue("title")
		description := r.FormValue("description")
		text := r.FormValue("text")
		images := news.Images

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		file, header, err := r.FormFile("images")
		if err == nil {
			defer file.Close()
			if header.Header.Get("Content-Type") != "image/jpeg" && header.Header.Get("Content-Type") != "image/png" {
				log.Printf("AddProject: Неверный тип формата файла: %s\n", header.Header.Get("Content-Type"))
				http.Redirect(w, r, "/500", 302)
				return
			}
			images, err = UploadImages(file)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}

			err = DeleteImages(news.Images)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}
		} else if err != http.ErrMissingFile {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		news.Images = images
		news.Title = title
		news.Description = description
		news.Text = text
		err = news.Update(db)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
		http.Redirect(w, r, "/admin", 302)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin/adminEditNews.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"csrfField": csrf.TemplateField(r),
		"news":      news,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func newsAdd(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	if r.Method == "POST" {

		title := r.FormValue("title")
		description := r.FormValue("description")
		text := r.FormValue("text")

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		file, header, err := r.FormFile("images")
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
		defer file.Close()

		if header.Header.Get("Content-Type") != "image/jpeg" && header.Header.Get("Content-Type") != "image/png" {
			log.Printf("AddProject: Неверный тип формата файла: %s\n", header.Header.Get("Content-Type"))
			http.Redirect(w, r, "/500", 302)
			return
		}

		images, err := UploadImages(file)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		var n models.News
		n.Title = title
		n.Description = description
		n.Text = text
		n.Images = images
		n.CountSee = 0
		n.Date = time.Now().Format("2006-01-02 15:04:05")
		err = n.Add(db)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			err = DeleteImages(images)
			if err != nil {
				log.Printf("%s\n", err)
			}
			return
		}

		http.Redirect(w, r, "/admin", 302)
		return
	}

	var news models.News
	tmpl := template.Must(template.ParseFiles("templates/admin/adminAddNews.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"csrfField": csrf.TemplateField(r),
		"news":      news,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

// Projects
func adminProjects(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	var page int64 = 1
	pageParam, ok := mux.Vars(r)["page"]
	if ok {
		page, err = strconv.ParseInt(pageParam, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
	}

	pagination := initProjectPaginator(page, 10)

	var pl models.ProjectList
	err = pl.GetProjectList(db, page, 10)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin/adminProjects.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"csrfField":  csrf.TemplateField(r),
		"pl":         pl,
		"pagination": pagination,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func projectsAdd(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		description := r.FormValue("description")
		favoriteParam := r.FormValue("favorite")
		favorite, err := strconv.ParseInt(favoriteParam, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		err = r.ParseMultipartForm(22 << 20)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		file, header, err := r.FormFile("images")
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
		defer file.Close()
		if (header.Header.Get("Content-Type") != "image/jpeg" && header.Header.Get("Content-Type") != "image/png") || (header.Size > 2<<20) {
			log.Printf("AddProject: Неверный тип формата файла: %s\n", header.Header.Get("Content-Type"))
			http.Redirect(w, r, "/413", 302)
			return
		}

		images, err := UploadImages(file)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		var project models.Project
		project.Name = name
		project.Description = description
		project.Images = images
		project.IsFavorite = favorite
		project.Date = time.Now().Format("2006-01-02 15:04:05")

		photos := r.MultipartForm.File["photos"]

		for i := range photos {
			if (photos[i].Header.Get("Content-Type") != "image/jpeg" && photos[i].Header.Get("Content-Type") != "image/png") || (photos[i].Size > 2<<20) {
				continue
			}

			file, err := photos[i].Open()
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}

			src, err := UploadImages(file)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}
			var photo models.Photo
			photo.Src = src
			photo.Date = time.Now().Format("2006-01-02 15:04:05")
			project.Photos = append(project.Photos, photo)
			file.Close()
		}

		video1Param := r.FormValue("video1")
		video2Param := r.FormValue("video2")
		video3Param := r.FormValue("video3")

		regexPattern := "https://www.youtube.com/watch\\?v="
		videoSrcPattern := "https://www.youtube.com/embed/"
		re := regexp.MustCompile(regexPattern)

		if re.Match([]byte(video1Param)) {
			strings := re.Split(video1Param, -1)
			if len(strings) > 1 {
				project.Video1.String = videoSrcPattern + strings[1]
			}
		}

		if re.Match([]byte(video2Param)) {
			strings := re.Split(video2Param, -1)
			if len(strings) > 1 {
				project.Video2.String = videoSrcPattern + strings[1]

			}
		}

		if re.Match([]byte(video3Param)) {
			strings := re.Split(video3Param, -1)
			if len(strings) > 1 {
				project.Video3.String = videoSrcPattern + strings[1]
			}
		}

		err = project.Add(db)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
		http.Redirect(w, r, "/admin/projects", 302)
		return
	}

	var project models.Project
	tmpl := template.Must(template.ParseFiles("templates/admin/adminAddProject.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"csrfField": csrf.TemplateField(r),
		"Project":   project,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func projectsEdit(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	idParams := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idParams, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var project models.Project
	err = project.Get(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("%s\n", err)
			http.Redirect(w, r, "/404", 302)
			return
		}
		log.Printf("%s\n", err)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		description := r.FormValue("description")
		images := project.Images
		favoriteParam := r.FormValue("favorite")
		favorite, err := strconv.ParseInt(favoriteParam, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		err = r.ParseMultipartForm(2 << 20)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		file, header, err := r.FormFile("images")
		if err == nil {
			defer file.Close()
			if (header.Header.Get("Content-Type") != "image/jpeg" && header.Header.Get("Content-Type") != "image/png") || (header.Size > 2<<20) {
				log.Printf("AddProject: Неверный тип формата файла: %s\n", header.Header.Get("Content-Type"))
				http.Redirect(w, r, "/500", 302)
				return
			}
			images, err = UploadImages(file)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}

			err = DeleteImages(project.Images)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}
		} else if err != http.ErrMissingFile {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		project.Name = name
		project.Description = description
		project.Images = images
		project.IsFavorite = favorite

		video1Param := r.FormValue("video1")
		video2Param := r.FormValue("video2")
		video3Param := r.FormValue("video3")

		regexPattern := "https://www.youtube.com/watch\\?v="
		videoSrcPattern := "https://www.youtube.com/embed/"
		re := regexp.MustCompile(regexPattern)

		if re.Match([]byte(video1Param)) || video1Param == "" {
			strings := re.Split(video1Param, -1)
			if len(strings) > 1 {
				project.Video1.String = videoSrcPattern + strings[1]
			}
		} else if video1Param == "" {
			project.Video2.String = ""
		}

		if re.Match([]byte(video2Param)) || video1Param == "" {
			strings := re.Split(video2Param, -1)
			if len(strings) > 1 {
				project.Video2.String = videoSrcPattern + strings[1]
			}
		} else if video2Param == "" {
			project.Video2.String = ""
		}

		if re.Match([]byte(video3Param)) {
			strings := re.Split(video3Param, -1)
			if len(strings) > 1 {
				project.Video3.String = videoSrcPattern + strings[1]
			}
		} else if video3Param == "" {
			project.Video3.String = ""
		}

		err = project.Update(db)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		http.Redirect(w, r, "/admin/projects", 302)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin/adminEditProject.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"csrfField": csrf.TemplateField(r),
		"project":   project,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func projectsDelete(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	idParam := r.FormValue("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var project models.Project
	err = project.Get(db, id)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	err = project.Delete(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	err = DeleteImages(project.Images)

	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	for i := range project.Photos {
		err = DeleteImages(project.Photos[i].Src)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
	}

	http.Redirect(w, r, "/admin/projects", 302)
	return
}
func projectPhotoAdd(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	err = r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	idProjectParam := r.FormValue("id")
	idProject, err := strconv.ParseInt(idProjectParam, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	photos := r.MultipartForm.File["photos"]

	var newPhotos []models.Photo

	for i := range photos {
		if (photos[i].Header.Get("Content-Type") != "image/jpeg" && photos[i].Header.Get("Content-Type") != "image/png") || (photos[i].Size > 2<<20) {
			continue
		}

		file, err := photos[i].Open()
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
		src, err := UploadImages(file)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
		var photo models.Photo
		photo.Src = src
		photo.Date = time.Now().Format("2006-01-02 15:04:05")
		newPhotos = append(newPhotos, photo)
		file.Close()
	}

	err = models.AddPhotoToProject(db, idProject, newPhotos)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	http.Redirect(w, r, "/admin/projects/"+idProjectParam+"/edit", 302)
	return
}
func projectPhotoDelete(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	idProjectParam := mux.Vars(r)["idProject"]

	idPhotoParam := mux.Vars(r)["idPhoto"]
	idPhoto, err := strconv.ParseInt(idPhotoParam, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var p models.Photo
	err = p.Get(db, idPhoto)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	err = p.Delete(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	err = DeleteImages(p.Src)
	if err != nil {
		log.Printf("%s\n", err)
		err = p.Add(db)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
	}
	http.Redirect(w, r, "/admin/projects/"+idProjectParam+"/edit", 302)
	return
}

// Admin Social
func adminSocial(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	var s models.Social
	err = s.Get(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if r.Method == "POST" {
		s.Facebook = r.FormValue("facebook")
		s.Telegram = r.FormValue("telegram")
		s.Youtube = r.FormValue("youtube")
		s.Viber = r.FormValue("viber")

		err = s.Update(db)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		session.AddFlash("Изминения успешно сохранены!")
		err = session.Save(r, w)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		http.Redirect(w, r, "/admin/social", 302)
		return
	}

	flashes := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin/adminSocial.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"flashes":   flashes,
		"social":    s,
		"csrfField": csrf.TemplateField(r),
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

// Admin Secure
func adminSecure(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	if r.Method == "POST" {
		var userId int64
		userId, ok := session.Values["admin"].(int64)
		if !ok {
			http.Redirect(w, r, "/500", 302)
			log.Printf("Ошибка при попытке досать id из сессии\n")
			return
		}

		oldPassword := r.FormValue("oldpass")
		newPass := r.FormValue("newpass")
		newPass2 := r.FormValue("newpass2")

		oldPassHash := sha256.Sum256([]byte(oldPassword))
		check, err := models.CheckUserPassword(db, userId, hex.EncodeToString(oldPassHash[:]))
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		if !check {
			session.AddFlash("Неверный пароль!")
			err = session.Save(r, w)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}
			http.Redirect(w, r, "/admin/secure", 302)
			return
		}

		if newPass != newPass2 {
			session.AddFlash("Пароли не совпадают!")
			err = session.Save(r, w)
			if err != nil {
				http.Redirect(w, r, "/500", 302)
				log.Printf("%s\n", err)
				return
			}
			http.Redirect(w, r, "/admin/secure", 302)
			return
		}

		newPassHash := sha256.Sum256([]byte(newPass))
		err = models.UpdateUserPassword(db, userId, hex.EncodeToString(newPassHash[:]))
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}

		session.AddFlash("Пароль успешно обновлен!")
		err = session.Save(r, w)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
		http.Redirect(w, r, "/admin/secure", 302)
		return
	}

	flashes := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin/adminSecure.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"csrfField": csrf.TemplateField(r),
		"flashes":   flashes,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

// Admin Mail
func adminMail(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	var checkStatus bool
	categoryParam, ok := mux.Vars(r)["category"]
	if ok {
		if categoryParam == "new" {
			checkStatus = false
		} else if categoryParam == "old" {
			checkStatus = true
		} else {
			http.Redirect(w, r, "/404", 302)
			return
		}
	}

	var page int64 = 1
	pageParam, ok := mux.Vars(r)["page"]
	if ok {
		page, err = strconv.ParseInt(pageParam, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n", err)
			return
		}
	}

	pagination := iniFeedBackPaginator(page, 20, checkStatus)

	var fbl models.FeedBackList
	if checkStatus {
		err = fbl.GetOldFedBack(db, page, 20)
	} else {
		err = fbl.GetNewFedBack(db, page, 20)
	}
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin/adminMail.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"pagination":  pagination,
		"fbl":         fbl,
		"checkStatus": checkStatus,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func adminCheckMail(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman-admin")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	if !isAuth(session) {
		http.Redirect(w, r, "/admin/login", 302)
		return
	}

	idParams := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idParams, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var fb models.FeedBack
	err = fb.Get(db, id)
	if err != nil && err != sql.ErrNoRows {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	} else if err == sql.ErrNoRows {
		http.Redirect(w, r, "/404", 302)
		log.Printf("%s\n", err)
	}

	err = fb.Check(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	http.Redirect(w, r, "/admin/mail/old", 302)
	return
}

// Admin Methods
