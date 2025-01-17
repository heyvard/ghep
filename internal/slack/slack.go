package slack

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"text/template"
	"time"
)

var (
	//go:embed templates/*.tmpl
	templates  embed.FS
	commitTmpl *template.Template
)

type Client struct {
	httpClient *http.Client
	token      string
}

func New(token string) (Client, error) {
	if token == "" {
		return Client{}, fmt.Errorf("missing Slack token")
	}

	var err error
	commitTmpl, err = template.ParseFS(templates, "templates/commit.tmpl")
	if err != nil {
		return Client{}, err
	}

	return Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		token: token,
	}, nil
}

func (c Client) PostMessage(payload []byte) error {
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type slackResponse struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
		Warn  string `json:"warning"`
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error posting message to Slack(%v): %v", resp.Status, body)
	}

	var slackResp slackResponse
	if err := json.Unmarshal([]byte(body), &slackResp); err != nil {
		return err
	}

	if !slackResp.Ok {
		return fmt.Errorf("error posting message to Slack: %v", slackResp.Error)
	}

	if slackResp.Warn != "" {
		slog.Info("warning posting message to Slack", "warn", slackResp.Warn)
	}

	return nil
}
