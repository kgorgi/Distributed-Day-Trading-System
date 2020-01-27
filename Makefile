SRC = ./src
OUTPUT = ../../build

.phony all:
all: build

# Build Commands
.phony build:
build: build-web build-transaction build-data build-audit build-generator

.phony build-web:
build-web: 
	cd $(SRC)/web && go build -o $(OUTPUT)/web

.phony build-transaction:
build-transaction: 
	cd $(SRC)/transaction && go build -o $(OUTPUT)/transaction

.phony build-data:
build-data: 
	cd $(SRC)/data && go build -o $(OUTPUT)/data

.phony build-audit:
build-audit: 
	cd $(SRC)/audit && go build -o $(OUTPUT)/audit

.phony build-generator:
build-generator: 
	cd $(SRC)/generator && go build -o $(OUTPUT)/generator

# Docker Compose Commands
.phony docker-deploy:
docker-deploy:
	docker-compose up --build

.phony docker-teardown:
docker-teardown:  
	docker-compose down

# Docker Container Commands
.phony docker-list:
docker-list: 
	docker ps

# Container Specific Commmands
# Example: make c=CONTAINER_NAME docker-shell 
.phony docker-shell:
docker-shell:
	docker exec -it $(c) bash

.phony docker-stop:
docker-stop:
	docker stop $(c)

.phony docker-start:
docker-start:
	docker start $(c)

.phony docker-remove:
docker-remove:
	docker rm $(c)