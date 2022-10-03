package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type credentialdiggerScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCredentialdiggerScanTestsUtils() credentialdiggerScanMockUtils {
	utils := credentialdiggerScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCredentialdiggerScan(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := credentialdiggerScanOptions{}

		utils := newCredentialdiggerScanTestsUtils()
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
