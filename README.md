# Distributed Day Trading System

A distributed day trading system built with Docker, Golang, and MongoDB that supports over 1000 user transactions per second. Features of this system include: 
- high performance
- load balancing
- auditing
- data redunancy
- server redundancy
- caching

This project was built for the University of Victoria's Software Systems Scalability (SENG 468) class. (Group: Extreme Workload)

## Deploy Developer Environment

-   Deploy with `make docker-deploy-dev`
-   All containers deployed on one machine
-   Uses the mock quote server
-   All containers are bound to ports on the host machine (hot-swappable)
-   To hot-swap a deployed container with a local instance
    1. Ensure that the developer environment is deployed (`make docker-deploy-dev`)
    2. Stop the docker container you want to swap with `make c=CONTAINER_NAME docker-stop`
    3. Start the server locally from the build folder
    4. The container has now been hot-swapped with a local instance

## Deploy Local Environment

-   Deploy with `make docker-deploy-local`
-   Docker containers communicate using docker addresses
-   Uses the mocked quote server
-   Logging is disabled from docker-compose

## Deploy Lab Environment (Deprecated Due to Lack of Lab Access)

-   This deployment occures across 3 machines
-   To deploy the web servers and load balancer `make docker-deploy-lab-web`
-   To deploy the transaction and data server `make docker-deploy-lab-transaction`
-   To deploy the audit server `make docker-deploy-lab-audit`
-   Deploy all servers on one machine `make docker-deploy-lab-all` (must modify urls.yml)
-   Uses the actual legacy quote server
-   Docker containers communicate using IP addresses (except mongoDB containers and data server)
-   Must be deployed on the SENG 468 lab's linux virtual machines
-   Logging is disabled from docker-compose

## Deploy Dev Lab Environment (Deprecated Due to Lack of Lab Access)

-   This deployment occures across 3 machines
-   To deploy the web servers and load balancer `make docker-deploy-dev-lab-web`
-   To deploy the transaction and data server `make docker-deploy-dev-lab-transaction`
-   To deploy the audit server `make docker-deploy-dev-lab-audit`
-   Deploy all servers on one machine `make docker-deploy-dev-lab-all` (must modify urls.yml)
-   Uses the actual legacy quote server
-   Docker containers communicate using IP addresses (except mongoDB containers)
-   Must be deployed on the SENG 468 lab's linux virtual machines

## Local Deployment Available Ports:

-   Load Balancer for Web Servers: 8080
-   Web Server: 5005
-   Web Server 2: 5006
-   Load Balancer Transaction Servers: 5000
-   Transaction Server: 5007
-   Transaction Server 2: 5008
-   Audit Server: 5002
-   Audit MongoDB: 5003
-   Audit MongoDB2: 5098
-   Audit MongoDB3: 5099
-   Data MongoDB: 27017
-   Data MongoDB2: 27018
-   Data MongoDB3: 27019
-   Quote Cache Server: 5004
-   Mock Quote Server:
-   Watchdog: 4443 (if dev)

## Lab Deployment Available Ports (Deprecated Due to Lack of Lab Access)

Note: That these port mappings are out of date due to deprecation.

-   Web Load Balancer: 44410
-   Transaction Server: 44411
-   Audit Server: 44412
-   Database Server: 44413
-   Quote Cache Server: (Docker Address)
-   Audit MongoDB: N/A (Docker Address)
-   Database MongoDB: N/A (Docker Address)
-   Web Server: N/A (Docker Address)
-   Web Server 2: N/A (Docker Address)

## How to Execute Prototype Day Trading System

-   Set the correct access for the replica set key.
    `chmod 400 ./src/databases/data-keyfile`
-   Build and run normally
    `make docker-teardown`
    `make docker-deploy-dev`
-   In the mongo shell for both the data and audit databases
    `use admin`
    `db.auth("admin", "xxx")`
-   In the data mongo shell run
    `rs.initiate({ _id: "rs0", members: [ { _id: 0, host: "data-mongodb:27017" }, { _id: 1, host: "data-mongodb2:27017" }, { _id: 2, host: "data-mongodb3:27017" } ] });`
-   In the audit mongo shell run
    `rs.initiate({ _id: "rs0", members: [ { _id: 0, host: "audit-mongodb:27017" }, { _id: 1, host: "audit-mongodb2:27017" }, { _id: 2, host: "audit-mongodb3:27017" } ] });`
-   Then on the docker host machine restart the transaction servers and the audit server
    `docker restart transaction-server transaction-server2 audit-server`
