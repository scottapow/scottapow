include .env

deploy:
	gcloud run deploy ${GCLOUD_SERVICE_NAME} --source .