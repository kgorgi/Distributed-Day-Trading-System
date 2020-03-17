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
-   Logging is disabled from docker-compose

## Deploy Lab Environment

-   This deployment occures across 3 machines
-   To deploy the web servers and load balancer `make docker-deploy-lab-web`
-   To deploy the transaction and data server `make docker-deploy-lab-transaction`
-   To deploy the audit server `make docker-deploy-lab-audit`
-   Deploy all servers on one machine `make docker-deploy-lab-all` (must modify urls.yml)
-   Uses the actual legacy quote server
-   Docker containers communicate using IP addresses (except mongoDB containers and data server)
-   Must be deployed on the SENG 468 lab's linux virtual machines
-   Logging is disabled from docker-compose

## Deploy Dev Lab Environment

-   This deployment occures across 3 machines
-   To deploy the web servers and load balancer `make docker-deploy-dev-lab-web`
-   To deploy the transaction and data server `make docker-deploy-dev-lab-transaction`
-   To deploy the audit server `make docker-deploy-dev-lab-audit`
-   Deploy all servers on one machine `make docker-deploy-dev-lab-all` (must modify urls.yml)
-   Uses the actual legacy quote server
-   Docker containers communicate using IP addresses (except mongoDB containers)
-   Must be deployed on the SENG 468 lab's linux virtual machines

## Local Deployment Ports

-   Web Server: 8080
-   Transaction Server: 5000
-   Database Server: 5001
-   Audit Server: 5002
-   Audit MongoDB: 5003
-   Quote Cache Server: 5004
-   Database MongoDB: 27017

## Lab Deployment Ports

-   Web Load Balancer: 44410
-   Transaction Server: 44411
-   Audit Server: 44412
-   Database Server: 44413
-   Quote Cache Server: (Docker Address)
-   Audit MongoDB: N/A (Docker Address)
-   Database MongoDB: N/A (Docker Address)
-   Web Server: N/A (Docker Address)
-   Web Server 2: N/A (Docker Address)
