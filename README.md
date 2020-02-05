# SENG 468

## Deploy Developer Environment

-   Deploy with `make docker-deploy-dev`
-   All containers are accessible locally using localhost (hot-swappable)
-   Uses the mock quote server
-   Rapid Container Development
    1. Deploy the docker containers
    2. Stop the docker container you want to modify with `make c=CONTAINER_NAME docker-stop`
    3. Start the server you want locally from the build folder
    4. The locally executed server will use the docker containers

## Deploy Local Environment

-   Deploy with `make docker-deploy-local`
-   Docker containers communicate using docker addresses
-   Only the web server is accessible locally from port 8080
-   Uses the mocked quote server

## Deploy Lab Environment

-   Deploy with `make docker-deploy-lab`
-   Docker containers communicate using docker addresses
-   Only the web server is accessible locally from port 8080
-   Must be deployed on the lab linux machines
-   Uses the actual legacy quote server

## Container Ports for Local Testing

-   Web Server: 8080
-   Transaction Server: 5000
-   Database Server: 5001
-   Audit Server: 5002
-   Audit MongoDB: 5003
-   Database MongoDB: 27017
