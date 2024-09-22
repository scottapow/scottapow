include .env

deploy:
	gcloud run deploy scottapow --source ./ --format json --region us-east1

run:
	go run main.go