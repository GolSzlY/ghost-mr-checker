package checker

import (
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
)

// MRStatus represents the status of an MR in relation to the master branch
type MRStatus string

const (
	StatusMissing MRStatus = "MISSING"
	StatusOpen    MRStatus = "OPEN"
	StatusMerged  MRStatus = "MERGED"
	StatusClosed  MRStatus = "CLOSED"
	StatusGhost   MRStatus = "GHOST"
)

// MRResult holds the result of checking an MR
type MRResult struct {
	ReleaseMR *gitlab.MergeRequest
	MasterMR  *gitlab.MergeRequest
	Status    MRStatus
}

// GitlabClientInterface defines the subset of GitLab client methods we need
// This helps with mocking in tests
type MergeRequestsService interface {
	ListProjectMergeRequests(pid interface{}, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequest, *gitlab.Response, error)
}

type CommitsService interface {
	ListCommits(pid interface{}, opt *gitlab.ListCommitsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Commit, *gitlab.Response, error)
}

// Checker holds the dependencies for the check operation
type Checker struct {
	Commits        CommitsService
	MR             MergeRequestsService
	ProjectID      string
	releaseCommits []*gitlab.Commit
}

// NewChecker creates a new Checker instance
func NewChecker(commits CommitsService, mr MergeRequestsService, projectID string) *Checker {
	return &Checker{
		Commits:   commits,
		MR:        mr,
		ProjectID: projectID,
	}
}

// Check compares MRs between release and master branches
func (c *Checker) Check(since time.Time) ([]MRResult, error) {
	// 0. Pre-load release commits for squash merge detection
	if err := c.loadReleaseCommits(since); err != nil {
		return nil, err
	}

	var allResults []MRResult

	// 1. Fetch MRs merged into release since the date
	targetRelease := "release"
	stateMerged := "merged"
	scope := "all"

	opt := &gitlab.ListProjectMergeRequestsOptions{
		TargetBranch: &targetRelease,
		State:        &stateMerged,
		Scope:        &scope,
		UpdatedAfter: &since,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	for {
		mrs, resp, err := c.MR.ListProjectMergeRequests(c.ProjectID, opt)
		if err != nil {
			return nil, err
		}

		for _, mr := range mrs {
			// Double check merged_at because UpdatedAfter includes comments etc.
			if mr.MergedAt != nil && mr.MergedAt.Before(since) {
				continue
			}

			// Check if code is actually in release (Ghost detection)
			// We check both SHA (fast) and Title (slow/squash)
			isMerged := c.isMergedToBranch(mr)

			if !isMerged {
				// It's a Ghost! (Merged in GitLab, but not in branch)
				allResults = append(allResults, MRResult{
					ReleaseMR: mr,
					Status:    StatusGhost,
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allResults, nil
}

func (c *Checker) loadReleaseCommits(since time.Time) error {
	refName := "release"
	opt := &gitlab.ListCommitsOptions{
		RefName: &refName,
		Since:   &since,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	for {
		commits, resp, err := c.Commits.ListCommits(c.ProjectID, opt)
		if err != nil {
			return err
		}
		c.releaseCommits = append(c.releaseCommits, commits...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil
}

func (c *Checker) isMergedToBranch(mr *gitlab.MergeRequest) bool {
	// 1. Check SHA (Fast path for non-squash merges)
	// We can iterate our cached commits instead of calling API again
	for _, commit := range c.releaseCommits {
		if commit.ID == mr.SHA {
			return true
		}
	}

	// 2. Check Title (Slow path for squash merges)
	// GitLab default squash message usually contains the title
	for _, commit := range c.releaseCommits {
		if containsTitle(commit.Title, mr.Title) || containsTitle(commit.Message, mr.Title) {
			return true
		}
	}

	return false
}

func containsTitle(text, title string) bool {
	return strings.Contains(text, title)
}
