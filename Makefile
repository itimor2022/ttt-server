build:
	docker build --platform=linux/amd64 -t tsdd .
push:
	docker tag tsdd itimor2022/tsdd:2.0
	docker push itimor2022/tsdd:2.0
deploy:
	docker build --platform=linux/amd64 -t tsdd .
	docker tag tsdd itimor2022/tsdd:2.0
	docker push itimor2022/tsdd:2.0
run-dev:
	docker-compose build;docker-compose up -d
stop-dev:
	docker-compose stop
env-test:
	docker-compose -f ./testenv/docker-compose.yaml up -d 