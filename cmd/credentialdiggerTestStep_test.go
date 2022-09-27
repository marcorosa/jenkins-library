package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type credentialdiggerTestStepMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCredentialdiggerTestStepTestsUtils() credentialdiggerTestStepMockUtils {
	utils := credentialdiggerTestStepMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCredentialdiggerTestStep(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := credentialdiggerTestStepOptions{}

		utils := newCredentialdiggerTestStepTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runCredentialdiggerTestStep(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := credentialdiggerTestStepOptions{}

		utils := newCredentialdiggerTestStepTestsUtils()

		// test
		err := runCredentialdiggerTestStep(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
