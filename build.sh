#!/bin/bash
PORT="3001"
# Append custom text to the Name
APP="golang-indicators-app"
### pull from server
git pull
### TO stop all containers Use This docker stop $(docker ps -aq)
### stop our Current Container
docker stop "${APP}"
### remove all containers use this docker rm $(docker ps -a -q)
### Remove Current Container
docker rm "${APP}"
### Prune everything  for all stopped containers
###The docker system prune command is a shortcut that prunes
### images, containers, and networks.
### Volumes are not pruned by default, and you must specify the --volumes flag for docker system prune to prune volumes.
# docker system prune -a -f
# docker system prune --volumes -f
### Remove ALL Logs
# sudo sh -c "truncate -s 0 /var/lib/docker/containers/*/*-json.log"
### build container
docker build -t tradingserviceapp .
### run docker
docker run -d --name "${APP}" -p $PORT:$PORT tradingserviceapp
docker update --restart on-failure "${APP}"
