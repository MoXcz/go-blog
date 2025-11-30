package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path"

	"github.com/a-h/templ"
	"github.com/gosimple/slug"
	"github.com/yuin/goldmark"
)

type Post struct {
	Metadata Meta
	Content  string
}

type application struct {
	root string
}

func main() {
	root := flag.String("root", "", "root value to link static files")
	flag.Parse()

	app := &application{
		root: *root,
	}

	posts, err := readPosts()
	if err != nil {
		log.Fatal(err)
	}

	err = app.initializeBlog("docs/", posts)
	if err != nil {
		log.Fatal(err)
	}

	// Create a page for each post.
	for _, post := range posts {
		// Create the output directory.
		dir := path.Join("docs/", post.Metadata.Date.Format("2006/01/02"), slug.Make(post.Metadata.Title))
		if err := os.MkdirAll(dir, 0755); err != nil && err != os.ErrExist {
			log.Fatalf("failed to create dir %q: %v", dir, err)
		}

		// Create the output file.
		name := path.Join(dir, "index.html")
		f, err := os.Create(name)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}

		// Convert the markdown to HTML, and pass it to the template.
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(post.Content), &buf); err != nil {
			log.Fatalf("failed to convert markdown to HTML: %v", err)
		}

		// Create an unsafe component containing raw HTML.
		content := Unsafe(buf.String())

		// Use templ to render the template containing the raw HTML.
		err = contentPage(post.Metadata.Title, content, app.root).Render(context.Background(), f)
		if err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
	}
}

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}
