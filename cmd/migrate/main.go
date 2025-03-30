package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/MoXcz/go-blog/db/schema"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "db/schema/foo.db")
	if err != nil {
		fmt.Println("Error opening database", err)
		return
	}

	provider, err := goose.NewProvider(database.DialectSQLite3, db, schema.Embed)
	if err != nil {
		log.Fatal(err)
	}
	// List migration sources the provider is aware of.
	log.Println("\n=== migration list ===")
	sources := provider.ListSources()
	for _, s := range sources {
		log.Printf("%-3s %-2v %v\n", s.Type, s.Version, filepath.Base(s.Path))
	}

	// List status of migrations before applying them.
	stats, err := provider.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("\n=== migration status ===")
	for _, s := range stats {
		log.Printf("%-3s %-2v %v\n", s.Source.Type, s.Source.Version, s.State)
	}

	log.Println("\n=== log migration output  ===")
	results, err := provider.Up(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("\n=== migration results  ===")
	for _, r := range results {
		log.Printf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration)
	}
}
