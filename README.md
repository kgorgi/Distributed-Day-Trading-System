# SENG 468

## Deploy Developer Environment

-   Deploy with `make docker-deploy-dev`
-   Uses the mock quote server
-   All containers are bound to ports on the host machine (hot-swappable)
-   To hot-swap a deployed container with a local instance
    1. Ensure that the developer environment is deployed (`make docker-deploy-dev`)
    2. Stop the docker container you want to swap with `make c=CONTAINER_NAME docker-stop`
    3. Start the server locally from the build folder
    4. The container has now been hot-swapped with a local instance

## Deploy Local Environment

-   Deploy with `make docker-deploy-local`
-   Uses the mocked quote server
-   Docker containers communicate using docker addresses
-   Only the web server is accessible locally from port 8080

## Deploy Lab Environment

-   Deploy with `make docker-deploy-lab`
-   Uses the actual legacy quote server
-   Docker containers communicate using docker addresses
-   Only the web server is accessible locally from port 8080
-   Must be deployed on the SENG 468 lab's linux virtual machines

## Container Ports for Reference

-   Web Server: 8080
-   Transaction Server: 5000
-   Database Server: 5001
-   Audit Server: 5002
-   Audit MongoDB: 5003
-   Database MongoDB: 27017
