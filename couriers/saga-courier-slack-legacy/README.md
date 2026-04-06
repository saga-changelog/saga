# saga-courier-slack-legacy

Delivers tales to a Slack channel via an incoming webhook.

## Configuration

### Route config (in `.saga/config.jsonnet`)

| Key | Description |
|-----|-------------|
| `channel` | Slack channel to post to, including the leading `#` (e.g. `#engineering`). Must match `^#[a-z0-9_-]+$`. |

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL` | Slack incoming webhook URL. Must be HTTPS and point at `hooks.slack.com/services/...`. |

The courier validates the webhook URL during `validate-route`, so a typo is caught before any `tell` calls start.

## Obtaining a webhook URL

1. Go to [https://api.slack.com/apps](https://api.slack.com/apps) and create a new app (or pick an existing one).
2. Under **Features** choose **Incoming Webhooks** and toggle them on.
3. Click **Add New Webhook to Workspace** and pick the channel you want to post to. Slack generates a URL of the form `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`.
4. Store that URL securely. It is the only secret the courier needs.

The webhook is bound to the channel you picked when you installed it, but the `channel` route config still takes effect: saga sends it in the payload so the same webhook can deliver to different channels from different routes (provided the Slack app has permission to post there).
