package cli_test

import (
	"bytes"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/cli"
)

func TestVerboseAcceptedByVersion(t *testing.T) {
	// --verbose is accepted by the version subcommand without error
	svc := &ax.MockWindowService{}
	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs([]string{"version", "--verbose"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version --verbose should not error, got: %v", err)
	}
}

func TestVerboseAcceptedByCompletion(t *testing.T) {
	// --verbose is accepted by the completion subcommand without error
	svc := &ax.MockWindowService{}
	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs([]string{"completion", "--verbose", "bash"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("completion --verbose should not error, got: %v", err)
	}
}

func TestHelpVerboseNotCorrupted(t *testing.T) {
	// --help --verbose produces normal help output
	svc := &ax.MockWindowService{}
	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs([]string{"--help", "--verbose"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	_ = cmd.Execute()
	output := out.String()
	if output == "" {
		t.Error("--help --verbose should produce help output")
	}
	// help output should contain the command name
	if !bytes.Contains([]byte(output), []byte("mado")) {
		t.Errorf("help output should contain 'mado', got:\n%s", output)
	}
}

func TestVerbosef_SilentOnWriteFailure(_ *testing.T) {
	// Verbosef does not panic on write failure
	w := &failWriter{}
	// if the function returns normally without panic, the test passes
	cli.Verbosef(true, w, "test %s", "message")
}

// failWriter is a Writer that always returns an error.
type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}
