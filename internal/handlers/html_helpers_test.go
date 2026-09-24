package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// fetchedPage is one HTTP response plus its parsed DOM, so render-contract
// assertions can inspect structure instead of substring-matching raw bytes.
type fetchedPage struct {
	URL    string
	Status int
	Header http.Header
	Body   string
	Doc    *html.Node
}

// fetch GETs url from server and parses the response body as HTML. A parse
// failure fails the test immediately; a non-HTML body still parses (into a
// document with little structure), which is itself a useful failure signal.
func fetch(t *testing.T, server *httptest.Server, url string) *fetchedPage {
	t.Helper()

	resp, err := http.Get(server.URL + url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("GET %s: read body: %v", url, err)
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("GET %s: parse HTML: %v", url, err)
	}

	return &fetchedPage{
		URL:    url,
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   string(body),
		Doc:    doc,
	}
}

// walk visits every element node in the document (including n itself).
func walk(n *html.Node, fn func(*html.Node)) {
	if n.Type == html.ElementNode {
		fn(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, fn)
	}
}

// elements returns every element node in doc, in document order.
func elements(doc *html.Node) []*html.Node {
	var out []*html.Node
	walk(doc, func(n *html.Node) { out = append(out, n) })
	return out
}

// attr returns the value of the named attribute on n, or "" if absent.
func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}

// byTag returns all elements in doc with the given tag name (lowercase).
func byTag(doc *html.Node, tag string) []*html.Node {
	var out []*html.Node
	walk(doc, func(n *html.Node) {
		if n.Data == tag {
			out = append(out, n)
		}
	})
	return out
}

// byTestID returns the first element with data-testid="id", or nil.
func byTestID(doc *html.Node, id string) *html.Node {
	for _, n := range elements(doc) {
		if attr(n, "data-testid") == id {
			return n
		}
	}
	return nil
}

// byTestIDAll returns every element with data-testid="id".
func byTestIDAll(doc *html.Node, id string) []*html.Node {
	var out []*html.Node
	for _, n := range elements(doc) {
		if attr(n, "data-testid") == id {
			out = append(out, n)
		}
	}
	return out
}

// textContent returns the concatenated text of n's subtree.
func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var rec func(*html.Node)
	rec = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}
	rec(n)
	return strings.TrimSpace(b.String())
}

// expectStatus fails the test when page.Status differs from want.
func expectStatus(t *testing.T, page *fetchedPage, want int) {
	t.Helper()
	if page.Status != want {
		t.Fatalf("GET %s: status = %d, want %d", page.URL, page.Status, want)
	}
}
