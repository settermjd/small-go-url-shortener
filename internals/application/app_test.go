package application

import (
	"bytes"
	"errors"
	"fmt"
	"gourlshortener/internals/models/mocks"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/gorilla/sessions"
	"golang.org/x/net/html"
)

func getTemplateDir(t *testing.T) string {
	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%s/../../templates/", path)
}

// getPageElement uses an XPath expression to search for an element
// within the supplied document. If available, it returns it. Otherwise, it
// returns an error.
func getPageElement(xpathQuery string, doc *html.Node) (*html.Node, error) {
	element := htmlquery.FindOne(doc, xpathQuery)
	if element == nil {
		return nil, errors.New("No element matching the supplied XPath expression was found")
	}
	return element, nil
}

func getAllPageElements(xpathQuery string, doc *html.Node) ([]*html.Node, error) {
	elements, err := htmlquery.QueryAll(doc, xpathQuery)
	if err != nil || elements == nil {
		return nil, errors.New("No element matching the supplied XPath expression was found")
	}
	return elements, nil
}

func getPageElementCount(xpathQuery string, doc *html.Node) (int, error) {
	elements, err := htmlquery.QueryAll(doc, xpathQuery)
	if err != nil || elements == nil {
		return 0, errors.New("No element matching the supplied XPath expression was found")
	}
	return len(elements), nil
}

func TestPingRoute(t *testing.T) {
	app := &App{}

	ts := httptest.NewTLSServer(app.Routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/api/ping")
	if err != nil {
		t.Fatal(err)
	}

	if rs.StatusCode != http.StatusOK {
		t.Errorf("got %d; want %d", rs.StatusCode, http.StatusOK)
	}
}

type testServer struct {
    *httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func TestCanShortenUrl(t *testing.T) {
	app := &App{
		urls:            &mocks.ShortenerDataModel{},
		store:           sessions.NewCookieStore([]byte("this-is-a-test-key")),
		templateBaseDir: getTemplateDir(t),
	}

	ts := newTestServer(t, app.Routes())
	defer ts.Close()

	var form = url.Values{}
	form.Add("url", "https://osnews.com")
	rs, err := ts.Client().PostForm(ts.URL+"/", form)
	if err != nil {
		t.Fatal(err)
	}
	if rs.StatusCode != http.StatusSeeOther {
		t.Errorf("got %d; want %d", rs.StatusCode, http.StatusSeeOther)
	}
}

func TestCanRetrieveDefaultRoute(t *testing.T) {
	app := &App{
		urls:            &mocks.ShortenerDataModel{},
		store:           sessions.NewCookieStore([]byte("this-is-a-test-key")),
		templateBaseDir: getTemplateDir(t),
	}

	ts := httptest.NewTLSServer(app.Routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	if rs.StatusCode != http.StatusOK {
		t.Errorf("got %d; want %d", rs.StatusCode, http.StatusOK)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	rowCount, err := getPageElementCount("//table/tbody/tr", doc)
	if err != nil || rowCount != 1 {
		t.Error("Table row was not found")
	}
	tableRow, err := getAllPageElements("//table/tbody/tr", doc)
	if err != nil || tableRow == nil {
		t.Error("Table row with URL data was not found")
	}
	for i, n := range tableRow {
		td := htmlquery.FindOne(n, "//td")
		if i == 0 {
			if strings.TrimSpace(htmlquery.InnerText(td)) != "http://shorten3d" {
				t.Error("URL was not shortened correctly")
			}
		}
		if i == 1 {
			if strings.TrimSpace(htmlquery.InnerText(td)) != "https://osnews.com" {
				t.Error("Original URL was not returned")
			}
		}
		if i == 2 {
			if strings.TrimSpace(htmlquery.InnerText(td)) != "2,120" {
				t.Error("Incorrect number of clicks was returned")
			}
		}
	}

}

func Test404NotFoundRoute(t *testing.T) {
	app := &App{
		templateBaseDir: getTemplateDir(t),
	}

	ts := httptest.NewTLSServer(app.Routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/api/notfound")
	if err != nil {
		t.Fatal(err)
	}

	if rs.StatusCode != http.StatusNotFound {
		t.Errorf("got %d; want %d", rs.StatusCode, http.StatusNotFound)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	title, err := getPageElement("//title", doc)
	if err != nil || title == nil {
		t.Errorf("Title tag was not found")
	} else {
		if htmlquery.InnerText(title) != "404 - Not Found" {
			t.Errorf("got '%s'; want '%s'", htmlquery.InnerText(title), "404 - Not Found")
		}
	}

	h1, err := getPageElement("//h1", doc)
	if err != nil || h1 == nil {
		t.Error("h1 tag was not found")
	} else {
		if htmlquery.InnerText(h1) != "A Go URL Shortener" {
			t.Errorf("got '%s'; want '%s'", htmlquery.InnerText(h1), "A Go URL Shortener")
		}
	}

	h2, err := getPageElement("//h2", doc)
	if err != nil || h2 == nil {
		t.Error("h2 tag was not found")
	} else {
		expectedResult := "404 - Not Found"
		if htmlquery.InnerText(h2) != expectedResult {
			t.Errorf("got '%s'; want '%s'", htmlquery.InnerText(h2), expectedResult)
		}
	}
}
