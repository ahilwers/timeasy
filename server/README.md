# timeasy-server

This is the server component of timeasy. The server is not completed yet and isn't currently used at all.

## Development

### Authentication

The authentication is handled by Keycloak. To start Keycloak, just use the docker-compose.yaml in the "keycloak" directory. There's also a realm export file named realm-export.json in this directory that you can use to import the timeasy realm into Keycloak. Keep in mind that even after importing the realm you'll need to add users to your realm.

### Service-Testing

In the directory "service-testing" you'll find a script named "login.sh" that helps you authenticating with Keycloak. It uses [httpie](https://httpie.io/) to contact the Keycloak service and spawns a shell where the shell variable "TOKEN" is filled with the token Keycloak returned during login. You can then use this token for your REST calls:

```
http localhost:8080/api/v1/projects "Authorization:Bearer $TOKEN" < project.json
```
