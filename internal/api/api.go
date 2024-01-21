package api

import (
	"log/slog"
	"net/http"

	"github.com/navikt/ghep/internal/github"
	"github.com/navikt/ghep/internal/slack"
)

type client struct {
	slack slack.Client
	teams []github.Team
}

func New(teams []github.Team, slack slack.Client) client {
	return client{
		slack: slack,
		teams: teams,
	}
}

func (c client) Run() error {
	slog.Info("Starting server")
	http.HandleFunc("/events", c.events)
	return http.ListenAndServe("127.0.0.1:8080", nil)
}