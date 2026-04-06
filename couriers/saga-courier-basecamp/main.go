package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/basecamp/basecamp-sdk/go/pkg/basecamp"
	"github.com/saga-changelog/saga/pkg/courier"
)

type basecampCourier struct{}

func (basecampCourier) Info() courier.Info {
	return courier.Info{
		Name:        "basecamp",
		Description: "Posts tales to a Basecamp message board",
		ConfigKeys: []courier.ConfigKey{
			{
				Name:        "project_id",
				Description: "Basecamp project ID (bucket ID)",
				Regex:       `^[0-9]+$`,
			},
			{
				Name:        "message_board_id",
				Description: "Basecamp message board ID",
				Regex:       `^[0-9]+$`,
			},
		},
	}
}

func (basecampCourier) ValidateRoute(_ context.Context, _ *courier.Route) error {
	return checkEnv()
}

func (basecampCourier) Tell(ctx context.Context, p *courier.Payload) error {
	if err := checkEnv(); err != nil {
		return err
	}

	accessToken, err := refreshAccessToken(ctx)
	if err != nil {
		return err
	}

	accountID := os.Getenv("SAGA_COURIER_BASECAMP__ACCOUNT_ID")
	boardID, err := strconv.ParseInt(p.Route.Config["message_board_id"], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid message_board_id: %w", err)
	}

	cfg := basecamp.DefaultConfig()
	token := &basecamp.StaticTokenProvider{Token: accessToken}
	client := basecamp.NewClient(cfg, token)
	account := client.ForAccount(accountID)

	msg, err := basecamp.NewMessagesService(account).Create(ctx, boardID, &basecamp.CreateMessageRequest{
		Subject: p.Tale.Title,
		Content: buildHTML(p),
		Status:  "active",
	})
	if err != nil {
		return fmt.Errorf("posting to Basecamp: %w", err)
	}

	fmt.Printf("Message posted to Basecamp: %s\n", msg.AppURL)
	return nil
}

func checkEnv() error {
	for _, key := range []string{
		"SAGA_COURIER_BASECAMP__REFRESH_TOKEN",
		"SAGA_COURIER_BASECAMP__CLIENT_ID",
		"SAGA_COURIER_BASECAMP__CLIENT_SECRET",
		"SAGA_COURIER_BASECAMP__ACCOUNT_ID",
	} {
		if os.Getenv(key) == "" {
			return fmt.Errorf("%s is not set", key)
		}
	}
	return nil
}

// refreshAccessToken exchanges the refresh token for a fresh access token.
func refreshAccessToken(ctx context.Context) (string, error) {
	form := url.Values{
		"type":          {"refresh"},
		"client_id":     {os.Getenv("SAGA_COURIER_BASECAMP__CLIENT_ID")},
		"client_secret": {os.Getenv("SAGA_COURIER_BASECAMP__CLIENT_SECRET")},
		"refresh_token": {os.Getenv("SAGA_COURIER_BASECAMP__REFRESH_TOKEN")},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://launchpad.37signals.com/authorization/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("creating refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("refreshing token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("token refresh returned empty access_token")
	}

	return result.AccessToken, nil
}

// buildHTML renders a tale body into the HTML subset Basecamp accepts,
// followed by a small italic footer line naming the chapter version.
// The tale title is not repeated in the body; it is sent as the
// Basecamp message subject.
func buildHTML(p *courier.Payload) string {
	var b strings.Builder

	for _, block := range p.Tale.Text.Blocks {
		switch block.Kind {
		case courier.BlockHeading:
			b.WriteString("<h2>")
			b.WriteString(renderInlinesHTML(block.Inlines))
			b.WriteString("</h2>\n")
		default:
			b.WriteString("<p>")
			b.WriteString(renderInlinesHTML(block.Inlines))
			b.WriteString("</p>\n")
		}
	}

	b.WriteString("<p><em>Released in ")
	b.WriteString(html.EscapeString(p.Chapter))
	b.WriteString("</em></p>\n")

	return b.String()
}

// renderInlinesHTML converts a flat Inline list to escaped HTML.
func renderInlinesHTML(inlines []courier.Inline) string {
	var b strings.Builder
	for _, in := range inlines {
		switch in.Style {
		case courier.InlineBold:
			b.WriteString("<strong>")
			b.WriteString(html.EscapeString(in.Text))
			b.WriteString("</strong>")
		case courier.InlineItalic:
			b.WriteString("<em>")
			b.WriteString(html.EscapeString(in.Text))
			b.WriteString("</em>")
		case courier.InlineLink:
			b.WriteString(`<a href="`)
			b.WriteString(html.EscapeString(in.URL))
			b.WriteString(`">`)
			b.WriteString(html.EscapeString(in.Text))
			b.WriteString("</a>")
		default:
			b.WriteString(html.EscapeString(in.Text))
		}
	}
	return b.String()
}

func main() {
	courier.Run(basecampCourier{})
}
