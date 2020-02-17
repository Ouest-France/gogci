package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type Gitlab struct {
	Token string
	URL   string
}

func (g *Gitlab) CreateMergeRequestNote(tmpl string, data interface{}) error {

	// Init gitlab client
	git := gitlab.NewClient(nil, g.Token)
	err := git.SetBaseURL(g.URL)
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

func (g *Gitlab) TerraformPlanRunning() error {

	var notif = "Terraform plan launched in `{{.Dir}}` for commit `{{.Commit}}` in pipeline [{{.PipelineID}}]({{.PipelineURL}}): [see job log]({{.Job}})"

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
	err = g.CreateMergeRequestNote(notif, data)

	return err
}

func (g *Gitlab) TerraformPlanSummary() error {

	var notif = "Terraform plan ran in `{{.Dir}}` for commit `{{.Commit}}` in pipeline [{{.PipelineID}}]({{.PipelineURL}}): [see job log]({{.Job}})"

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
	err = g.CreateMergeRequestNote(notif, data)

	return err
}
