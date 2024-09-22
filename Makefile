include .env

deploy:
	gcloud run deploy scottapow --source ./ --update-secrets GOOGLE_KEY=GOOGLE_KEY:latest,GOOGLE_SECRET=GOOGLE_SECRET:latest,SESSION_SECRET=SESSION_SECRET:latest,DB_CONN_STR=DB_CONN_STR:latest --update-env-vars HOST=https://www.scottpowell.dev --format json --region us-east1

run:
	go run main.go