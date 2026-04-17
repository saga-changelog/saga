package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"os"
	"strings"

	"github.com/saga-changelog/saga/pkg/courier"
)

var (
	channelRe    = regexp.MustCompile(`^#[a-z0-9_-]+$`)
	linkSanitize = strings.NewReplacer("<", "", ">", "", "|", "")
)

type slackCourier struct{}

func (slackCourier) Info() courier.Info {
	return courier.Info{
		Name:        "slack-legacy",
		Description: "Delivers tales to Slack channels via incoming webhooks",
		ConfigKeys: []courier.ConfigKey{
			{
				Name:        "channel",
				Description: "Slack channel to post to (e.g., #engineering)",
				Regex:       `^#[a-z0-9_-]+$`,
			},
		},
	}
}

func (slackCourier) ValidateRoute(_ context.Context, route *courier.Route) error {
	if _, err := webhookURL(); err != nil {
		return err
	}
	ch := route.Config["channel"]
	if ch == "" {
		return fmt.Errorf("config key \"channel\" is required")
	}
	if !channelRe.MatchString(ch) {
		return fmt.Errorf("config key \"channel\" value %q does not match required pattern %s", ch, channelRe)
	}
	return nil
}

func (slackCourier) Tell(ctx context.Context, p *courier.Payload) error {
	webhook, err := webhookURL()
	if err != nil {
		return err
	}

	channel := p.Route.Config["channel"]
	text := buildMessage(p)

	body, err := json.Marshal(map[string]string{
		"channel": channel,
		"text":    text,
	})
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack API returned %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("Posted to %s\n", channel)
	return nil
}

// webhookURL reads SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL and validates it points
// at a Slack incoming webhook. Returns the raw URL string on success.
func webhookURL() (string, error) {
	raw := os.Getenv("SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL")
	if raw == "" {
		return "", fmt.Errorf("SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL is not set")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL is not a valid URL: %w", err)
	}
	if u.Scheme != "https" {
		return "", fmt.Errorf("SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL must use https (got %q)", u.Scheme)
	}
	if u.Host != "hooks.slack.com" {
		return "", fmt.Errorf("SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL host must be hooks.slack.com (got %q)", u.Host)
	}
	if !strings.HasPrefix(u.Path, "/services/") {
		return "", fmt.Errorf("SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL path must start with /services/ (got %q)", u.Path)
	}
	return raw, nil
}

// buildMessage formats a single tale for Slack. The tale title is shown
// as a bold header line, the body is rendered as mrkdwn, and the chapter
// version is rendered as a small italic footer at the bottom. Chapter
// version never appears as a title.
func buildMessage(p *courier.Payload) string {
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
// by blank lines. Headings become bold lines (mrkdwn has no real
// heading syntax).
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
		case courier.BlockCodeBlock:
			b.WriteString("```\n")
			b.WriteString(block.Text)
			b.WriteString("\n```")
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
		case courier.InlineCode:
			b.WriteString("`")
			b.WriteString(in.Text)
			b.WriteString("`")
		case courier.InlineLink:
			// Slack link format: <url|visible text>. Neither url nor
			// label may contain '<', '>', or '|'.
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
// message text: &, <, and >. Formatting characters (*, _) inside plain
// text are left alone because Slack only treats them as formatting when
// they surround a word; escaping them aggressively would be noisier
// than helpful for our content.
func escapeMrkdwn(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

func main() {
	courier.Run(slackCourier{})
}
