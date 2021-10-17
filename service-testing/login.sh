
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
    --auth timeasy:356038e2-6469-4c9c-b861-8ed5f0cda283 \
    http://localhost:8180/auth/realms/timeasy/protocol/openid-connect/token \
    'Content-Type:application/x-www-form-urlencoded' \
    username=$username \
    password=$password \
    grant_type=password | jq --raw-output '.access_token'`

echo $TOKEN
fish