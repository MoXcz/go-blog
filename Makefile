build:
	templ generate && go run . -root="go-blog"

run:
	templ generate && go run .
