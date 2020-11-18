package gitlab

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/xanzy/go-gitlab"
)

func (c *Client) CheckOldestMergeRequest() (bool, error) {

	// Init gitlab client
	git := gitlab.NewClient(nil, c.Token)
	err := git.SetBaseURL(c.URL)
	if err != nil {
		return false, fmt.Errorf("failed to set gitlab client base url: %w", err)
	}

	// Get project and merge request IDs from Gitlab CI env vars
	if os.Getenv("CI_PROJECT_ID") == "" || os.Getenv("CI_MERGE_REQUEST_IID") == "" {
		return false, errors.New("CI_PROJECT_ID or CI_MERGE_REQUEST_IID env var is not defined, GOGCI must run in a Merge Request")
	}
	projectID, err := strconv.Atoi(os.Getenv("CI_PROJECT_ID"))
	if err != nil {
		return false, fmt.Errorf("failed to parse CI_PROJECT_ID env var: %w", err)
	}
	mrIID, err := strconv.Atoi(os.Getenv("CI_MERGE_REQUEST_IID"))
	if err != nil {
		return false, fmt.Errorf("failed to parse CI_MERGE_REQUEST_IID env var: %w", err)
	}

	// Get project open merge requests
	mrs, _, err := git.MergeRequests.ListProjectMergeRequests(projectID, &gitlab.ListProjectMergeRequestsOptions{State: gitlab.String("opened")})
	if err != nil {
		return false, fmt.Errorf("failed to list project merge requests: %w", err)
	}

	// Checks if any merge requests passed have an older iid
	for _, mr := range mrs {
		if mr.IID < mrIID {
			return false, nil
		}
	}

	// Return open merge requests
	return true, nil
}
