#!/bin/bash
# remove all containers logs and all stop images..
docker system prune -a -f 
sudo sh -c "truncate -s 0 /var/lib/docker/containers/*/*-json.log"