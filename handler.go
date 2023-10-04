package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/russross/blackfriday/v2"
)

const (
	INDEX              = "index"
	BLOG               = "blog"
	PHOTOS             = "photos"
	PHOTOS_PER_REQUEST = 15
)

var (
	tmpl *template.Template

	//go:embed content
	content embed.FS
)

type (
	Post struct {
		DisplayTitle string
		LinkTitle    string
		DisplayDate  string
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

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

// Index is the handler function for the index page.
func (h Handler) Index(w http.ResponseWriter, r *http.Request) {
	pageData := map[string]any{
		"ActivePage": INDEX,
	}
	tmpl.ExecuteTemplate(w, "index.html", pageData)
}

// Blog is the handler function for blog posts. It will deligate to BlogList if no
// specific post is requested.
func (h Handler) Blog(w http.ResponseWriter, r *http.Request) {
	// If the URL path is "/blog" or "/blog/", call the blogList handler.
	if r.URL.Path == "/blog" || r.URL.Path == "/blog/" {
		h.BlogList(w, r) // Call the blog handler
		return
	}

	title := path.Base(r.URL.Path)

	// Read the markdown content of the blog post from the file system.
	markdown, err := fs.ReadFile(content, "content/"+title+".md")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	htmlContent := convertMarkdownToHTML(markdown)

	pageData := map[string]any{
		"ActivePage": BLOG,
		"Content":    htmlContent,
	}
	tmpl.ExecuteTemplate(w, "blog_post.html", pageData)
}

// BlogList is the handler function for the blog list page.
func (h Handler) BlogList(w http.ResponseWriter, r *http.Request) {
	posts, err := readBlogPostMetadata()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := map[string]any{
		"ActivePage": BLOG,
		"Posts":      posts,
	}
	tmpl.ExecuteTemplate(w, "blog.html", pageData)
}

// Photos is the handler function for the photos page. It returns a paginated list of photos.
func (h Handler) Photos(w http.ResponseWriter, r *http.Request) {

	// Get the page number from the query string.
	var (
		page     int
		nextPage int
		err      error
	)
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if page < 1 {
		page = 1
	}
	nextPage = page + 1

	start := (page - 1) * PHOTOS_PER_REQUEST
	end := start + PHOTOS_PER_REQUEST

	dir, err := os.ReadDir("static/photos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Sort the directory in reverse order so that the newest photos are first.
	sort.Slice(dir, func(i, j int) bool {
		d1, _ := strings.CutPrefix(dir[i].Name(), "/static/photos/")
		d2, _ := strings.CutPrefix(dir[j].Name(), "/static/photos/")
		date1, _ := time.Parse("2006-01-02", strings.Split(d1, "_")[0])
		date2, _ := time.Parse("2006-01-02", strings.Split(d2, "_")[0])
		return date1.After(date2)
	})

	if end > len(dir) {
		end = len(dir)
		nextPage = -1
	}
	pagDir := dir[start:end]
	photos := make([]Photo, len(pagDir))
	for i, file := range pagDir {
		photos[i] = parsePhotoFile(file)
	}

	pageData := map[string]any{
		"ActivePage": PHOTOS,
		"Photos":     photos,
		"NextPage":   nextPage,
	}

	active_tmpl := "photos.html"
	if page > 1 {
		active_tmpl = "photos_gallery.html"
	}
	err = tmpl.ExecuteTemplate(w, active_tmpl, pageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func setupRoutes() {
	h := Handler{}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", h.Index)
	http.HandleFunc("/blog/", h.Blog)
	http.HandleFunc("/photos/", h.Photos)
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
		posts[i].DisplayDate = strings.Join(strings.Split(strings.Trim(fileName[1], ".md"), "-"), "/")
		posts[i].Date = strings.Trim(fileName[1], ".md")
	}
	return posts, nil
}

func parsePhotoFile(file fs.DirEntry) Photo {
	source := "/static/photos/" + file.Name()
	fileName := strings.Split(file.Name(), ".webp")[0]

	// The file name is in the format YYYY-MM-DD_name-of-photo.webp
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
