package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/MoXcz/go-blog/internal/file"
	"github.com/yuin/goldmark"
)

type Meta struct {
	Title string
	Date  time.Time
}

func readPosts() ([]Post, error) {
	var posts []Post
	pages, err := filepath.Glob("./entries/*md")
	if err != nil {
		return nil, err
	}

	goldmark.New()
	for _, page := range pages {
		name := filepath.Base(page)
		content, err := os.ReadFile(page)
		if err != nil {
			fmt.Printf("Error: could not get content for page %s\n", name)
		}
		metadata, idx := readFrontmatter(content)

		posts = append(posts, Post{
			Content:  string(content[idx:]),
			Metadata: metadata,
		})
	}

	return posts, nil
}

const sep = "---"

func readFrontmatter(content []byte) (Meta, int) {
	const remainingSep = 4
	var m Meta

	// The content of the page CANNOT "---", as doing so will make this separation
	// to stop working and find the last index somewhere unexpected.
	// It subtracts 1 to remove a trailing newline
	idx := bytes.LastIndex(content, []byte(sep))

	// '4' here removes "---\n" at the beginning
	fms := bytes.SplitSeq(content[remainingSep:idx-1], []byte("\n"))

	// It expects something like this:
	// date: 01-Feb-2025
	// title: Title example
	for fm := range fms {
		sepData := bytes.Split(fm, []byte(":"))
		switch string(sepData[0]) {
		case "date":
			d := string(bytes.TrimSpace(sepData[1]))
			date, err := time.Parse("02-Jan-2006", d)
			if err != nil {
				fmt.Println("invalid date", err)
			}
			m.Date = date
		case "title":
			m.Title = string(bytes.TrimSpace(sepData[1]))
		default:
			fmt.Println("unhandled metadata")
		}
	}

	return m, idx + remainingSep
}

func (app *application) initializeBlog(rootPath string, posts []Post) error {
	if err := os.Mkdir(rootPath, 0755); err != nil {
		if errors.Is(err, os.ErrExist) {
			fmt.Printf("%s already exists!\n", rootPath)
			err = file.ClearDir(rootPath)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	err := file.CopyDir("static/", rootPath+"static/")
	if err != nil {
		return err
	}

	name := path.Join(rootPath, "index.html")
	f, err := os.Create(name)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			fmt.Println("index.html already exists!")
		} else {
			return fmt.Errorf("failed to create output file: %v", err)
		}
	}

	err = indexPage(posts, app.root).Render(context.Background(), f)
	if err != nil {
		log.Fatalf("failed to write index page: %v", err)
	}

	fmt.Println("Blog initialized successfuly")
	return nil
}
