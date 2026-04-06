# saga-courier-slack-app

Delivers tales to Slack channels using the Web API with Block Kit formatting. This produces richer messages than `slack-legacy`: a proper header for the title, structured sections for the body, and a subtle context footer for the version.

## Configuration

### Route config (in `.saga/config.jsonnet`)

| Key | Description |
|-----|-------------|
| `channel` | Slack channel to post to, including the leading `#` (e.g. `#engineering`). Must match `^#[a-z0-9_-]+$`. |

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGA_COURIER_SLACK_APP__BOT_TOKEN` | Bot User OAuth Token (`xoxb-...`). |

## Setting up the Slack app

### 1. Create the app

Go to [https://api.slack.com/apps](https://api.slack.com/apps) and click **Create New App**. Choose **From manifest** and pick the workspace to install it in. Then paste in this manifest (modify name):

```json
{
    "display_information": {
        "name": "Printy"
    },
    "features": {
        "bot_user": {
            "display_name": "Printy",
            "always_online": false
        }
    },
    "oauth_config": {
        "scopes": {
            "bot": [
                "chat:write",
                "chat:write.public"
            ]
        }
    }
}
```

### 2. Install to workspace

In the menu, navigate to **OAuth & Permissions**. Under the heading **OAuth Tokens**, click **Install to <Your Workspace>** and authorize. Copy the **Bot User OAuth Token** (starts with `xoxb-`).

### 3. Use the token

Store the token securely and export it as:

```sh
export SAGA_COURIER_SLACK_APP__BOT_TOKEN="xoxb-..."
```

The bot can post to any public channel immediately thanks to `chat:write.public`. For private channels, invite the bot first with `/invite @YourBotName`.
