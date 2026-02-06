dev:
	git ls-files | entr -c -r go run .

statics: static/htmx.min.js static/three-dots.svg

static/htmx.min.js:
	curl -o $@ https://unpkg.com/htmx.org@2.0.2/dist/htmx.min.js

static/three-dots.svg:
	curl -o $@ http://samherbert.net/svg-loaders/svg-loaders/three-dots.svg
