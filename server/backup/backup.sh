 #!/bin/bash

# Variables
BACKUP_DIR=./backup
DB_NAME=timeasy
DB_USER=myuser
DB_PASSWORD=mypassword
CONTAINER_NAME=timeasy-database

# Get current date and time for backup file
TIMESTAMP=$(date +"%F_%T")
BACKUP_FILE=$BACKUP_DIR/backup_$DB_NAME_$TIMESTAMP.sql

# Run pg_dump inside the PostgreSQL container
docker exec -t $CONTAINER_NAME pg_dump -U $DB_USER $DB_NAME > $BACKUP_FILE

echo "Backup completed: $BACKUP_FILE"
