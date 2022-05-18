package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v29/github"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getAllPullRequestFiles returns all commit files in a pull request.
func getAllPullRequestFiles(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	number int,
) ([]*github.CommitFile, error) {
	logger := log.FromContext(ctx).WithValues(
		"owner", owner,
		"repo", repo,
		"number", number,
	)

	var res []*github.CommitFile
	for offset := 0; ; offset++ {
		logger = logger.WithValues("offset", offset)
		files, resp, err := client.PullRequests.ListFiles(
			ctx,
			owner, repo, number,
			&github.ListOptions{Page: offset + 1, PerPage: 100},
		)
		if err != nil {
			logger.Error(err, "Failed to list files changed by a pull request")
			return nil, fmt.Errorf("list commit files: %w", err)
		}
		res = append(res, files...)
		if offset+1 >= resp.LastPage {
			break
		}
	}
	return res, nil
}

// getPullRequestSize returns the total number of changed lines of the specified pull request.
func getPullRequestChangedLines(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	number int,
) (int, error) {
	files, err := getAllPullRequestFiles(ctx, client, owner, repo, number)
	if err != nil {
		return 0, fmt.Errorf("get all commit files: %w", err)
	}

	// TODO(kkohtaka): Filter out linguist-generated files

	change := 0
	for _, file := range files {
		change += *file.Additions + *file.Deletions
	}
	return change, nil
}

// setLabelOnPullRequest checks the current labels on the pull request.  If there exists a label for pull request size,
// the function replaces it with the proper label.  Otherwise, the function just attach the proper label.
func setLabelOnPullRequest(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	number int,
	size size,
) error {
	logger := log.FromContext(ctx).WithValues(
		"owner", owner,
		"repo", repo,
		"number", number,
		"label", size.getLabel(),
	)

	for offset := 0; ; offset++ {
		labels, resp, err := client.Issues.ListLabelsByIssue(
			ctx,
			owner, repo, number,
			&github.ListOptions{Page: offset + 1, PerPage: 100},
		)
		if err != nil {
			logger.Error(err, "Failed to list labels on a pull request")
			return fmt.Errorf("list labels by issue: %w", err)
		}
		for _, label := range labels {
			newLogger := logger.WithValues("remove", label.GetName())
			if strings.HasPrefix(label.GetName(), labelPrefix) {
				if label.GetName() == size.getLabel() {
					newLogger.Info("The pull request already has the label")
					return nil
				}
				// Remove the current label for pull request size
				if _, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label.GetName()); err != nil {
					newLogger.Error(err, "Failed to remove a label from a pull request")
					return fmt.Errorf("remove a label from a pull request: %w", err)
				}
				newLogger.Info("A label was removed from the pull request")
			}
		}
		if offset+1 >= resp.LastPage {
			break
		}
	}

	if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, number, []string{size.getLabel()}); err != nil {
		logger.Error(err, "Failed to add a label to a pull request")
		return fmt.Errorf("add a label to a pull request: %w", err)
	}
	return nil
}
