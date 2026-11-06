# Manual API testing

These files are development aids for trying the REST API. Start the server, PostgreSQL and Keycloak as described in the [server README](../README.md); create a test user and obtain a Keycloak access token for that user. The examples below require [HTTPie](https://httpie.io/) and assume the API is listening on `localhost:8080`.

Set `TOKEN` in your shell to the access token using a secure method; do not commit it or share it in logs. Then, for example:

```bash
http GET localhost:8080/api/v1/billing/status
http GET localhost:8080/api/v1/projects "Authorization:Bearer $TOKEN"
http GET localhost:8080/api/v1/timeentries "Authorization:Bearer $TOKEN"
```

The billing-status endpoint is public; projects and time entries require a valid token and, if subscription checks are enabled, an active entitlement or an allowlisted user. Tokens expire; obtain a new one if authentication fails.

[`project.json`](project.json) and [`timeentry.json`](timeentry.json) are sample request bodies. Projects require a `name` and `color`; time entries require a `projectId` and RFC 3339 `startTime` (with optional `endTime`). Before posting a time entry, replace the example `projectId` with an existing project's ID; the sample timestamps are fixed example values. For example, from this directory:

```bash
http POST localhost:8080/api/v1/projects "Authorization:Bearer $TOKEN" < project.json
```

## Legacy login helper

[`login.sh`](login.sh) is an old local-testing helper, not a secure authentication workflow. It contains a hard-coded development client secret, sends a password grant to `localhost:8180`, prints the username, password and token, and starts a new Bash shell. The credentials may be visible in terminal output, process listings and shell history. The secret may not match your Keycloak setup. Do not use this script with real credentials or production systems; obtain a token through your configured authentication flow instead.
