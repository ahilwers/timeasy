#!/bin/sh

export _TEST_QUARKUS_DATASOURCE_JDBC_DRIVER=org.testcontainers.jdbc.ContainerDatabaseDriver
./mvnw quarkus:dev


