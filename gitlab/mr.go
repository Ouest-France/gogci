package gitlab

import (
	"os"
	"strconv"

	"github.com/xanzy/go-gitlab"
)

func (c *Client) CheckOldestMergeRequest() (bool, error) {

	// Init gitlab client
	git := gitlab.NewClient(nil, c.Token)
	err := git.SetBaseURL(c.URL)
	if err != nil {
		return false, err
	}

	// Get project and merge request IDs from Gitlab CI env vars
	projectID, err := strconv.Atoi(os.Getenv("CI_PROJECT_ID"))
	if err != nil {
		return false, err
	}
	mrIID, err := strconv.Atoi(os.Getenv("CI_MERGE_REQUEST_IID"))
	if err != nil {
		return false, err
	}

	// Get project open merge requests
	mrs, _, err := git.MergeRequests.ListProjectMergeRequests(projectID, &gitlab.ListProjectMergeRequestsOptions{State: gitlab.String("opened")})
	if err != nil {
		return false, err
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
