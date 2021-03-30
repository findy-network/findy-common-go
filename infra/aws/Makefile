build:
	docker build -t s3-tool .

run: build
	docker run -it --rm \
	-e AWS_DEFAULT_REGION=$(AWS_DEFAULT_REGION) \
	-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
	-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
	-v $(PWD)/temp:/temp \
	s3-tool $(S3_BUCKET_NAME) $(S3_BUCKET_FOLDER) /temp
