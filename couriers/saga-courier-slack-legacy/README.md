# saga-courier-slack-legacy

Delivers tales to a Slack channel via an incoming webhook. This aproach has been named "legacy" by Slack a long time ago, but continues to work fine. Setting it up is fast and easy. The downside is that it has limited capability of formatting messages. For nicer looking tales in slack, use the `slack-app` courier.

## Configuration

### Route config (in `.saga/config.jsonnet`)

| Key | Description |
|-----|-------------|
| `channel` | Slack channel to post to, including the leading `#` (e.g. `#engineering`). Must match `^#[a-z0-9_-]+$`. |

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL` | Slack incoming webhook URL. Must be HTTPS and point at `hooks.slack.com/services/...`. |

The courier validates the webhook URL during `validate-route`, so a typo is caught before any `tell` calls start. Actual validity/authority of the webhook URL is not verified beforehand.

## Obtaining a webhook URL

1. Go to `https://YOUR-WORKSPACE-NAME.slack.com/marketplace/A0F7XDUAZ-incoming-webhooks` (Add your workspace url name in the URL)
2. Click **Add to Slack**.
3. Pick a channel and **Add Incoming WebHooks Integration**. Slack generates a URL of the form `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`.
4. Store that URL securely. It is the only secret the courier needs.

The webhook is bound to the channel you picked when you installed it, but the `channel` route config still takes effect: saga sends it in the payload so the same webhook can deliver to different channels from different routes (provided the Slack app has permission to post there).
