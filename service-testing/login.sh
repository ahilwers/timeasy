
#!/bin/bash

if [ -z "$1" ]; then
    echo 'You must specify a username and a password'
    exit 1
fi

if [ -z "$2" ]; then
    echo 'You must specify a username and a password'
    exit 1
fi

username=$1
password=$2

echo $username
echo $password

export TOKEN=`http --form \
    --auth timeasy-server:eLhnG89XXcG0qtQ6xs05klSBMaxQ89Fd \
    http://localhost:8180/realms/timeasy/protocol/openid-connect/token \
    'Content-Type:application/x-www-form-urlencoded' \
    username=$username \
    password=$password \
    grant_type=password | jq --raw-output '.access_token'`

echo $TOKEN
bash
