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
	//CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)

	ListByOrg(ctx context.Context, owner string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
}

func credentialdiggerTestStep(config credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) {
	ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, "", []string{})
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}
	//err = runCredentialdiggerTestStep(ctx, &config, telemetryData, client.Issues)
	err = runCredentialdiggerTestStep(ctx, &config, telemetryData, client.Repositories)
	//err = runGHList(ctx, &config, telemetryData, client.Repositories)
	if err != nil {
		//log.Entry().WithError(err).Fatal("Failed to comment on issue")
		log.Entry().WithError(err).Fatal("Failed to run custom function")
	}
}

//func runGHList(ctx context.Context, config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData, service credentialdiggerTestStepService) error {
//	s := strings.Split(config.Repository, "/")
//	owner := s[len(s)-2]
//	// repoName := s[len(s)-1]
//	repos, resp, err := service.ListByOrg(ctx, owner)
//	//newcomment, resp, err := service.CreateComment(ctx, owner, repoName, config.Number, &issueComment)
//
//	return nil
//}

func runCredentialdiggerTestStep(ctx context.Context, config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData, service credentialdiggerTestStepService) error {

	//issueComment := github.IssueComment{
	//	Body: &config.Body,
	//}

	// TODO get Repo name and Owner from repo url
	s := strings.Split(config.Repository, "/")
	owner := s[len(s)-2]
	//repoName := s[len(s)-1]

	// Create new comment
	//newcomment, resp, err := service.CreateComment(ctx, owner, repoName, config.Number, &issueComment)
	//if err != nil {
	//	log.Entry().Errorf("GitHub response code %v", resp.Status)
	//	return errors.Wrapf(err, "Error occurred when creating comment on issue %v", config.Number)
	//}
	//log.Entry().Debugf("New issue comment created for issue %v: %v", config.Number, newcomment)

	// List reposIssues
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, resp, err := service.ListByOrg(ctx, owner, opt)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return errors.Wrapf(err, "Error occurred when listing repos of owner %v", owner)
	}
	log.Entry().Debugf("Repos listed for owner %v", owner)
	log.Entry().Info(repos)

	for i, r := range repos {
		log.Entry().Info(i, r)
	}

	return nil
}
