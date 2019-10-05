package main

import (
	"database/sql"
	"flag"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
)

var router *mux.Router
var db *sql.DB
var session = sessions.NewCookieStore([]byte(os.Getenv("roman_session_key")))

func main() {
	session.MaxAge(0)

	var runScripts bool
	flag.BoolVar(&runScripts, "init-db", false, "Run 'schema.sql' scripts for database")
	flag.Parse()

	var err error
	db, err = connectionToMysqlServer()
	defer db.Close()
	if err != nil {
		log.Printf("ConnectionToMySqlServer: %s\n", err)
		return
	}

	if runScripts {
		err = createDbSchema()
		if err != nil {
			log.Printf("createDbSchema: %s\n", err)
			return
		}
	}

	router = routerInit()

	CSRF := csrf.Protect(
		[]byte(os.Getenv("roman_session_key")),
		csrf.FieldName("authenticity_token"),
		csrf.Secure(false),
		csrf.HttpOnly(false),
		csrf.Path("/"),
		csrf.MaxAge(0),
	)

	log.Fatal(http.ListenAndServe(":8085", CSRF(router)))
	return
}
