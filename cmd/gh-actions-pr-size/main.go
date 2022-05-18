package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/go-github/v29/github"
	"github.com/spf13/cobra"

	"golang.org/x/oauth2"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var rootCmd = &cobra.Command{
	Use:   "pr-size",
	Short: "pr-size is a GitHub action for labeling Pull Requests's size",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootRun(cmd.Context())
	},
}

func main() {
	logger := zap.New()
	ctx := log.IntoContext(context.Background(), zap.New())
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Error(err, "Could not process the command.")
	}
}

func rootRun(ctx context.Context) error {
	logger := log.FromContext(ctx)

	if eventType := os.Getenv("GITHUB_EVENT_NAME"); eventType != "pull_request" {
		return fmt.Errorf(
			"unsupported event type %q is specified: event types other than \"pull_request\" is not supported",
			eventType,
		)
	}

	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return errors.New("mandatory environment variable GITHUB_EVENT_PATH is not specified")
	}

	payload, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return fmt.Errorf("unable to read an event file at %q: %w", eventPath, err)
	}

	var event github.PullRequestEvent
	err = json.Unmarshal(payload, &event)
	if err != nil {
		return fmt.Errorf("unable to unmarshal an event payload to JSON: %w", err)
	}

	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	number := event.GetPullRequest().GetNumber()

	logger.Info("Successfully read an event payload",
		"owner", owner,
		"repo", repo,
		"number", number,
	)

	var tc *http.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}
	client := github.NewClient(tc)

	changed, err := getPullRequestChangedLines(ctx, client, owner, repo, number)
	if err != nil {
		return fmt.Errorf("unable to get the number of changed lines in a pull request: %w", err)
	}

	size := newSize(changed)
	logger.Info("Got a size of a pull request", "size", size.String())

	err = setLabelOnPullRequest(ctx, client, owner, repo, number, size)
	if err != nil {
		return fmt.Errorf("unable to set a label on a pull request: %w", err)
	}
	logger.Info("Set a label to represent a pull request size", "size", size.String())
	return nil
}
