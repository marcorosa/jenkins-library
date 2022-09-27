package cmd

import (
	"context"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
)

const piperTempDb string = "piper_step_db.db"
const reportTempName string = "findings.csv"

type credentialdiggerTestStepService interface {
	//CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)

	ListByOrg(ctx context.Context, owner string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
}

type credentialdiggerUtils interface {
	command.ExecRunner

	piperutils.FileUtils
}

type credentialdiggerUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newCDUtils() credentialdiggerUtils {
	utils := credentialdiggerUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func credentialdiggerTestStep(config credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) {
	//ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, "", []string{})
	//if err != nil {
	//	log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	//}
	//err = runCredentialdiggerTestStep(ctx, &config, telemetryData, client.Issues)  // commentIssue step
	//err = runCredentialdiggerTestStep(ctx, &config, telemetryData, client.Repositories)  // list repos by org
	err := runTestScanPR(&config, telemetryData) // scan PR with CD
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

func executeCredentialDiggerProcess(utils credentialdiggerUtils, args []string) error {
	return utils.RunExecutable("credentialdigger", args...)
}

func runTestScanPR(config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) error {
	service := newCDUtils()
	// 1
	log.Entry().Info("Load rules")
	cmd_list := []string{"add_rules", "--sqlite", piperTempDb, "/credential-digger-ui/backend/rules.yml"}
	err := executeCredentialDiggerProcess(service, cmd_list)
	if err != nil {
		log.Entry().Error("failed running credentialdigger add_rules")
		return err
	}
	log.Entry().Info("Rules added")

	// 2
	log.Entry().Info("Scan PR")
	log.Entry().Info("Scan PR ", config.Number, " from repo ", config.Repository)
	log.Entry().Infof("  Token: '%s'", config.Token)
	cmd_list = []string{"scan_pr", config.Repository, "--sqlite", piperTempDb,
		"--pr", strconv.Itoa(config.Number),
		"--debug",
		"--force",
		"--api_endpoint", config.APIURL,
		"--git_token", config.Token}
	leaks := executeCredentialDiggerProcess(service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks in this PR")
		// log.Entry().Warn("%v potential leaks found", leaks)
	} else {
		log.Entry().Info("No leaks found")
		// There is no need to print the discoveries if there are none
		return nil
	}

	// 3
	log.Entry().Info("Get discoveries")
	cmd_list = []string{"get_discoveries", config.Repository, "--sqlite", piperTempDb,
		"--state", "new",
		"--save", reportTempName}
	err = executeCredentialDiggerProcess(service, cmd_list)
	if err != nil {
		log.Entry().Error("failed running credentialdigger get_discoveries")
		return err
	}
	log.Entry().Info("Scan complete")

	return nil
}

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
