package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialdiggerTestStepCommand(t *testing.T) {
	t.Parallel()

	testCmd := CredentialdiggerTestStepCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "credentialdiggerTestStep", testCmd.Use, "command name incorrect")

}
