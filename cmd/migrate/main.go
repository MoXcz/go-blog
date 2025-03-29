package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", ":memory")
	if err != nil {
		fmt.Println("Error opening database", err)
		return
	}
	data, err := os.ReadFile("db/schema/schema.sql")
	if err != nil {
		fmt.Println("Error reading schema:", err)
		return
	}
	query := string(data)
	if _, err := db.Exec(query); err != nil {
		log.Fatalln("Error executing query", err)
	}
}
