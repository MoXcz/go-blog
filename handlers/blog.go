package handlers

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/MoXcz/go-blog/components"
	"github.com/MoXcz/go-blog/internal/database"
	"github.com/MoXcz/go-blog/types"
	"github.com/a-h/templ"
	"github.com/gosimple/slug"
	"github.com/yuin/goldmark"
)

type Env struct {
	DB *database.Queries
}

func newPost(title, content string) types.Post {
	return types.Post{
		Title:   title,
		Content: content,
	}
}

func HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	if err := components.CreatePostForm().Render(r.Context(), w); err != nil {
		return
	}
}

func (env *Env) HandleCreatePostSubmit(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		return
	}

	_, err := env.DB.CreatePost(r.Context(), database.CreatePostParams{
		Title:   slug.Make(title),
		Content: content,
	})
	if err != nil {
		log.Printf("error saving post: %v", err)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (env *Env) HandleGetPosts(w http.ResponseWriter, r *http.Request) {
	resPosts, err := env.DB.GetPosts(r.Context())
	if err != nil {
		log.Printf("error fetching posts: %v", err)
		return
	}

	var posts []types.Post
	for _, post := range resPosts {
		posts = append(posts, newPost(post.Title, post.Content))
	}

	component := components.Posts(posts)
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("failed to render component page: %v", err)
		return
	}
}

func (env *Env) HandleGetPost(w http.ResponseWriter, r *http.Request) {
	blogTitle := r.PathValue("blogTitle")
	post, err := env.DB.GetPost(r.Context(), (slug.Make(blogTitle)))
	if err != nil {
		log.Printf("error fetching post: %v", err)
		http.NotFound(w, r)
		return
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(post.Content), &buf); err != nil {
		log.Printf("failed to convert markdown: %v", err)
		return
	}

	content := Unsafe(buf.String())
	component := components.PostDetail(newPost(post.Title, post.Content), content)

	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("failed to render template: %v", err)
		return
	}
}

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

func (env *Env) HandleSearch(w http.ResponseWriter, r *http.Request) {
	resPosts, err := env.DB.GetPosts(r.Context())
	if err != nil {
		log.Printf("error fetching posts: %v", err)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("search"))
	var posts []types.Post
	for _, post := range resPosts {
		posts = append(posts, newPost(post.Title, post.Content))
	}

	var results []types.Post
	for _, post := range posts {
		if strings.Contains(strings.ToLower(post.Title), query) {
			results = append(results, post)
		}
	}

	components.PostList(results).Render(r.Context(), w)
}
