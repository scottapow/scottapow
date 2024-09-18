include .env

deploy:
	gcloud run deploy ${GCLOUD_SERVICE_NAME} --source .

run:
	go run main.go