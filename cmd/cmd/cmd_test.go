package cmd

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartCmd(t *testing.T) {
	actual := new(bytes.Buffer)

	// Override rootCmd's output to the test buffer
	rootCmd.SetOut(actual)
	rootCmd.SetErr(actual)

	// Reconfigure slog to write to the same buffer
	logger := slog.New(slog.NewTextHandler(actual, nil))
	slog.SetDefault(logger)

	// Set the command arguments
	rootCmd.SetArgs([]string{"cmd", "start", "testnetwork"})

	// Execute the command
	err := rootCmd.Execute()

	// Assert no error occurred
	assert.NoError(t, err)

	// Check that the expected message is in the buffer
	assert.Contains(t, actual.String(), "network testnetwork is not configured", "Expected output is not found")
}
