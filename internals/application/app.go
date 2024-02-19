package application

import (
	"database/sql"
	"fmt"
	"gourlshortener/internals/models"
	"gourlshortener/internals/utils"
	"log"
	"net/http"
	"net/url"
	"text/template"
	"time"

	urlverifier "github.com/davidmytton/url-verifier"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
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

func serverError(w http.ResponseWriter, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type App struct {
	urls            *models.ShortenerDataModel
	store           *sessions.CookieStore
	templateBaseDir string
}

func NewApp(dbFile, authKey, templateBaseDir string) App {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return App{
		urls:            &models.ShortenerDataModel{DB: db},
		store:           sessions.NewCookieStore([]byte(authKey)),
		templateBaseDir: templateBaseDir,
	}
}

func (a *App) CloseDB() {
	a.urls.DB.Close()
}

func (a *App) setErrorInFlash(error string, w http.ResponseWriter, r *http.Request) {
	session, err := a.store.Get(r, "flash-session")
	if err != nil {
		fmt.Println(err.Error())
	}
	session.AddFlash(error, "error")
	session.Save(r, w)
}

// getDefaultRoute retrieves a list of the stored shortened URLS and
// renders them in a table on the default route, along with a form for
// shortening a URL.
func (a *App) getDefaultRoute(w http.ResponseWriter, r *http.Request) {
	tmplFile := fmt.Sprintf("%s/templates/default.html", a.templateBaseDir)
	tmpl, err := template.New("default.html").
		Funcs(template.FuncMap{
			"formatClicks": utils.FormatClicks,
		}).
		ParseFiles(tmplFile)
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

	session, err := a.store.Get(r, "flash-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		URLData: urls,
	}

	fm := session.Flashes("error")
	if fm != nil {
		if error, ok := fm[0].(string); ok {
			pageData.Error = error
		} else {
			fmt.Printf("Session flash did not contain an error message. Contained %s.\n", fm[0])
		}
	}
	session.Save(r, w)

	err = tmpl.Execute(w, pageData)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
	}
}

// shortenURL processes the URL shortener form. It generates a shortened
// URL for the original URL and stores them both in the database. After
// the details have been saved, the user is redirected to the default route.
func (a *App) shortenURL(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
		return
	}

	originalURL := r.PostForm.Get("url")
	if originalURL == "" {
		a.setErrorInFlash("Please provide a URL to shorten.", w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Verify if the URL supplied is a genuine and workable URL
	verifier := urlverifier.NewVerifier()
	verifier.EnableHTTPCheck()
	result, err := verifier.Verify(originalURL)

	if err != nil || !result.HTTP.IsSuccess {
		fmt.Println(err.Error())
		a.setErrorInFlash("The URL was not reachable.", w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	parsedUrl, err := url.Parse(originalURL)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
		return
	}
	shortenedURL := parsedUrl.Scheme + "://" + utils.GenerateShortenedURL()

	_, err = a.urls.Insert(originalURL, shortenedURL, 0)
	if err != nil {
		fmt.Println(err.Error())
		a.setErrorInFlash("We weren't able to shorten the URL.", w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fmt.Printf("Redirecting to the default route, after shortening %s to %s and persisting it.", originalURL, shortenedURL)

	// Redirect to the default route
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// openShortenedRoute retrieves the original URL from the shortened URL provided
// and, if retrieved from the database, redirects the user to the shortened URL.
func (a *App) openShortenedRoute(w http.ResponseWriter, r *http.Request) {
	shortenedURL := r.URL.Query().Get("url")
	fmt.Printf("Attempting to retrieve %s.\n", shortenedURL)

	urlData, err := a.urls.Get(shortenedURL)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
		return
	}

	err = a.urls.IncrementClicks(shortenedURL)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
		return
	}

	fmt.Printf("Redirecting to %s.\n", urlData.OriginalURL)

	// Redirect to the default route
	http.Redirect(w, r, urlData.OriginalURL, http.StatusSeeOther)
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	tmplFile := fmt.Sprintf("%s/templates/404.html", a.templateBaseDir)
	tmpl, err := template.New("404.html").ParseFiles(tmplFile)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	err = tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
	}
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	w.Write([]byte(fmt.Sprintf("%d", t.Unix())))
}

// routes creates the application's routing table
func (a *App) Routes() http.Handler {
	router := httprouter.New()
	fileServer := http.FileServer(http.Dir("./static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/", a.getDefaultRoute)
	router.HandlerFunc(http.MethodGet, "/open", a.openShortenedRoute)
	router.HandlerFunc(http.MethodPost, "/", a.shortenURL)
	router.HandlerFunc(http.MethodGet, "/api/ping", a.ping)
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.notFound(w, r)
	})
	standard := alice.New()

	return standard.Then(router)
}
