#!/bin/bash

# Variables
BACKUP_FILE=$1
DB_NAME=timeasy
DB_USER=myuser
DB_PASSWORD=mypassword
CONTAINER_NAME=timeasy-database

# Restore the database
cat $BACKUP_FILE | docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME

echo "Database restored from: $BACKUP_FILE"
