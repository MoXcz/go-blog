package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

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

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, info.Mode())
}

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return CopyFile(path, target)
	})
}

func ClearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		p := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	return nil
}

func initializeBlog(rootPath string, posts []Post) error {
	if err := os.Mkdir(rootPath, 0755); err != nil {
		if errors.Is(err, os.ErrExist) {
			fmt.Printf("%s already exists!\n", rootPath)
			err = ClearDir(rootPath)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	err := CopyDir("static/", rootPath+"static/")
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

	err = indexPage(posts).Render(context.Background(), f)
	if err != nil {
		log.Fatalf("failed to write index page: %v", err)
	}

	fmt.Println("Blog initialized successfuly")
	return nil
}
