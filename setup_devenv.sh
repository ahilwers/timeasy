#!/bin/sh

cd keycloak
docker-compose up -d
cd ..
cd postgresql
docker-compose up -d
