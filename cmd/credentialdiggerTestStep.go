package cmd

import (
	"context"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type credentialdiggerTestStepService interface {
	// TODO: get owner and repo from repository string
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

func credentialdiggerTestStep(config credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) {
	ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, "", []string{})
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}
	err = runCredentialdiggerTestStep(ctx, &config, telemetryData, client.Issues)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to comment on issue")
	}
}

func runCredentialdiggerTestStep(ctx context.Context, config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData, service credentialdiggerTestStepService) error {

	issueComment := github.IssueComment{
		Body: &config.Body,
	}

	// TODO get Repo name and Owner from repo url
	s := strings.Split(config.Repository, "/")
	owner := s[len(s)-2]
	repoName := s[len(s)-1]
	newcomment, resp, err := service.CreateComment(ctx, owner, repoName, config.Number, &issueComment)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return errors.Wrapf(err, "Error occurred when creating comment on issue %v", config.Number)
	}
	log.Entry().Debugf("New issue comment created for issue %v: %v", config.Number, newcomment)

	return nil
}
