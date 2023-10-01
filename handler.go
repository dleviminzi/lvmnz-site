package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/russross/blackfriday/v2"
)

//go:embed content
var content embed.FS

//go:embed static/photos
var photos embed.FS

// Initialize templates
var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

type (
	Post struct {
		DisplayTitle string
		LinkTitle    string
		Date         string
	}

	Photo struct {
		Name   string
		Source string
		Date   string
	}

	Handler struct {
	}
)

func (h Handler) index(w http.ResponseWriter, r *http.Request) {
	// TODO: add ActivePage (generall pass in map object that contains whatever we need)
	// we can do this because it is all server side and I don't need to worry about
	// any "contract". What joy
	pageData := map[string]any{
		"ActivePage": "index",
	}
	tmpl.ExecuteTemplate(w, "index.html", pageData)
}

func (h Handler) blog(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/blog" || r.URL.Path == "/blog/" {
		h.blogList(w, r) // Call the blog handler
		return
	}

	title := path.Base(r.URL.Path)
	fmt.Printf("received request to fetch blog post: %s\n", title)
	markdown, err := fs.ReadFile(content, "content/"+title+".md")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	htmlContent := convertMarkdownToHTML(markdown)

	pageData := map[string]any{
		"ActivePage": "blog",
		"Content":    htmlContent,
	}
	tmpl.ExecuteTemplate(w, "blog_post.html", pageData)
}

func (h Handler) blogList(w http.ResponseWriter, r *http.Request) {
	posts, err := readBlogPostMetadata()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := map[string]any{
		"ActivePage": "blog",
		"Posts":      posts,
	}
	tmpl.ExecuteTemplate(w, "blog.html", pageData)
}

func (h Handler) photos(w http.ResponseWriter, r *http.Request) {
	dir, err := fs.ReadDir(photos, "static/photos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	photos := make([]Photo, len(dir))
	sort.Slice(dir, func(i, j int) bool {
		return dir[i].Name() > dir[j].Name()
	})
	for i, file := range dir {
		photos[i] = parsePhotoFile(file)
	}

	pageData := map[string]any{
		"ActivePage": "photos",
		"Photos":     photos,
	}
	err = tmpl.ExecuteTemplate(w, "photos.html", pageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func setupRoutes() {
	h := Handler{}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", h.index)
	http.HandleFunc("/blog/", h.blog)
	http.HandleFunc("/photos/", h.photos)
}

func convertMarkdownToHTML(markdown []byte) string {
	htmlContent := blackfriday.Run(markdown)
	return string(htmlContent)
}

func readBlogPostMetadata() ([]Post, error) {
	dir, err := fs.ReadDir(content, "content")
	if err != nil {
		return []Post{}, err
	}

	posts := make([]Post, len(dir))
	for i, file := range dir {
		fileName := strings.Split(file.Name(), "_")
		if len(fileName) != 2 {
			return []Post{}, errors.New("invalid post, missing metadata")
		}
		posts[i].DisplayTitle = strings.Join(strings.Split(fileName[0], "-"), " ")
		posts[i].LinkTitle = fileName[0]
		posts[i].Date = strings.Trim(fileName[1], ".md")
	}
	return posts, nil
}

func parsePhotoFile(file fs.DirEntry) Photo {
	source := "/static/photos/" + file.Name()
	fileName := strings.Split(file.Name(), ".jpeg")[0]

	metaDataParts := strings.Split(fileName, "_")
	dateParts := strings.Split(metaDataParts[0], "-")
	date := fmt.Sprintf("%s/%s/%s", dateParts[1], dateParts[2], dateParts[0])
	name := strings.Join(strings.Split(metaDataParts[1], "-"), " ")

	return Photo{
		Source: source,
		Date:   date,
		Name:   name,
	}
}
