package checker

import (
	"testing"
	"time"

	"github.com/xanzy/go-gitlab"
)

type MockGitlabClient struct {
	ListProjectMergeRequestsFunc func(pid interface{}, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequest, *gitlab.Response, error)
	CompareFunc                  func(pid interface{}, opt *gitlab.CompareOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Compare, *gitlab.Response, error)
	ListCommitsFunc              func(pid interface{}, opt *gitlab.ListCommitsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Commit, *gitlab.Response, error)
}

func (m *MockGitlabClient) ListProjectMergeRequests(pid interface{}, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequest, *gitlab.Response, error) {
	return m.ListProjectMergeRequestsFunc(pid, opt, options...)
}

func (m *MockGitlabClient) Compare(pid interface{}, opt *gitlab.CompareOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Compare, *gitlab.Response, error) {
	if m.CompareFunc != nil {
		return m.CompareFunc(pid, opt, options...)
	}
	return &gitlab.Compare{Diffs: []*gitlab.Diff{}}, &gitlab.Response{}, nil
}

func (m *MockGitlabClient) ListCommits(pid interface{}, opt *gitlab.ListCommitsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Commit, *gitlab.Response, error) {
	if m.ListCommitsFunc != nil {
		return m.ListCommitsFunc(pid, opt, options...)
	}
	return []*gitlab.Commit{}, &gitlab.Response{}, nil
}

func TestCheck(t *testing.T) {
	since := time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC)
	mergedAt := time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC)

	mockClient := &MockGitlabClient{
		ListProjectMergeRequestsFunc: func(pid interface{}, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequest, *gitlab.Response, error) {
			// Mock response for fetching release MRs
			if opt.TargetBranch != nil && *opt.TargetBranch == "release" {
				return []*gitlab.MergeRequest{
					{
						IID:          1,
						Title:        "Feature A",
						SourceBranch: "feature/a",
						MergedAt:     &mergedAt,
						SHA:          "sha1",
					},
					{
						IID:          2,
						Title:        "Feature B",
						SourceBranch: "feature/b",
						MergedAt:     &mergedAt,
						SHA:          "sha2",
					},
				}, &gitlab.Response{NextPage: 0}, nil
			}

			return nil, nil, nil
		},
		ListCommitsFunc: func(pid interface{}, opt *gitlab.ListCommitsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Commit, *gitlab.Response, error) {
			// Mock commits in release branch
			return []*gitlab.Commit{
				{ID: "sha1", Title: "Feature A", Message: "Feature A"},
				{ID: "commit2", Title: "Some squashed commit", Message: "Feature B\n\nSquashed from feature/b"},
			}, &gitlab.Response{NextPage: 0}, nil
		},
		CompareFunc: func(pid interface{}, opt *gitlab.CompareOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Compare, *gitlab.Response, error) {
			// Logic: Compare(from=release, to=sha)
			// If diffs empty -> Merged (Real MR) -> Skip
			// If diffs exist -> Not Merged (Ghost MR) -> Report

			if opt.To != nil && *opt.To == "sha1" {
				// Feature A: Real MR (Merged to release)
				// Expect empty diffs
				return &gitlab.Compare{
					Diffs: []*gitlab.Diff{},
				}, &gitlab.Response{}, nil
			}
			if opt.To != nil && *opt.To == "sha2" {
				// Feature B: Ghost MR (Not actually in release)
				// Expect diffs
				return &gitlab.Compare{
					Diffs: []*gitlab.Diff{{NewPath: "ghost.txt"}},
				}, &gitlab.Response{}, nil
			}
			return nil, nil, nil
		},
	}

	c := NewChecker(mockClient, mockClient, "123")
	results, err := c.Check(since)

	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	// Both MRs should be found in release (sha1 by SHA, sha2 by title in message)
	// So we expect 0 ghosts
	if len(results) != 0 {
		t.Fatalf("Expected 0 results (no ghosts), got %d", len(results))
	}
}
