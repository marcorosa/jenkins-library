package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type credentialdiggerScanPullRequestMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCredentialdiggerScanPullRequestTestsUtils() credentialdiggerScanPullRequestMockUtils {
	utils := credentialdiggerScanPullRequestMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCredentialdiggerScanPullRequest(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := credentialdiggerScanPullRequestOptions{}

		utils := newCredentialdiggerScanPullRequestTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runCredentialdiggerScanPullRequest(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := credentialdiggerScanPullRequestOptions{}

		utils := newCredentialdiggerScanPullRequestTestsUtils()

		// test
		err := runCredentialdiggerScanPullRequest(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
