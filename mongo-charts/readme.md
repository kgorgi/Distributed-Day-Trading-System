# Mongo Charts  
## Deployment
I used the instructions at this link: https://docs.mongodb.com/charts/onprem/installation/

Here's how I did it locally:  
1. cd into this directory
2. `docker run -d -p 6004:27017 --name chartsDB mongo`  
   This container will be used to store the mongo-charts config
3. `docker swarm init`
4. `docker pull quay.io/mongodb/charts:19.12.1`
5. `docker run --rm quay.io/mongodb/charts:19.12.1 charts-cli test-connection 'mongodb://host.docker.internal:6004'`  
   Make sure that the connection worked
6. `echo "mongodb://host.docker.internal:6004" | docker secret create charts-mongodb-uri -`
7. `docker stack deploy -c charts-docker-swarm-19.12.1.yml mongodb-charts`
8. `docker service ls`  
   Make sure that the service is up (REPLICAS  1/1)
9. `docker service logs <service ID>`
    Make sure that there's no errors
10. `docker exec -it 5e2dbf100123 charts-cli add-user --first-name "goh" --last-name "dato" --email "goh@goh.ca" --password "1234567" --role "UserAdmin"`
11. Visit `localhost:8889`
12. Log in


To clear everything while trouble shooting:  
`docker stack rm mongodb-charts`  
`docker secret rm charts-mongodb-uri`  
`docker volume prune`  
`docker swarm init` again? (IDK if this does anything)
If the db got messed up:  
`docker stop chartsDB`
`docker rm chartsDB`


## Using Charts

connect to DBs:  
`mongodb://admin:admin@host.docker.internal:27017`  
-> extremeworkload -> system.profile

`mongodb://admin:admin@host.docker.internal:5003`  

The profiler logging is sort of cryptic.
Here are some useful fields:
- ns -> name space, this gets collection names
- op -> operation, i.e. {query,update,command,...}
- millis -> time spent on execution  
- ts -> timestamp of execution

Create an extra field in the dataDB by using this:  
`{"$arrayElemAt": [{"$objectToArray":"$$ROOT.command"},0]}`  
This will give extra information on the command type
