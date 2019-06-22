AWS_DEFAULT_REGION ?= eu-west-1

build: 
	docker run --rm \
	--volume ${CURDIR}:/app \
	--workdir /app \
	vjftw/vuepress \
	build

deploy:
	docker run --rm \
	--volume ${CURDIR}:/app \
	--workdir /app \
	--env AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
	--env AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
	--env AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
	anigeo/awscli:latest \
	s3 cp .vuepress/dist/. s3://velocityci.io --acl public-read --recursive --cache-control max-age=120

serve:
	docker run --rm \
	-it \
	--volume ${CURDIR}:/app \
	--workdir /app \
	-p 8080:8080 \
	vjftw/vuepress \
	dev --host 0.0.0.0