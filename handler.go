package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/russross/blackfriday/v2"
)

const (
	INDEX              = "index"
	BLOG               = "blog"
	PHOTOS             = "photos"
	TEST               = "test"
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
		tigrisClient *s3.Client
	}
)

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func newHandler() (*Handler, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	endpoint := os.Getenv("AWS_ENDPOINT_URL_S3")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		slog.Error("creating aws config", "error", err)
		return nil, err
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &Handler{
		tigrisClient: client,
	}, nil
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

	sort.Slice(posts, func(i, j int) bool {
		date1, _ := time.Parse("01-02-2006", posts[i].Date)
		date2, _ := time.Parse("01-02-2006", posts[j].Date)
		return date1.After(date2)
	})

	pageData := map[string]any{
		"ActivePage": BLOG,
		"Posts":      posts,
	}
	tmpl.ExecuteTemplate(w, "blog.html", pageData)
}

// Photos is the handler function for the photos page. It returns a paginated list of photos.
func (h Handler) Photos(w http.ResponseWriter, r *http.Request) {
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

	// TODO: switch to using Tigris instead of embedding all the photos
	res, err := h.tigrisClient.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String("lvmnz-photos"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	start := len(res.Contents) - ((page - 1) * PHOTOS_PER_REQUEST) - 1
	end := max(start-PHOTOS_PER_REQUEST, 0)
	if end == 0 {
		// indicates that there are no more pages
		nextPage = -1
	}

	photos := make([]Photo, 0, PHOTOS_PER_REQUEST)
	for i := start; i > end; i -= 1 {
		item := res.Contents[i]
		name := aws.ToString(item.Key)
		url := fmt.Sprintf("https://fly.storage.tigris.dev/lvmnz-photos/%s", name)

		t, err := time.Parse("2006-01-02", strings.Split(name, "_")[0])
		if err != nil {
			slog.Error("parsing photo date", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info("photo", "name", name, "url", url, "date", t.Format("01/02/2006"), "index", start-i)
		photos = append(photos, Photo{
			Name:   name,
			Source: url,
			Date:   t.Format("01/02/2006"),
		})
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

func setupRoutes() error {
	h, err := newHandler()
	if err != nil {
		return err
	}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", h.Index)
	http.HandleFunc("/blog/", h.Blog)
	http.HandleFunc("/photos/", h.Photos)
	return nil
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
