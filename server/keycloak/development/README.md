# Keycloak Authentication

The authentication is done by Keycloak. You can use this docker-compose for testing. There's also a realm-export.json in this directory that you can use to
import the realm into you Keycloak installation.

Remember to set the client secrets for the client "timeasy-server" correctly.

## Themes

You can add your themes to the themes directory. If you use the provided docker-compose.yaml these themes are imported
into Keycloak.

The Tailcloakify theme from https://github.com/ALMiG-Kompressoren-GmbH/tailcloakify is already included.
