
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

http --form \
    --auth timeasy:YvvUGHtasN1drdIj1htni8g3woY5l5D6 \
    http://localhost:8180/realms/timeasy/protocol/openid-connect/token \
    'Content-Type:application/x-www-form-urlencoded' \
    username=$username \
    password=$password \
    grant_type=password

