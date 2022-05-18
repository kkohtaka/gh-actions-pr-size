package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestPRSize(t *testing.T) {
	var setup = func(t *testing.T) {
		t.Setenv("GITHUB_EVENT_NAME", "pull_request")

		f, err := os.CreateTemp("", "event-*.json")
		require.NoError(t, err)
		t.Cleanup(func() {
			os.RemoveAll(f.Name())
		})
		event := &github.PullRequestEvent{
			Repo: &github.Repository{
				Owner: &github.User{
					Login: github.String("kkohtaka"),
				},
				Name: github.String("gh-actions-pr-size"),
			},
			PullRequest: &github.PullRequest{
				Number: github.Int(42),
			},
		}
		data, err := json.Marshal(event)
		require.NoError(t, err)
		_, err = f.Write(data)
		require.NoError(t, err)
		t.Setenv("GITHUB_EVENT_PATH", f.Name())

		httpmock.Activate()
		t.Cleanup(func() {
			httpmock.DeactivateAndReset()
		})
		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/kkohtaka/gh-actions-pr-size/pulls/42/files",
			func(_ *http.Request) (*http.Response, error) {
				return httpmock.NewJsonResponse(200, []github.CommitFile{
					{
						Additions: github.Int(100),
						Deletions: github.Int(200),
					},
				})
			},
		)

		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/kkohtaka/gh-actions-pr-size/issues/42/labels",
			func(_ *http.Request) (*http.Response, error) {
				return httpmock.NewJsonResponse(200, []github.Label{})
			},
		)

		var gotCreatedLabels []string
		httpmock.RegisterResponder(
			"POST",
			"https://api.github.com/repos/kkohtaka/gh-actions-pr-size/issues/42/labels",
			func(req *http.Request) (*http.Response, error) {
				var labels []string
				if err := json.NewDecoder(req.Body).Decode(&labels); err != nil {
					return httpmock.NewStringResponse(
						400,
						fmt.Sprintf("unable to decode request body: %v", err),
					), nil
				}
				gotCreatedLabels = append(gotCreatedLabels, labels...)
				return httpmock.NewBytesResponse(200, nil), nil
			},
		)
	}

	t.Run("A normal condition", func(t *testing.T) {
		setup(t)
		err := PRSizeCmd.ExecuteContext(context.Background())
		require.NoError(t, err)
	})

	t.Run("An unsupported event type is specified.", func(t *testing.T) {
		setup(t)
		t.Setenv("GITHUB_EVENT_NAME", "push")
		err := PRSizeCmd.ExecuteContext(context.Background())
		require.ErrorContains(
			t,
			err,
			"unsupported event type \"push\" is specified: event types other than \"pull_request\" is not supported",
		)
	})

	t.Run("A path to event payload JSON file is not specified.", func(t *testing.T) {
		setup(t)
		t.Setenv("GITHUB_EVENT_PATH", "")
		err := PRSizeCmd.ExecuteContext(context.Background())
		require.ErrorContains(
			t,
			err,
			"mandatory environment variable GITHUB_EVENT_PATH is not specified",
		)
	})

	t.Run("An event payload JSON file doesn't exist.", func(t *testing.T) {
		setup(t)
		t.Setenv("GITHUB_EVENT_PATH", "not-exist.json")
		err := PRSizeCmd.ExecuteContext(context.Background())
		require.ErrorContains(
			t,
			err,
			"unable to read an event file at \"not-exist.json\"",
		)
	})

	t.Run("An event payload file cannot be decoded as JSON.", func(t *testing.T) {
		setup(t)
		f, err := os.CreateTemp("", "event-*.txt")
		require.NoError(t, err)
		t.Cleanup(func() {
			os.RemoveAll(f.Name())
		})
		require.NoError(t, err)
		_, err = f.Write([]byte("This is not a JSON file."))
		require.NoError(t, err)
		t.Setenv("GITHUB_EVENT_PATH", f.Name())
		err = PRSizeCmd.ExecuteContext(context.Background())
		require.ErrorContains(
			t,
			err,
			"unable to unmarshal an event payload to JSON",
		)
	})

	t.Run("GitHub Pull Request Files API returns an error.", func(t *testing.T) {
		setup(t)
		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/kkohtaka/gh-actions-pr-size/pulls/42/files",
			httpmock.NewErrorResponder(errors.New("unable to process issue labels API")),
		)
		err := PRSizeCmd.ExecuteContext(context.Background())
		require.ErrorContains(
			t,
			err,
			"unable to get the number of changed lines in a pull request",
		)
	})

	t.Run("GitHub Issue Labels API returns an error.", func(t *testing.T) {
		setup(t)
		httpmock.RegisterResponder(
			"POST",
			"https://api.github.com/repos/kkohtaka/gh-actions-pr-size/issues/42/labels",
			httpmock.NewErrorResponder(errors.New("unable to process issue labels API")),
		)
		err := PRSizeCmd.ExecuteContext(context.Background())
		require.ErrorContains(
			t,
			err,
			"unable to set a label on a pull request",
		)
	})
}
