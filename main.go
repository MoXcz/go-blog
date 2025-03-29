package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/MoXcz/go-blog/handlers"
	"github.com/MoXcz/go-blog/internal/database"
	_ "modernc.org/sqlite"
)

func main() {
	mux := http.NewServeMux()
	listenAddr := ":3000"
	dbConn, err := sql.Open("sqlite", "./db/schema/foo.db")
	defer dbConn.Close()
	if err != nil {
		log.Fatalln(err)
	}
	dbQueries := database.New(dbConn)

	env := handlers.Env{
		DB: dbQueries,
	}

	mux.HandleFunc("GET /", env.HandleGetPosts)
	mux.HandleFunc("GET /posts/{blogTitle}", env.HandleGetPost)

	mux.HandleFunc("GET /search", env.HandleSearch)

	mux.HandleFunc("GET /create", handlers.HandleCreatePost)
	mux.HandleFunc("POST /create", env.HandleCreatePostSubmit)

	srv := http.Server{
		Handler: mux,
		Addr:    listenAddr,
	}

	log.Printf("Listening on port %s", listenAddr)
	log.Fatal(srv.ListenAndServe())
}
