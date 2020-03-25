package gitlab

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/xanzy/go-gitlab"
)

func (c *Client) CheckMergeRequestApproved() (bool, error) {

	// Init gitlab client
	git := gitlab.NewClient(nil, c.Token)
	err := git.SetBaseURL(c.URL)
	if err != nil {
		return false, fmt.Errorf("failed to set Gitlab client URL: %s", err)
	}

	// Get project and merge request IDs from Gitlab CI env vars
	if os.Getenv("CI_PROJECT_ID") == "" || os.Getenv("CI_MERGE_REQUEST_IID") == "" {
		return false, errors.New("CI_PROJECT_ID or CI_MERGE_REQUEST_IID env var is not defined, GOGCI must run in a Merge Request")
	}
	projectID, err := strconv.Atoi(os.Getenv("CI_PROJECT_ID"))
	if err != nil {
		return false, fmt.Errorf("failed to parse CI_PROJECT_ID env var: %s", err)
	}
	mrID, err := strconv.Atoi(os.Getenv("CI_MERGE_REQUEST_IID"))
	if err != nil {
		return false, fmt.Errorf("failed to parse CI_MERGE_REQUEST_IID env var: %s", err)
	}

	// Get merge request approval state
	approvalState, _, err := git.MergeRequestApprovals.GetApprovalState(projectID, mrID)
	if err != nil {
		return false, fmt.Errorf("failed to get merge request approval state: %s", err)
	}

	// Return true if all approval rules are ok
	return c.CheckApprovalRules(approvalState), nil
}

func (c *Client) CheckApprovalRules(approval *gitlab.MergeRequestApprovalState) bool {

	// For each rule check if number of approvals is less thant required
	for _, rule := range approval.Rules {
		if rule.ApprovalsRequired > len(rule.ApprovedBy) {
			return false
		}
	}

	return true
}
