package gitlab

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/xanzy/go-gitlab"
)

func (c *Client) CreateMergeRequestNote(tmpl string, data interface{}) error {

	// Init gitlab client
	git := gitlab.NewClient(nil, c.Token)
	err := git.SetBaseURL(c.URL)
	if err != nil {
		return err
	}

	// Init template
	t, err := template.New("t").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %s", err)
	}

	// Process template
	var comment bytes.Buffer
	err = t.Execute(&comment, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %s", err)
	}
	body := comment.String()

	// Set processed template as body
	n := &gitlab.CreateMergeRequestNoteOptions{
		Body: &body,
	}

	// Get project and merge request IDs from Gitlab CI env vars
	projectID, err := strconv.Atoi(os.Getenv("CI_PROJECT_ID"))
	if err != nil {
		return err
	}
	mrID, err := strconv.Atoi(os.Getenv("CI_MERGE_REQUEST_IID"))
	if err != nil {
		return err
	}

	// Create comment on MR
	_, _, err = git.Notes.CreateMergeRequestNote(projectID, mrID, n)

	return err
}

func (c *Client) TerraformInitFailed() error {

	var notif = " :red_circle: Terraform init **failed** in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}

func (c *Client) TerraformPlanRunning() error {

	var notif = "Terraform plan running in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}

func (c *Client) TerraformPlanFailed(output string) error {

	var notif = " :red_circle: Terraform plan **failed** in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL, Stdout string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
		Stdout:      output,
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}

func (c *Client) TerraformPlanSummary(output string) error {

	var notif = "Terraform plan ran in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

**Plan summary**: {{.Summary}}

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Extract summary
	nochanges, err := regexp.MatchString("No changes. Infrastructure is up-to-date.", output)
	if err != nil {
		return fmt.Errorf("failed to match string: %s", err)
	}

	summary := "No changes. Infrastructure is up-to-date."
	if !nochanges {
		r, err := regexp.Compile("([0-9]+) to add, ([0-9]+) to change, ([0-9]+) to destroy")
		if err != nil {
			return fmt.Errorf("failed to compile regex: %s", err)
		}
		summary = r.FindString(output)
	}

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL, Summary, Stdout string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
		Summary:     summary,
		Stdout:      output,
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}

func (c *Client) TerraformApplyRunning() error {

	var notif = "Terraform apply running in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}

func (c *Client) TerraformApplyFailed(output string) error {

	var notif = " :red_circle: Terraform apply **failed** in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL, Stdout string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
		Stdout:      output,
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}

func (c *Client) TerraformApplySummary(output string) error {

	var notif = "Terraform apply ran in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

**Apply summary**: {{.Summary}}

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Extract summary
	r, err := regexp.Compile("([0-9]+) added, ([0-9]+) changed, ([0-9]+) destroyed")
	if err != nil {
		return fmt.Errorf("failed to compile regex: %s", err)
	}
	summary := r.FindString(output)

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL, Summary, Stdout string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
		Summary:     summary,
		Stdout:      output,
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}

func (c *Client) TerraformApplyNotApproved() error {

	var notif = ":no_entry: Terraform apply **not authorized**, merge request must be approved, in dir `{{.Dir}}` for commit `{{.Commit}}` in pipeline `{{.PipelineID}}`." + `

:memo: [see job log]({{.Job}}) | :arrow_forward: [see pipeline]({{.PipelineURL}})`

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Extract subdir in path
	wd = strings.Replace(wd, os.Getenv("CI_PROJECT_DIR"), "", 1)
	if wd == "" {
		wd = "."
	}

	// Collect data for templating
	data := struct {
		Dir, Commit, Job, PipelineID, PipelineURL, Stdout string
	}{
		Dir:         wd,
		Commit:      os.Getenv("CI_COMMIT_SHORT_SHA"),
		Job:         os.Getenv("CI_JOB_URL"),
		PipelineID:  os.Getenv("CI_PIPELINE_ID"),
		PipelineURL: os.Getenv("CI_PIPELINE_URL"),
	}

	// Create comment
	err = c.CreateMergeRequestNote(notif, data)

	return err
}
