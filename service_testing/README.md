# Service testing

To test the service I added some scripts that user httpie (https://httpie.io/) for accessing the service.

## Login

To login on KeyCloak you can just user the login.sh script and pass the user and the password as argument

    ./login.sh user password

This script sets the enironment variable *TOKEN* and opens a new fish shell where this variable is available. From now on you can fire your requests against the serice using the token.

If you want to login in with another user please exit (with the *exit* command) from the new shell before using the login script again.

## HTTP-Requests

To send a HTTP-request to the service with the token you can do it like so:

    http localhost:8080/api/admin "Authorization:Bearer $TOKEN"

 