package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v29/github"

	"go.uber.org/zap"
)

// getAllPullRequestFiles returns all commit files in a pull request.
func getAllPullRequestFiles(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	number int,
) ([]*github.CommitFile, error) {
	var res []*github.CommitFile
	for offset := 0; ; offset++ {
		files, resp, err := client.PullRequests.ListFiles(
			ctx,
			owner, repo, number,
			&github.ListOptions{Page: offset + 1, PerPage: 100},
		)
		if err != nil {
			logger.Error("Failed to remove a label from a pull request",
				zap.String("owner", owner),
				zap.String("repo", repo),
				zap.Int("number", number),
				zap.Int("offset", offset),
				zap.Error(err),
			)
			return nil, fmt.Errorf("list commit files: %w", err)
		}
		res = append(res, files...)
		if offset+1 >= resp.LastPage {
			break
		}
	}
	return res, nil
}

// getPullRequestSize determines a size of a pull request by calculating a total number of changed lines.
func getPullRequestSize(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	number int,
) (size, error) {
	files, err := getAllPullRequestFiles(ctx, client, owner, repo, number)
	if err != nil {
		return sizeUnknown, fmt.Errorf("get all commit files: %w", err)
	}

	// TODO(kkohtaka): Filter out linguist-generated files

	change := 0
	for _, file := range files {
		change += *file.Additions + *file.Deletions
	}
	return newSize(change), nil
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
	for offset := 0; ; offset++ {
		labels, resp, err := client.Issues.ListLabelsByIssue(
			ctx,
			owner, repo, number,
			&github.ListOptions{Page: offset + 1, PerPage: 100},
		)
		if err != nil {
			logger.Error("Failed to list labels on a pull request",
				zap.String("owner", owner),
				zap.String("repo", repo),
				zap.Int("number", number),
				zap.Error(err),
			)
			return fmt.Errorf("list labels by issue: %w", err)
		}
		for _, label := range labels {
			if strings.HasPrefix(label.GetName(), labelPrefix) {
				if label.GetName() == size.getLabel() {
					logger.Debug("The pull request already has the label",
						zap.String("owner", owner),
						zap.String("repo", repo),
						zap.Int("number", number),
						zap.String("label", label.GetName()),
					)
					return nil
				}
				// Remove the current label for pull request size
				if _, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label.GetName()); err != nil {
					logger.Error("Failed to remove a label from a pull request",
						zap.String("owner", owner),
						zap.String("repo", repo),
						zap.Int("number", number),
						zap.String("label", label.GetName()),
						zap.Error(err),
					)
					return fmt.Errorf("remove a label from a pull request: %w", err)
				}
				logger.Info("A label was removed from the pull request",
					zap.String("owner", owner),
					zap.String("repo", repo),
					zap.Int("number", number),
					zap.String("label", label.GetName()),
				)
			}
		}
		if offset+1 >= resp.LastPage {
			break
		}
	}

	if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, number, []string{size.getLabel()}); err != nil {
		logger.Error("Failed to add a label to a pull request",
			zap.String("owner", owner),
			zap.String("repo", repo),
			zap.Int("number", number),
			zap.String("label", size.getLabel()),
			zap.Error(err),
		)
		return fmt.Errorf("add a label to a pull request: %w", err)
	}
	return nil
}
