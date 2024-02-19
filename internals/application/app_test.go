package application

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func getTemplateDir(t *testing.T) string {
	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%s/../../", path)
}

// getPageElementByXpathQuery uses an XPath expression to search for an element
// within the supplied document. If available, it returns it. Otherwise, it
// returns an error.
func getPageElementByXpathQuery(xpathQuery string, doc *html.Node) (*html.Node, error) {
	element := htmlquery.FindOne(doc, xpathQuery)
	if element == nil {
		return nil, errors.New("No element matching the supplied XPath expression was found.")
	}
	return element, nil
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
	title, err := getPageElementByXpathQuery("//title", doc)
	if err != nil || title == nil {
		t.Error("Title tag was not found")
	}
	if htmlquery.InnerText(title) != "404 - Not Found" {
		t.Errorf("got '%s'; want '%s'", htmlquery.InnerText(title), "404 - Not Found")
	}

	h1, err := getPageElementByXpathQuery("//h1", doc)
	if err != nil || h1 == nil {
		t.Error("h1 tag was not found")
	}
	if htmlquery.InnerText(h1) != "A Go URL Shortener" {
		t.Errorf("got '%s'; want '%s'", htmlquery.InnerText(h1), "A Go URL Shortener")
	}

	h2, err := getPageElementByXpathQuery("//h2", doc)
	if err != nil || h2 == nil {
		t.Error("h2 tag was not found")
	}
	expectedResult := "404 - Not Found"
	if htmlquery.InnerText(h2) != expectedResult {
		t.Errorf("got '%s'; want '%s'", htmlquery.InnerText(h2), expectedResult)
	}
}
