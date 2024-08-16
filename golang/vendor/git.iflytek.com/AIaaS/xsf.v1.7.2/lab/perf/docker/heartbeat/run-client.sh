#!/usr/bin/env bash

docker-compose -f client.yml down
docker-compose -f client.yml pull
docker-compose -f client.yml up
