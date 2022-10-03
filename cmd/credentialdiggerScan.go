package cmd

import (
	"context"
	"strconv"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v45/github"
)

const piperDbName string = "piper_step_db.db"
const piperReportName string = "findings.csv"

type credentialdiggerScanService interface {
	List(ctx context.Context, owner string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error)
}

type credentialdiggerUtils interface {
	command.ExecRunner

	piperutils.FileUtils

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The credentialdiggerScanUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type credentialdiggerUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to credentialdiggerScanUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// credentialdiggerScanUtilsBundle and forward to the implementation of the dependency.
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

func credentialdiggerScan(config credentialdiggerScanOptions, telemetryData *telemetry.CustomData) error {
	utils := newCDUtils()
	// 1: Add rules
	log.Entry().Info("Load rules")
	err := credentialdiggerAddRules(&config, telemetryData, &utils)
	if err != nil {
		log.Entry().Error("Failed running credentialdigger add_rules")
		return err
	}
	log.Entry().Info("Rules added")

	// 2: Scan the repository
	// Choose between scan-snapshot, scan-pr, and full-scan (with this priority
	// order)
	switch {
	case config.Snapshot != "":
		log.Entry().Debug("Scan snapshot")
		// if a Snapshot is declared, run scan_snapshot
		err = credentialdiggerScanSnapshot(&config, telemetryData, &utils) // scan Snapshot with CD
	case config.PrNumber != 0: // int type is not nillable in golang
		log.Entry().Debug("Scan PR")
		// if a PrNumber is declared, run scan_pr
		err = credentialdiggerScanPR(&config, telemetryData, &utils) // scan PR with CD
	default:
		// The default case is the normal full scan
		log.Entry().Debug("Full scan repo")
		err = credentialdiggerFullScan(&config, telemetryData, &utils) // full scan with CD
	}

	// 3: Get discoveries
	err = credentialdiggerGetDiscoveries(&config, telemetryData, &utils)
	if err != nil {
		// The exit number is the number of discoveries
		//log.Entry().WithError(err).Fatal("Failed to run custom function")
		log.Entry().Errorf("There are %v findings to review", err)
	}

	return nil
}

func executeCredentialDiggerProcess(utils credentialdiggerUtils, args []string) error {
	return utils.RunExecutable("credentialdigger", args...)
}

func credentialdiggerAddRules(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service *credentialdiggerUtils) error {
	// TODO: implement custom rules support
	cmd_list := []string{"add_rules", "--sqlite", piperDbName, "/credential-digger-ui/backend/rules.yml"}
	return executeCredentialDiggerProcess(*service, cmd_list)
}

func credentialdiggerGetDiscoveries(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service *credentialdiggerUtils) error {
	log.Entry().Info("Get discoveries")
	cmd_list := []string{"get_discoveries", config.Repository, "--sqlite", piperDbName,
		"--save", piperReportName}
	// Export all the discoveries or export only new ones
	if !config.ExportAll {
		cmd_list = append(cmd_list, "--state", "new")
	}
	err := executeCredentialDiggerProcess(*service, cmd_list)
	if err != nil {
		log.Entry().Error("failed running credentialdigger get_discoveries")
		log.Entry().Error(err)
		return err
	}
	log.Entry().Info("Scan complete")
	return nil
}

func credentialdiggerBuildCommonArgs(config *credentialdiggerScanOptions) []string {
	/*Some arguments are the same for all the scan flavors. Build them here
	* not to duplicate code.*/
	scan_args := []string{}
	// Repository url and sqlite db (always mandatory)
	scan_args = append(scan_args, config.Repository, "--sqlite", piperDbName)
	//git token is not mandatory for base credential digger tool, but in
	//piper it is
	scan_args = append(scan_args, "--git_token", config.Token)
	//debug
	if config.Debug {
		log.Entry().Debug("Run the scan in debug mode")
		scan_args = append(scan_args, "--debug")
	}
	//models
	if len(config.Models) > 0 {
		log.Entry().Debugf("Enable models %v", config.Models)
		scan_args = append(scan_args, "--models")
		scan_args = append(scan_args, config.Models...)
	}

	return scan_args
}

func credentialdiggerScanSnapshot(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service *credentialdiggerUtils) error {
	log.Entry().Infof("Scan Snapshot %v from repo %v", config.Snapshot, config.Repository)
	cmd_list := []string{"scan_snapshot",
		"--snapshot", config.Snapshot}
	cmd_list = append(cmd_list, credentialdiggerBuildCommonArgs(config)...)
	leaks := executeCredentialDiggerProcess(*service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks in this Snapshot")
		return leaks
	} else {
		log.Entry().Info("No leaks found")
		return nil
	}
}

func credentialdiggerScanPR(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service *credentialdiggerUtils) error {
	log.Entry().Infof("Scan PR %v from repo %v", config.PrNumber, config.Repository)
	cmd_list := []string{"scan_pr",
		"--pr", strconv.Itoa(config.PrNumber),
		"--api_endpoint", config.APIURL}
	cmd_list = append(cmd_list, credentialdiggerBuildCommonArgs(config)...)
	leaks := executeCredentialDiggerProcess(*service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks in this PR")
		return leaks
	} else {
		log.Entry().Info("No leaks found")
		return nil
	}
}

func credentialdiggerFullScan(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service *credentialdiggerUtils) error {
	log.Entry().Infof("Full scan of repository %v", config.Repository)
	cmd_list := []string{"scan"}
	cmd_list = append(cmd_list, credentialdiggerBuildCommonArgs(config)...)
	leaks := executeCredentialDiggerProcess(*service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks")
		log.Entry().Warnf("%v potential leaks found", leaks)
		return leaks
	} else {
		log.Entry().Info("No leaks found")
		return nil
	}
}
