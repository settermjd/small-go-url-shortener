package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"gourlshortener/internals/models"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"time"

	urlverifier "github.com/davidmytton/url-verifier"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"

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

// uniqid returns a unique id string useful when generating random strings.
// It was lifted from https://www.php2golang.com/method/function.uniqid.html.
func uniqid(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000

	return fmt.Sprintf("%s%08x%05x", prefix, sec, usec)
}

// GenerateShortenedURL generates and returns a short URL string.
func (a *App) GenerateShortenedURL() string {
	var (
		randomChars   = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
		randIntLength = 27
		stringLength  = 32
	)

	str := make([]rune, stringLength)

	for char := range str {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(randIntLength)))
		if err != nil {
			panic(err)
		}

		str[char] = randomChars[nBig.Int64()]
	}

	hash := sha256.Sum256([]byte(uniqid(string(str))))
	encodedString := base64.StdEncoding.EncodeToString(hash[:])

	return encodedString[0:9]
}

// getDefaultRoute retrieves a list of the stored shortened URLS and
// renders them in a table on the default route, along with a form for
// shortening a URL.
func (a *App) getDefaultRoute(w http.ResponseWriter, r *http.Request) {
	tmplFile := "./templates/default.html"
	tmpl, err := template.New("default.html").Funcs(functions).ParseFiles(tmplFile)
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

	session, err := store.Get(r, "flash-session")
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

// getAllUrls returns all of the available URL data in the urls table as a
// JSON-formatted string
func (a *App) getAllUrls(w http.ResponseWriter, r *http.Request) {
	urls, err := a.urls.Latest()
	if err != nil {
		fmt.Printf("Could not retrieve all URLs, because %s.\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`{"error":"Could not retrieve URL data."}`))
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(urls)
	if err != nil {
		fmt.Printf("Could not retrieve all URLs, because %s.\n", err)
		w.Write([]byte(`{"error":"Could not retrieve URL data."}`))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func setErrorInFlash(error string, w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "flash-session")
	if err != nil {
		fmt.Println(err.Error())
	}
	session.AddFlash(error, "error")
	session.Save(r, w)
}

var functions = template.FuncMap{
	"formatClicks": formatClicks,
}

func formatClicks(clicks int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%v", number.Decimal(clicks))
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

	// Verify if the URL supplied is a genuine and workable URL
	verifier := urlverifier.NewVerifier()
	verifier.EnableHTTPCheck()
	result, err := verifier.Verify(originalURL)

	if err != nil || !result.HTTP.IsSuccess {
		fmt.Println(err.Error())
		setErrorInFlash("The URL was not reachable.", w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	parsedUrl, err := url.Parse(originalURL)
	if err != nil {
		fmt.Println(err.Error())
		serverError(w, err)
		return
	}
	shortenedURL := parsedUrl.Scheme + "://" + a.GenerateShortenedURL()

	_, err = a.urls.Insert(originalURL, shortenedURL, 0)
	if err != nil {
		fmt.Println(err.Error())
		setErrorInFlash("We weren't able to shorten the URL.", w, r)
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

// routes creates the application's routing table
func (a *App) routes() http.Handler {
	router := httprouter.New()
	fileServer := http.FileServer(http.Dir("./static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Define the web-based routes
	router.HandlerFunc(http.MethodGet, "/", a.getDefaultRoute)
	router.HandlerFunc(http.MethodGet, "/open", a.openShortenedRoute)
	router.HandlerFunc(http.MethodPost, "/", a.shortenURL)

	// Define the API routes
	router.HandlerFunc(http.MethodGet, "/api/", a.getAllUrls)

	standard := alice.New()

	return standard.Then(router)
}

var store = sessions.NewCookieStore([]byte("a-secret-string"))

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
