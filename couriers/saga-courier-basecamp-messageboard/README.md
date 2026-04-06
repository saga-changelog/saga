# saga-courier-basecamp-messageboard

Delivers tales to a Basecamp message board.

## Configuration

### Route config (in `.saga/config.jsonnet`)

| Key | Description |
|-----|-------------|
| `project_id` | Basecamp project (bucket) ID, the number after `/buckets/` in the Basecamp URL |
| `message_board_id` | Message board ID, the number after `/message_boards/` in the URL |

Given a URL like `https://3.basecamp.com/ACCOUNT/buckets/PROJECT/message_boards/BOARD`:
- `project_id` = `PROJECT`
- `message_board_id` = `BOARD`
- `ACCOUNT` is the account ID (passed as an env var, not in the route config)

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGA_COURIER_BASECAMP_MESSAGEBOARD__ACCOUNT_ID` | Basecamp account ID (the number after `3.basecamp.com/` in the URL) |
| `SAGA_COURIER_BASECAMP_MESSAGEBOARD__CLIENT_ID` | OAuth app client ID |
| `SAGA_COURIER_BASECAMP_MESSAGEBOARD__CLIENT_SECRET` | OAuth app client secret |
| `SAGA_COURIER_BASECAMP_MESSAGEBOARD__REFRESH_TOKEN` | OAuth refresh token |

The courier exchanges the refresh token for a fresh access token on every invocation.

## Obtaining OAuth credentials

### 1. Register an integration

Go to [https://launchpad.37signals.com/integrations](https://launchpad.37signals.com/integrations) and create a new integration. Set the redirect URI to a URL you own and control, its fine if it is a non-existing page. For example `aDomainYouActuallyOwn.com/fake-basecamp-oauth-return-url`.

Note down the **client ID** and **client secret**.

### 2. Authorize

You can use Basecamp's "Authorization dialog Preview" on the integration page to authorize and obtain a verification code. Alternatively, visit:

```
https://launchpad.37signals.com/authorization/new?type=web_server&client_id=YOUR_CLIENT_ID&redirect_uri=YOUR_REDIRECT_URI
```

After authorizing, you'll be redirected to your redirect URI with a `code` parameter in the query string.

### 3. Exchange for tokens

```bash
curl -s -X POST https://launchpad.37signals.com/authorization/token \
  -d type=web_server \
  -d client_id=YOUR_CLIENT_ID \
  -d client_secret=YOUR_CLIENT_SECRET \
  -d redirect_uri=YOUR_REDIRECT_URI \
  -d code=VERIFICATION_CODE | jq .
```

This returns `access_token` and `refresh_token`. Store the **refresh token** securely; the courier uses it to obtain fresh access tokens automatically.

### 4. Find your account ID

Your account ID is visible in Basecamp URLs as the number after `https://3.basecamp.com/`. You can also list your accounts:

```bash
curl -s -H "Authorization: Bearer ACCESS_TOKEN" \
  https://launchpad.37signals.com/authorization.json | jq '.accounts[] | {id, name}'
```
