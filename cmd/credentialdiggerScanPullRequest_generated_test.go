package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialdiggerScanPullRequestCommand(t *testing.T) {
	t.Parallel()

	testCmd := CredentialdiggerScanPullRequestCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "credentialdiggerScanPullRequest", testCmd.Use, "command name incorrect")

}
