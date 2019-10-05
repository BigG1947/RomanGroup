package main

import (
	"database/sql"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	htmlTemplate "html/template"
	"log"
	"net/http"
	"roman_group/models"
	"strconv"
	textTemplate "text/template"
	"time"
)

func routerInit() *mux.Router {
	router := mux.NewRouter()
	handler404 := router.HandleFunc("/404", error404).GetHandler()
	router.NotFoundHandler = handler404

	// Dir serve
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.PathPrefix("/upload-images/").Handler(http.StripPrefix("/upload-images/", http.FileServer(http.Dir("upload-images"))))

	// Main Routes
	router.HandleFunc("/", index)
	router.HandleFunc("/rss", rss)
	router.HandleFunc("/about", aboutUs)
	router.HandleFunc("/500", error500)
	router.HandleFunc("/500", error413)

	// News
	router.HandleFunc("/news", news)
	router.HandleFunc("/news/page={page:[0-9]+}", news)
	router.HandleFunc("/news/{id:[0-9]+}", singleNews)

	// Projects
	router.HandleFunc("/projects", projects)
	router.HandleFunc("/projects/page={page:[0-9]+}", projects)
	router.HandleFunc("/projects/{id:[0-9]+}", singleProjects)

	// FeedBacks
	router.HandleFunc("/feedbacks/add", addFeedBack).Methods("POST")

	// Admin News
	router.HandleFunc("/admin", adminNews)
	router.HandleFunc("/admin/news", adminNews)
	router.HandleFunc("/admin/news/page={page:[0-9]+}", adminNews)
	router.HandleFunc("/admin/news/add", newsAdd).Methods("GET", "POST")
	router.HandleFunc("/admin/news/{id:[0-9]+}/edit", newsEdit).Methods("GET", "POST")
	router.HandleFunc("/admin/news/delete", newsDelete).Methods("POST")

	// Admin Projects
	router.HandleFunc("/admin/projects", adminProjects)
	router.HandleFunc("/admin/projects/page={page:[0-9]+}", adminProjects)
	router.HandleFunc("/admin/projects/add", projectsAdd).Methods("GET", "POST")
	router.HandleFunc("/admin/projects/{id:[0-9]+}/edit", projectsEdit).Methods("GET", "POST")
	router.HandleFunc("/admin/projects/delete", projectsDelete).Methods("POST")
	router.HandleFunc("/admin/projects/{idProject:[0-9]+}/photo/{idPhoto:[0-9]+}/delete", projectPhotoDelete)
	router.HandleFunc("/admin/project/photo/add", projectPhotoAdd).Methods("POST")

	// AdminSocial
	router.HandleFunc("/admin/social", adminSocial).Methods("GET", "POST")

	// Admin Secure
	router.HandleFunc("/admin/secure", adminSecure).Methods("GET", "POST")

	// Admin Main
	router.HandleFunc("/admin/mail", adminMail)
	router.HandleFunc("/admin/mail/{category}", adminMail)
	router.HandleFunc("/admin/mail/{category}/page={page:[0-9]+}", adminMail)
	router.HandleFunc("/admin/mail/feedback/{id:[0-9]+}/check", adminCheckMail)

	// Admin Login
	router.HandleFunc("/admin/login", login).Methods("GET", "POST")
	router.HandleFunc("/admin/logout", logout)

	return router
}

func error404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/error404.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func error413(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(413)
	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/error413.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
func error500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/error500.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

func addFeedBack(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman_group_fb")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n")
		return
	}

	access, ok := session.Values["feedback-access"]
	if ok && !access.(bool) {
		session.AddFlash("Вы недавно уже отправляли вопрос. Проверьте пожалуйста вашу электронную почту!")
		err = session.Save(r, w)
		if err != nil {
			http.Redirect(w, r, "/500", 302)
			log.Printf("%s\n")
			return
		}
		http.Redirect(w, r, "/#contacts", 302)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	theme := r.FormValue("theme")
	text := r.FormValue("text")

	var fb models.FeedBack
	fb.Name = name
	fb.Email = email
	fb.Theme = theme
	fb.Text = text
	fb.Date = time.Now().Format("2006-01-02 15:04:05")

	err = fb.Add(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	session.Values["feedback-access"] = false
	session.Options.MaxAge = 60 * 60 * 2
	session.AddFlash("Ваше сообщения успешно отправленно! Ожидайте ответ на указанную вами электронную почту!")
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	http.Redirect(w, r, "/#contacts", 302)
	return
}

func singleProjects(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var project models.Project
	err = project.Get(db, id)
	if err != nil {
		http.Redirect(w, r, "/404", 302)
		log.Printf("%s\n", err)
		return
	}

	var s models.Social
	err = s.Get(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/gallery.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"project": project,
		"social":  s,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

func projects(w http.ResponseWriter, r *http.Request) {
	var page int64 = 1
	var err error
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

	var s models.Social
	err = s.Get(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/projects.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"pl":         pl,
		"pagination": pagination,
		"social":     s,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

func singleNews(w http.ResponseWriter, r *http.Request) {
	var news models.News
	idParam := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	err = news.Get(db, id)
	if err != nil && err != sql.ErrNoRows {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	} else if err == sql.ErrNoRows {
		http.Redirect(w, r, "/404", 302)
		return
	}

	var s models.Social
	err = s.Get(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	news.CountSee++
	err = news.IncrementCounter(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	text := htmlTemplate.HTML(news.Text)
	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/fullnews.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"news":   news,
		"text":   text,
		"social": s,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

func news(w http.ResponseWriter, r *http.Request) {
	var page int64 = 1
	var err error
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

	var s models.Social
	err = s.Get(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/news.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"nl":         nl,
		"pagination": pagination,
		"social":     s,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

func index(w http.ResponseWriter, r *http.Request) {
	session, err := session.Get(r, "roman_group_fb")
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var nl models.NewsList
	err = nl.GetLatestNews(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var pl models.ProjectList
	err = pl.GetFavoriteProjectList(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	var s models.Social
	err = s.Get(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	flashes := session.Flashes()
	if session.IsNew {
		session.Options.MaxAge = -1
	}
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := htmlTemplate.Must(htmlTemplate.New("index.html").Funcs(htmlTemplate.FuncMap{
		"safeSocial": safeSocial,
	}).ParseFiles("templates/index.html"))

	err = tmpl.Execute(w, map[string]interface{}{
		"nl":        nl,
		"pl":        pl,
		"social":    s,
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

func aboutUs(w http.ResponseWriter, r *http.Request) {

	var s models.Social
	err := s.Get(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}

	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles("templates/about us.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"social": s,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}

func rss(w http.ResponseWriter, r *http.Request) {
	var nl models.NewsList
	err := nl.GetLatestNews(db)
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	tmpl := textTemplate.Must(textTemplate.ParseFiles("./rss.xml"))
	err = tmpl.Execute(w, map[string]interface{}{
		"nl": nl,
	})
	if err != nil {
		http.Redirect(w, r, "/500", 302)
		log.Printf("%s\n", err)
		return
	}
	return
}
