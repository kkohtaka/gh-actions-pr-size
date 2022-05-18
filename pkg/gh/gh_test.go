package gh_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/jarcoal/httpmock"
	"github.com/kkohtaka/gh-actions-pr-size/pkg/gh"
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
					page := getPage(req)
					if page-1 >= len(tt.commitFiles) {
						return nil, fmt.Errorf("invalid query value")
					}
					resp, err := httpmock.NewJsonResponse(200, tt.commitFiles[page-1])
					if err != nil {
						return nil, err
					}
					resp.Header.Set("Content-Type", "application/json")
					resp.Header.Set("Link", generateLinkHeaderValue(baseURL, page, len(tt.commitFiles)))
					return resp, nil
				},
			)

			got, err := gh.GetPullRequestChangedLines(
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

	got, err := gh.GetPullRequestChangedLines(
		context.Background(),
		github.NewClient(client),
		"kkohtaka",
		"gh-actions-pr-size",
		42,
	)
	assert.ErrorContains(t, err, "get all commit files: list commit files: ")
	assert.Zero(t, got)
}

func TestSetLabelOnPullRequest(t *testing.T) {
	tcs := []struct {
		name   string
		labels [][]*github.Label
		size   gh.Size

		wantDeletedLabels []string
		wantCreatedLabels []string
	}{
		{
			name: "The pull request doesn't have labels.",
			labels: [][]*github.Label{
				{},
			},
			size: gh.SizeXL,
			wantCreatedLabels: []string{
				gh.SizeXL.GetLabel(),
			},
		},
		{
			name: "The pull request already has another size label.",
			labels: [][]*github.Label{
				{
					{
						Name: github.String(gh.SizeL.GetLabel()),
					},
				},
			},
			size: gh.SizeXL,
			wantDeletedLabels: []string{
				gh.SizeL.GetLabel(),
			},
			wantCreatedLabels: []string{
				gh.SizeXL.GetLabel(),
			},
		},
		{
			name: "The pull request already has the target size label.",
			labels: [][]*github.Label{
				{
					{
						Name: github.String(gh.SizeXL.GetLabel()),
					},
				},
			},
			size: gh.SizeXL,
		},
		{
			name: "The pull request has another size label and non-size labels.",
			labels: [][]*github.Label{
				{
					{
						Name: github.String("foo"),
					},
				},
				{
					{
						Name: github.String("bar"),
					},
					{
						Name: github.String(gh.SizeS.GetLabel()),
					},
					{
						Name: github.String("baz"),
					},
				},
				{
					{
						Name: github.String("qux"),
					},
				},
			},
			size: gh.SizeM,
			wantDeletedLabels: []string{
				gh.SizeS.GetLabel(),
			},
			wantCreatedLabels: []string{
				gh.SizeM.GetLabel(),
			},
		},
		{
			name: "The pull request has multiple size labels.",
			labels: [][]*github.Label{
				{
					{
						Name: github.String(gh.SizeL.GetLabel()),
					},
					{
						Name: github.String(gh.SizeM.GetLabel()),
					},
				},
			},
			size: gh.SizeXL,
			wantDeletedLabels: []string{
				gh.SizeL.GetLabel(),
				gh.SizeM.GetLabel(),
			},
			wantCreatedLabels: []string{
				gh.SizeXL.GetLabel(),
			},
		},
	}
	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			httpmock.ActivateNonDefault(client)
			defer httpmock.DeactivateAndReset()

			var (
				gotDeletedLabels []string
				gotCreatedLabels []string
			)
			const baseURL = "https://api.github.com/repos/kkohtaka/gh-actions-pr-size/issues/42/labels"
			httpmock.RegisterResponder(
				"GET",
				baseURL,
				func(req *http.Request) (*http.Response, error) {
					page := getPage(req)
					if page-1 >= len(tt.labels) {
						return nil, fmt.Errorf("invalid query value")
					}
					resp, err := httpmock.NewJsonResponse(200, tt.labels[page-1])
					if err != nil {
						return nil, err
					}
					resp.Header.Set("Content-Type", "application/json")
					resp.Header.Set("Link", generateLinkHeaderValue(baseURL, page, len(tt.labels)))
					return resp, nil
				},
			)
			httpmock.RegisterResponder(
				"DELETE",
				fmt.Sprintf("=~^%s/(.*)$", baseURL),
				func(req *http.Request) (*http.Response, error) {
					label, err := httpmock.GetSubmatch(req, 1)
					if err != nil {
						return httpmock.NewStringResponse(400, "unable to get a label name"), nil
					}
					gotDeletedLabels = append(gotDeletedLabels, label)
					resp := httpmock.NewBytesResponse(200, nil)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				},
			)
			httpmock.RegisterResponder(
				"POST",
				baseURL,
				func(req *http.Request) (*http.Response, error) {
					var labels []string
					if err := json.NewDecoder(req.Body).Decode(&labels); err != nil {
						return httpmock.NewStringResponse(
							400,
							fmt.Sprintf("unable to decode request body: %v", err),
						), nil
					}
					gotCreatedLabels = append(gotCreatedLabels, labels...)
					resp := httpmock.NewBytesResponse(200, nil)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				},
			)

			err := gh.SetLabelOnPullRequest(
				context.Background(),
				github.NewClient(client),
				"kkohtaka",
				"gh-actions-pr-size",
				42,
				tt.size,
			)
			require.NoError(t, err)

			assert.Equal(t, tt.wantDeletedLabels, gotDeletedLabels)
			assert.Equal(t, tt.wantCreatedLabels, gotCreatedLabels)
		})
	}
}

