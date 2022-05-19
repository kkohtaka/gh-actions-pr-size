package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestPRSize(t *testing.T) {
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

	err = PRSizeCmd.ExecuteContext(context.Background())
	require.NoError(t, err)
}
