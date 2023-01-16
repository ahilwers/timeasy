
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

export TOKEN=`http POST http://localhost:8080/api/v1/login username=$username password=$password | jq --raw-output '.token'`

echo $TOKEN
bash
