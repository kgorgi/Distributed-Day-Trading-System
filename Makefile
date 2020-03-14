SHELL := /bin/bash

SRC = ./src
OUTPUT = ../../build

.phony all:
all: build

# Build Commands
.phony build:
build: build-web build-transaction build-data build-audit build-generator build-quote-mock build-quote-cache

.phony build-web:
build-web: 
	cd $(SRC)/web && go build -o $(OUTPUT)/web.exe

.phony build-transaction:
build-transaction: 
	cd $(SRC)/transaction && go build -o $(OUTPUT)/transaction.exe

.phony build-data:
build-data: 
	cd $(SRC)/data && go build -o $(OUTPUT)/data.exe

.phony build-audit:
build-audit: 
	cd $(SRC)/audit && go build -o $(OUTPUT)/audit.exe

.phony build-generator:
build-generator: 
	cd $(SRC)/generator && go build -o $(OUTPUT)/generator.exe

.phony build-quote-mock:
build-quote-mock: 
	cd $(SRC)/quote-mock && go build -o $(OUTPUT)/quote-mock.exe

.phony build-quote-cache:
build-quote-cache: 
	cd $(SRC)/quote-cache && go build -o $(OUTPUT)/quote-cache.exe

.phony format:
format:
	gofmt -w ./src

.phony test-e2e:
test-e2e:
	cd $(SRC)/test/end-to-end && go test -v

# Docker Local Deployment Commands
.phony docker-deploy-dev:
docker-deploy-dev:
	docker-compose -f docker-compose.yml -f docker-compose.local.yml -f docker-compose.dev.yml --compatibility up --build

.phony docker-deploy-dev-d:
docker-deploy-dev-d:
	docker-compose -f docker-compose.yml -f docker-compose.local.yml -f docker-compose.dev.yml --compatibility up --build -d

.phony docker-deploy-local:
docker-deploy-local:
	docker-compose -f docker-compose.yml -f docker-compose.local.yml up --build -d

.phony docker-redeploy-dev:
docker-redeploy:
	docker-compose -f docker-compose.yml -f docker-compose.local.yml -f docker-compose.dev.yml --compatibility up --build --force-recreate --no-deps -d $(c)

# Docker Lab Deployment Commands 
LAB_DEPLOY = docker-compose -f docker-compose.yml -f docker-compose.lab.yml up --build -d

.phony docker-deploy-lab-all:
docker-deploy-lab: build
	$(LAB_DEPLOY)

.phony docker-deploy-lab-web:
docker-deploy-lab-web: build
	$(LAB_DEPLOY) load web web2  

.phony docker-deploy-lab-transaction:
docker-deploy-lab-transaction: build
	$(LAB_DEPLOY) transaction data dataDB quote-cache

.phony docker-deploy-lab-audit:
docker-deploy-lab-audit: build
	$(LAB_DEPLOY) audit auditDB 

# Docker Lab Dev Deployment Commands 
LAB_DEV_DEPLOY = docker-compose -f docker-compose.yml -f docker-compose.lab.yml -f docker-compose.dev-lab.yml up --build

.phony docker-deploy-dev-lab-all:
docker-deploy-dev-lab-all: build
	$(LAB_DEV_DEPLOY)

.phony docker-deploy-dev-lab-web:
docker-deploy-dev-lab-web: build
	$(LAB_DEV_DEPLOY) load web web2 

.phony docker-deploy-dev-lab-transaction:
docker-deploy-dev-lab-transaction: build
	$(LAB_DEV_DEPLOY) transaction data dataDB quote-cache

.phony docker-deploy-dev-lab-audit:
docker-deploy-dev-lab-audit: build
	$(LAB_DEV_DEPLOY) audit auditDB 

# Docker Cleanup
.phony docker-teardown:
docker-teardown:  
	docker-compose down --remove-orphans -v

.phony docker-clean:
docker-clean:
	docker system prune && docker volume prune

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

# Utility Commands
.phony cert-generate:
cert-generate:
	cd ./ssl && sudo openssl req -newkey rsa:2048 -nodes -keyout key.pem -x509 -days 365 -out cert.pem -subj '/CN=localhost'