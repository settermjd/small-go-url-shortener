package main

import (
	"database/sql"
	"flag"
	"gourlshortener/internals/application"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	env, ok := os.LookupEnv("ENVIRONMENT")
	if ok && strings.ToLower(env) == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	dbFile := strings.TrimPrefix(os.Getenv("DATABASE_URL"), "sqlite:")
	authKey := os.Getenv("AUTHENTICATION_KEY")
	templateBaseDir := os.Getenv("TEMPLATE_BASEDIR")
	staticDir := os.Getenv("STATIC_DIR")

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := application.NewApp(db, authKey, templateBaseDir, staticDir)
	addr := flag.String("addr", ":"+port, "HTTP network address")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.Routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
