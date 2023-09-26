package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/russross/blackfriday/v2"
)

//go:embed content
var content embed.FS

// Initialize templates
var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

type (
	PostList struct {
		Posts []Post
	}

	Post struct {
		DisplayTitle string
		LinkTitle    string
		Date         string
	}

	Handler struct {
	}
)

func (h Handler) index(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func (h Handler) about(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "about.html", nil)
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
	converted := map[string]any{
		"Content": htmlContent,
	}
	tmpl.ExecuteTemplate(w, "blog_post.html", converted)
}

func (h Handler) blogList(w http.ResponseWriter, r *http.Request) {
	posts, err := readBlogPostMetadata()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "blog.html", posts)
}

func setupRoutes() {
	h := Handler{}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", h.index)
	http.HandleFunc("/blog/", h.blog)
	http.HandleFunc("/about/", h.about)
}

func convertMarkdownToHTML(markdown []byte) string {
	htmlContent := blackfriday.Run(markdown)
	return string(htmlContent)
}

func readBlogPostMetadata() (PostList, error) {
	dir, err := fs.ReadDir(content, "content")
	if err != nil {
		return PostList{}, err
	}

	var postList PostList
	postList.Posts = make([]Post, len(dir))
	for i, file := range dir {
		fileName := strings.Split(file.Name(), "_")
		if len(fileName) != 2 {
			return PostList{}, errors.New("invalid post, missing metadata")
		}
		postList.Posts[i].DisplayTitle = strings.Join(strings.Split(fileName[0], "-"), " ")
		postList.Posts[i].LinkTitle = fileName[0]
		postList.Posts[i].Date = strings.Trim(fileName[1], ".md")
		fmt.Printf("found post: %s\n", fileName[0])
	}
	return postList, nil
}
