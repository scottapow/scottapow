include .env

deploy:
	gcloud run deploy scottapow --source ./ --update-secrets GOOGLE_KEY=GOOGLE_KEY:latest,GOOGLE_SECRET=GOOGLE_SECRET:latest,SESSION_SECRET=SESSION_SECRET:latest,DB_CONN_STR=DB_CONN_STR:latest,JWT_SECRET=JWT_SECRET:latest --update-env-vars HOST=https://www.scottpowell.dev --format json --region us-east1

run/air-go:
	go run github.com/cosmtrek/air@v1.51.0 \
		--build.cmd "go build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
		--build.exclude_dir "web/node_modules" \
		--build.include_ext "go" \
		--build.stop_on_error "false" \
		--misc.clean_on_exit true
run/air-assets:
	go run github.com/cosmtrek/air@v1.51.0 \
		--build.cmd "templ generate --notify-proxy" \
		--build.bin "true" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_dir "web/public" \
		--build.include_ext "js,css"
run/templ:
	templ generate --watch --proxy="http://localhost:3000" --cmd="go run *.go" --open-browser=false
run/tailwind:
	cd web && npm run build:css:watch
run:
	make -j4 run/templ run/air-go run/tailwind run/air-assets

buildweb:
	cd web && npm run build:css