func TestSetLabelOnPullRequestReturnsError(t *testing.T) {
	const baseURL = "https://api.github.com/repos/kkohtaka/gh-actions-pr-size/issues/42/labels"

	t.Run(
		"GitHub API that listing issue labels returns an error.",
		func(t *testing.T) {
			client := &http.Client{}
			httpmock.ActivateNonDefault(client)
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(
				"GET",
				baseURL,
				httpmock.NewErrorResponder(fmt.Errorf("test for error handling")),
			)

			err := gh.SetLabelOnPullRequest(
				context.Background(),
				github.NewClient(client),
				"kkohtaka",
				"gh-actions-pr-size",
				42,
				gh.SizeXL,
			)
			assert.ErrorContains(t, err, "list labels by issue: ")
		},
	)

	t.Run(
		"GitHub API that deleting an issue label returns an error.",
		func(t *testing.T) {
			client := &http.Client{}
			httpmock.ActivateNonDefault(client)
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(
				"GET",
				baseURL,
				func(req *http.Request) (*http.Response, error) {
					resp, err := httpmock.NewJsonResponse(200, []*github.Label{
						{
							Name: github.String(gh.SizeL.GetLabel()),
						},
					})
					if err != nil {
						return nil, err
					}
					resp.Header.Set("Content-Type", "application/json")
					resp.Header.Set("Link", generateLinkHeaderValue(baseURL, 1, 1))
					return resp, nil
				},
			)
			httpmock.RegisterResponder(
				"DELETE",
				fmt.Sprintf("%s/%s", baseURL, gh.SizeL.GetLabel()),
				httpmock.NewErrorResponder(fmt.Errorf("test for error handling")),
			)

			err := gh.SetLabelOnPullRequest(
				context.Background(),
				github.NewClient(client),
				"kkohtaka",
				"gh-actions-pr-size",
				42,
				gh.SizeXL,
			)
			assert.ErrorContains(t, err, "remove a label from a pull request: ")
		},
	)

	t.Run(
		"GitHub API that creating an issue label returns an error.",
		func(t *testing.T) {
			client := &http.Client{}
			httpmock.ActivateNonDefault(client)
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(
				"GET",
				baseURL,
				func(req *http.Request) (*http.Response, error) {
					resp, err := httpmock.NewJsonResponse(200, []*github.Label{
						{
							Name: github.String(gh.SizeL.GetLabel()),
						},
					})
					if err != nil {
						return nil, err
					}
					resp.Header.Set("Content-Type", "application/json")
					resp.Header.Set("Link", generateLinkHeaderValue(baseURL, 1, 1))
					return resp, nil
				},
			)
			httpmock.RegisterResponder(
				"DELETE",
				fmt.Sprintf("%s/%s", baseURL, gh.SizeL.GetLabel()),
				httpmock.NewBytesResponder(200, nil),
			)
			httpmock.RegisterResponder(
				"POST",
				baseURL,
				httpmock.NewErrorResponder(fmt.Errorf("test for error handling")),
			)

			err := gh.SetLabelOnPullRequest(
				context.Background(),
				github.NewClient(client),
				"kkohtaka",
				"gh-actions-pr-size",
				42,
				gh.SizeXL,
			)
			assert.ErrorContains(t, err, "add a label to a pull request: ")
		},
	)
}

func getPage(req *http.Request) int {
	page := 1
	if v, ok := req.URL.Query()["page"]; ok && len(v) > 0 {
		if v, err := strconv.Atoi(v[0]); err == nil {
			page = v
		}
	}
	return page
}

func generateLinkHeaderValue(baseURL string, page, amount int) string {
	var links []string
	links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="first"`, baseURL, 1))
	if page-1 >= 1 {
		links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="prev"`, baseURL, page-1))
	}
	if page+1 <= amount {
		links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="next"`, baseURL, page+1))
	}
	links = append(links, fmt.Sprintf(`<%s?page=%d>; rel="last"`, baseURL, amount))
	return strings.Join(links, ", ")
}
