package cmd

import (
	"strconv"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

const piperTempDb string = "piper_step_db.db"
const reportTempName string = "findings.csv"

type credentialdiggerScanPullRequestUtils interface {
	command.ExecRunner

	piperutils.FileUtils
	//FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The credentialdiggerScanPullRequestUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

func executeCredentialDigger(utils credentialdiggerScanPullRequestUtils, args []string) error {
	return utils.RunExecutable("credentialdigger", args...)
}

func verifyGithubConnection(config credentialdiggerScanPullRequestOptions) {
	ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, "", nil)
	if err != nil {
		log.Entry().WithError(err).Warning("Failed to get GitHub client")
		log.Entry().Error(err)
	} else {
		log.Entry().Info("GitHub client works")
	}
}

type credentialdiggerScanPullRequestUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to credentialdiggerScanPullRequestUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// credentialdiggerScanPullRequestUtilsBundle and forward to the implementation of the dependency.
	// TODO
}

func newCredentialdiggerScanPullRequestUtils() credentialdiggerScanPullRequestUtils {
	utils := credentialdiggerScanPullRequestUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func credentialdiggerScanPullRequest(config credentialdiggerScanPullRequestOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newCredentialdiggerScanPullRequestUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runCredentialdiggerScanPullRequest(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Credential Digger scan failed")
	}
}

func runCredentialdiggerScanPullRequest(config *credentialdiggerScanPullRequestOptions, telemetryData *telemetry.CustomData, utils credentialdiggerScanPullRequestUtils) error {
	log.Entry().Info("Execute scan of pull request with Credential Digger")
	verifyGithubConnection(&config)

	log.Entry().Info("Load rules")
	// TODO: dump rules to file
	// TODO: add rules from temp file
	cmd_list := []string{"add_rules", "--sqlite", piperTempDb, "/credential-digger-ui/backend/rules.yml"}
	err := executeCredentialDigger(utils, cmd_list)
	if err != nil {
		log.Entry().Error("failed running credentialdigger add_rules")
		return err
	}
	log.Entry().Info("Rules added")
	//res := exec.Command("sqlite3", piperTempDb, "\"select * from rules;\"").Run()
	// res := exec.Command("python -c", "\"", "import sqlite3; conn = sqlite3.connect('", piperTempDb,
	// 	"'); cursor=conn.cursor(); print(cursor.execute('select * from rules').fetchall())", "\"").Run()
	// log.Entry().Info("%v", res)

	log.Entry().Info("Scan PR")
	//log.Entry().Warn("Use token %v", config.Token)
	log.Entry().Infof("  Token: '%s'", config.Token)
	// TODO
	cmd_list = []string{"scan_pr", config.Repository, "--sqlite", piperTempDb,
		"--pr", strconv.Itoa(config.PrNumber),
		"--debug",
		"--api_endpoint", config.APIURL,
		"--git_token", config.Token}

	// Inherit Debug from general pipeline "Verbose" parameter
	// if GeneralConfig.Verbose {
	// 	log.Entry().Debug("Execute scan in debug mode")
	// 	cmd_list = append(cmd_list, "--debug")
	// }
	// TODO: append models

	leaks := executeCredentialDigger(utils, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks in this PR")
		// log.Entry().Warn("%v potential leaks found", leaks)
		log.Entry().Warn("%v", leaks)
	} else {
		log.Entry().Info("No leaks found")
		// There is no need to print the discoveries if there are none
		return nil
	}
	// res = exec.Command("python -c", "\"", "import sqlite3; conn = sqlite3.connect('", piperTempDb,
	// 	"'); cursor=conn.cursor(); print(cursor.execute('select * from discoveries').fetchall())", "\"").Run()
	// log.Entry().Info("%v", res)

	// Get discoveries
	log.Entry().Info("Get discoveries")
	cmd_list = []string{"get_discoveries", config.Repository, "--sqlite", piperTempDb,
		"--state", "new",
		"--save", reportTempName}
	err = executeCredentialDigger(utils, cmd_list)
	if err != nil {
		log.Entry().Error("failed running credentialdigger get_discoveries")
		return err
	}

	log.Entry().Info("Scan complete")

	// Persist the report in the workspace
	reports := []piperutils.Path{{Target: reportTempName, Name: "reportFindings", Mandatory: true}}
	err = piperutils.PersistReportsAndLinks("credentialdiggerScanPullRequest", "", utils, reports, nil)
	if err != nil {
		log.Entry().Error("Failed to persist report")
	} else {
		log.Entry().Info("Report successfully persisted")
	}

	// Example of calling methods from external dependencies directly on utils:
	//exists, err := utils.FileExists("file.txt")
	//if err != nil {
	//	// It is good practice to set an error category.
	//	// Most likely you want to do this at the place where enough context is known.
	//	log.SetErrorCategory(log.ErrorConfiguration)
	//	// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
	//	return fmt.Errorf("failed to check for important file: %w", err)
	//}
	//if !exists {
	//	log.SetErrorCategory(log.ErrorConfiguration)
	//	return fmt.Errorf("cannot run without important file")
	//}

	return nil
}
