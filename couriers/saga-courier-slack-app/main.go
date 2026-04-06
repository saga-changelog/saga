package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/saga-changelog/saga/pkg/courier"
)

var (
	channelRe    = regexp.MustCompile(`^#[a-z0-9_-]+$`)
	linkSanitize = strings.NewReplacer("<", "", ">", "", "|", "")
)

type slackAppCourier struct{}

func (slackAppCourier) Info() courier.Info {
	return courier.Info{
		Name:        "slack-app",
		Description: "Delivers tales to Slack channels via the Web API with Block Kit formatting",
		ConfigKeys: []courier.ConfigKey{
			{
				Name:        "channel",
				Description: "Slack channel to post to (e.g., #engineering)",
				Regex:       `^#[a-z0-9_-]+$`,
			},
		},
	}
}

func (slackAppCourier) ValidateRoute(ctx context.Context, route *courier.Route) error {
	token, err := botToken()
	if err != nil {
		return err
	}
	ch := route.Config["channel"]
	if ch == "" {
		return fmt.Errorf("config key \"channel\" is required")
	}
	if !channelRe.MatchString(ch) {
		return fmt.Errorf("config key \"channel\" value %q does not match required pattern %s", ch, channelRe)
	}
	if err := authTest(ctx, token); err != nil {
		return err
	}
	return nil
}

// authTest calls Slack's auth.test endpoint to verify the bot token is
// valid and the bot is installed in a workspace. This is a lightweight,
// side-effect-free call.
func authTest(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://slack.com/api/auth.test", nil)
	if err != nil {
		return fmt.Errorf("creating auth.test request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling auth.test: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth.test returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parsing auth.test response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("auth.test failed: %s", result.Error)
	}

	return nil
}

func (slackAppCourier) Tell(ctx context.Context, p *courier.Payload) error {
	token, err := botToken()
	if err != nil {
		return err
	}

	channel := p.Route.Config["channel"]
	blocks := buildBlocks(p)
	fallback := buildFallback(p)

	payload := map[string]any{
		"channel": channel,
		"text":    fallback,
		"blocks":  blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://slack.com/api/chat.postMessage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting to Slack: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parsing Slack response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack API error: %s", result.Error)
	}

	fmt.Printf("Posted to %s\n", channel)
	return nil
}

// botToken reads SAGA_COURIER_SLACK_APP__BOT_TOKEN and validates it
// looks like a Slack bot token.
func botToken() (string, error) {
	raw := os.Getenv("SAGA_COURIER_SLACK_APP__BOT_TOKEN")
	if raw == "" {
		return "", fmt.Errorf("SAGA_COURIER_SLACK_APP__BOT_TOKEN is not set")
	}
	if !strings.HasPrefix(raw, "xoxb-") {
		return "", fmt.Errorf("SAGA_COURIER_SLACK_APP__BOT_TOKEN must start with xoxb-")
	}
	return raw, nil
}

// buildBlocks constructs a Block Kit block array for a tale.
//
//   - Title → header block (plain_text, Slack constraint)
//   - Body paragraphs → section blocks with mrkdwn
//   - Body headings → section blocks with bold mrkdwn
//   - Version footer → context block with mrkdwn
func buildBlocks(p *courier.Payload) []map[string]any {
	var blocks []map[string]any

	// Title as header block.
	blocks = append(blocks, map[string]any{
		"type": "header",
		"text": map[string]any{
			"type": "plain_text",
			"text": p.Tale.Title,
		},
	})

	// Body blocks.
	for _, block := range p.Tale.Text.Blocks {
		var text string
		switch block.Kind {
		case courier.BlockHeading:
			text = "*" + renderInlines(block.Inlines) + "*"
		default:
			text = renderInlines(block.Inlines)
		}
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": text,
			},
		})
	}

	// Version footer as context block.
	blocks = append(blocks, map[string]any{
		"type": "context",
		"elements": []map[string]any{
			{
				"type": "mrkdwn",
				"text": "_Released in " + escapeMrkdwn(p.Chapter) + "_",
			},
		},
	})

	return blocks
}

// buildFallback produces a plain mrkdwn string used as the top-level
// "text" field. Slack uses this for notifications and accessibility
// when blocks are present.
func buildFallback(p *courier.Payload) string {
	var b strings.Builder

	b.WriteString("*")
	b.WriteString(escapeMrkdwn(p.Tale.Title))
	b.WriteString("*\n")

	text := renderText(p.Tale.Text)
	if text != "" {
		b.WriteString("\n")
		b.WriteString(text)
		b.WriteString("\n")
	}

	b.WriteString("\n_Released in ")
	b.WriteString(escapeMrkdwn(p.Chapter))
	b.WriteString("_")

	return b.String()
}

// renderText renders a TaleText to Slack mrkdwn. Blocks are separated
// by blank lines. Headings become bold lines.
func renderText(text courier.TaleText) string {
	var b strings.Builder
	for i, block := range text.Blocks {
		if i > 0 {
			b.WriteString("\n\n")
		}
		switch block.Kind {
		case courier.BlockHeading:
			b.WriteString("*")
			b.WriteString(renderInlines(block.Inlines))
			b.WriteString("*")
		default:
			b.WriteString(renderInlines(block.Inlines))
		}
	}
	return b.String()
}

// renderInlines flattens a list of Inlines to a Slack mrkdwn string.
func renderInlines(inlines []courier.Inline) string {
	var b strings.Builder
	for _, in := range inlines {
		switch in.Style {
		case courier.InlineBold:
			b.WriteString("*")
			b.WriteString(escapeMrkdwn(in.Text))
			b.WriteString("*")
		case courier.InlineItalic:
			b.WriteString("_")
			b.WriteString(escapeMrkdwn(in.Text))
			b.WriteString("_")
		case courier.InlineLink:
			b.WriteString("<")
			b.WriteString(linkSanitize.Replace(in.URL))
			b.WriteString("|")
			b.WriteString(linkSanitize.Replace(in.Text))
			b.WriteString(">")
		default:
			b.WriteString(escapeMrkdwn(in.Text))
		}
	}
	return b.String()
}

// escapeMrkdwn escapes the three characters Slack treats specially in
// message text: &, <, and >.
func escapeMrkdwn(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

func main() {
	courier.Run(slackAppCourier{})
}
