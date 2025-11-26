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
	Branch    string // "release" or "master"
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
	masterCommits  []*gitlab.Commit
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
	// 0. Pre-load commits for both branches for squash merge detection
	if err := c.loadReleaseCommits(since); err != nil {
		return nil, err
	}
	if err := c.loadMasterCommits(since); err != nil {
		return nil, err
	}

	var allResults []MRResult

	// 1. Check release branch
	releaseResults, err := c.checkBranch("release", since)
	if err != nil {
		return nil, err
	}
	allResults = append(allResults, releaseResults...)

	// 2. Check master branch
	masterResults, err := c.checkBranch("master", since)
	if err != nil {
		return nil, err
	}
	allResults = append(allResults, masterResults...)

	return allResults, nil
}

// checkBranch checks for ghost MRs in a specific branch
func (c *Checker) checkBranch(branchName string, since time.Time) ([]MRResult, error) {
	var results []MRResult
	stateMerged := "merged"
	scope := "all"

	opt := &gitlab.ListProjectMergeRequestsOptions{
		TargetBranch: &branchName,
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

			// Check if code is actually in the branch (Ghost detection)
			isMerged := c.isMergedToBranch(mr, branchName)

			if !isMerged {
				// It's a Ghost! (Merged in GitLab, but not in branch)
				results = append(results, MRResult{
					ReleaseMR: mr,
					Status:    StatusGhost,
					Branch:    branchName,
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return results, nil
}

func (c *Checker) loadReleaseCommits(since time.Time) error {
	return c.loadCommitsForBranch("release", since, &c.releaseCommits)
}

func (c *Checker) loadMasterCommits(since time.Time) error {
	return c.loadCommitsForBranch("master", since, &c.masterCommits)
}

func (c *Checker) loadCommitsForBranch(branchName string, since time.Time, commits *[]*gitlab.Commit) error {
	opt := &gitlab.ListCommitsOptions{
		RefName: &branchName,
		Since:   &since,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	for {
		branchCommits, resp, err := c.Commits.ListCommits(c.ProjectID, opt)
		if err != nil {
			return err
		}
		*commits = append(*commits, branchCommits...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil
}

func (c *Checker) isMergedToBranch(mr *gitlab.MergeRequest, branchName string) bool {
	var commits []*gitlab.Commit

	// Select the appropriate commit cache based on branch
	switch branchName {
	case "release":
		commits = c.releaseCommits
	case "master":
		commits = c.masterCommits
	default:
		return false
	}

	// 1. Check SHA (Fast path for non-squash merges)
	for _, commit := range commits {
		if commit.ID == mr.SHA {
			return true
		}
	}

	// 2. Check Title (Slow path for squash merges)
	// GitLab default squash message usually contains the title
	for _, commit := range commits {
		if containsTitle(commit.Title, mr.Title) || containsTitle(commit.Message, mr.Title) {
			return true
		}
	}

	return false
}

func containsTitle(text, title string) bool {
	return strings.Contains(text, title)
}
