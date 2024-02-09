package main

import (
	"database/sql"
	"fmt"
	"gourlshortener/internals/models"
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"

	_ "modernc.org/sqlite"
)

// This stores the template data for the default route
//
// This is the original URL that was submitted in the form, if any,
// the shortened URL version of the original URL, if the form was
// processed, and a list of already shortened URLs along with the
// number of times the shortened URL was clicked.
type PageData struct {
	Error, OriginalURL, ShortenedURL string
	URLData                          []*models.ShortenerData
}

type Shortener interface {
	Shorten() string
}

type App struct {
	db   *sql.DB
	urls *models.ShortenerDataModel
}

func serverError(w http.ResponseWriter, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func newApp(dbFile string) App {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return App{db: db, urls: &models.ShortenerDataModel{DB: db}}
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Ready to rock!")
	templatesFiles := []string{
		"./templates/default.html",
	}
	tmpl, err := template.ParseFiles(templatesFiles...)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
		return
	}

	urls, err := a.urls.Latest()
	if err != nil {
		fmt.Printf("Could not retrieve all URLs, because %s.\n", err)
		return
	}

	pageData := PageData{
		URLData: urls,
	}
	fmt.Println("Ready to write the template data to the ResponseWriter")
	err = tmpl.Execute(w, pageData)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
	}
}

func main() {
	app := newApp("data/database.sqlite3")


func (a *App) routes() http.Handler {
	router := httprouter.New()
	fileServer := http.FileServer(http.Dir("./static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/", a.getDefaultRoute)
	router.HandlerFunc(http.MethodPost, "/", a.processForm)
	standard := alice.New()

	return standard.Then(router)
}

func main() {
	app := newApp("data/database.sqlite3")
	addr := flag.String("addr", ":8080", "HTTP network address")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	defer app.db.Close()

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
