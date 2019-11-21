# Deployments


## Compose

```
export DATAVOL=/home/fils/Data/P418/dataVolumes
```

## Swarm
if you have not:
docker swarm init

```
docker stack deploy --compose-file=gleaner-dev_generic.yaml gleaner

docker-compose -f gleaner-compose.yml down
docker-compose -f gleaner-compose.yml up -d
docker-compose -f gleaner-dev_generic.yaml up -d
docker-compose -f gleaner-dev_generic.yaml down

docker rm $(docker ps -a -f status=exited -q)

docker stack deploy --compose-file=gleaner-dev_generic.yaml gleaner
docker stack ls
docker stack ps gleaner
docker stack services gleaner
docker stack rm gleaner
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)
docker ps
```

```
docker-compose up -d
docker-compose scale web=3
docker-compose ps you can see there are 4 containers run
```

-------
windows
----
Modify windows.env with path to storage: DATAVOL=
```
docker-compose -f .\gleaner-dev.yaml -f .\gleaner-dev.windows.yaml up -d
docker-compose ps you can see there are 4 containers run
```