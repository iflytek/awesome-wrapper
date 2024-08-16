#!/usr/bin/env bash

docker-compose -f server.yml down
docker-compose -f server.yml pull
docker-compose -f server.yml up --scale server=1 --scale server2=1 --scale server3=1
