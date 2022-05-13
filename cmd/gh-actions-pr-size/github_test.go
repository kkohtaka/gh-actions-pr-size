package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPullRequestChangedLines(t *testing.T) {
	tcs := []struct {
		name        string
		commitFiles [][]*github.CommitFile
		want        int
	}{
		{
			name: "The target pull request changes a single file.",
			commitFiles: [][]*github.CommitFile{
				{
					{
						Additions: github.Int(100),
						Deletions: github.Int(200),
					},
				},
			},
			want: 300,
		},
		{
			name: "The target pull request changes multiple files.",
			commitFiles: [][]*github.CommitFile{
				{
					{
						Additions: github.Int(100),
						Deletions: github.Int(200),
					},
					{
						Additions: github.Int(300),
						Deletions: github.Int(400),
					},
				},
			},
			want: 1000,
		},
		{
			name: "The target pull request changes multiple files with pagenation.",
			commitFiles: [][]*github.CommitFile{
				{
					{
						Additions: github.Int(100),
						Deletions: github.Int(200),
					},
				},
				{
					{
						Additions: github.Int(300),
						Deletions: github.Int(400),
					},
				},
				{
					{
						Additions: github.Int(500),
						Deletions: github.Int(600),
					},
				},
			},
			want: 2100,
		},
	}
	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			httpmock.ActivateNonDefault(client)
			defer httpmock.DeactivateAndReset()

			const baseURL = "https://api.github.com/repos/kkohtaka/gh-actions-pr-size/pulls/42/files"
			httpmock.RegisterResponder(
				"GET",
				baseURL,
				func(req *http.Request) (*http.Response, error) {
					page := 1
					if v, ok := req.URL.Query()["page"]; ok && len(v) > 0 {
						if v, err := strconv.Atoi(v[0]); err == nil {
							page = v
						}
					}
					if page-1 >= len(tt.commitFiles) {
						return nil, fmt.Errorf("invalid query value")
					}

					var links []string
					links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="first"`, baseURL, 1))
					if page-1 >= 1 {
						links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="prev"`, baseURL, page-1))
					}
					if page+1 <= len(tt.commitFiles) {
						links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="next"`, baseURL, page+1))
					}
					links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="last"`, baseURL, len(tt.commitFiles)))

					resp, err := httpmock.NewJsonResponse(200, tt.commitFiles[page-1])
					if err != nil {
						return nil, err
					}
					resp.Header.Set("Content-Type", "application/json")
					resp.Header.Set("Link", strings.Join(links, ", "))
					return resp, nil
				},
			)

			got, err := getPullRequestChangedLines(
				context.Background(),
				github.NewClient(client),
				"kkohtaka",
				"gh-actions-pr-size",
				42,
			)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPullRequestChangedLinesReturnsError(t *testing.T) {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	const baseURL = "https://api.github.com/repos/kkohtaka/gh-actions-pr-size/pulls/42/files"
	httpmock.RegisterResponder(
		"GET",
		baseURL,
		httpmock.NewErrorResponder(fmt.Errorf("test for error handling")),
	)

	got, err := getPullRequestChangedLines(
		context.Background(),
		github.NewClient(client),
		"kkohtaka",
		"gh-actions-pr-size",
		42,
	)
	assert.ErrorContains(t, err, "get all commit files: list commit files: ")
	assert.Zero(t, got)
}
