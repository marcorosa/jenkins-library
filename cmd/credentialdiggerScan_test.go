package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type credentialdiggerScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCDTestsUtils() credentialdiggerScanMockUtils {
	utils := credentialdiggerScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestCredentialdiggerFullScan(t *testing.T) {
	t.Run("Valid full scan with discoveries", func(t *testing.T) {
		// TODO
	})
	t.Run("Valid full scan without discoveries", func(t *testing.T) {
		// TODO
	})

	t.Run("Invalid repository", func(t *testing.T) {
		// TODO
	})
	t.Run("Invalid git token", func(t *testing.T) {
		// TODO
	})
	t.Run("Invalid API endpoint", func(t *testing.T) {
		// TODO
	})
	t.Run("Invalid ML models", func(t *testing.T) {
		// TODO
	})
}

func TestCredentialdiggerScanSnapshot(t *testing.T) {
	t.Run("Valid scan snapshot with discoveries", func(t *testing.T) {
		// TODO
	})
	t.Run("Valid scan snapshot without discoveries", func(t *testing.T) {
		// TODO
	})
	t.Run("Invalid snapshot in scan snapshot", func(t *testing.T) {
		// TODO
	})
}

func TestCredentialdiggerScanPR(t *testing.T) {
	t.Run("Valid scan pull request with discoveries", func(t *testing.T) {
		// TODO
	})
	t.Run("Valid scan pull request without discoveries", func(t *testing.T) {
		// TODO
	})
	t.Run("Invalid pr number in scan pull request", func(t *testing.T) {
		// TODO
	})
}

func TestCredentialdiggerAddRules(t *testing.T) {
	t.Run("Valid standard rules", func(t *testing.T) {
		config := credentialdiggerScanOptions{}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerAddRules(&config, nil, utils))
	})
	t.Run("Valid external rules", func(t *testing.T) {
		rulesExt := "https://raw.githubusercontent.com/SAP/credential-digger/main/ui/backend/rules.yml"
		config := credentialdiggerScanOptions{RulesDownloadURL: rulesExt}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerAddRules(&config, nil, utils))
	})
	t.Run("Invalid external rules link", func(t *testing.T) {
		rulesExt := "https://broken-link.com/fakerules"
		config := credentialdiggerScanOptions{RulesDownloadURL: rulesExt}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerAddRules(&config, nil, utils))
	})
	t.Run("Invalid external rules file format", func(t *testing.T) {
		rulesExt := "https://raw.githubusercontent.com/SAP/credential-digger/main/requirements.txt"
		config := credentialdiggerScanOptions{RulesDownloadURL: rulesExt}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerAddRules(&config, nil, utils))
	})

}

/*
func TestRunCredentialdiggerScan(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := credentialdiggerScanOptions{}

		utils := newCDTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runCredentialdiggerScan(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := credentialdiggerScanOptions{}

		utils := newCredentialdiggerScanTestsUtils()

		// test
		err := runCredentialdiggerScan(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})

}
*/

func TestCredentialdiggerGetDiscoveries(t *testing.T) {
	t.Run("Valid get discoveries", func(t *testing.T) {
		// TODO
	})
	t.Run("Empty discoveries", func(t *testing.T) {
		// TODO
	})
}
