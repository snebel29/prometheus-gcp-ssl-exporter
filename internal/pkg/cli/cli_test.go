package cli

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestCLIWrongArgsExitCode(t *testing.T) {
	expectedExitCode := 1
	if os.Getenv("TESTING") == "true" {
		NewCLI()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCLIWrongArgsExitCode")
	cmd.Env = append(os.Environ(), "TESTING=true")
	err := cmd.Run()
	if err.Error() == fmt.Sprintf("exit status %d", expectedExitCode) {
		return
	}
	t.Errorf("process ran with %v, want exit status %d", err, expectedExitCode)
}

func TestCLIGoodArgumentsParse(t *testing.T) {
	m := "/metrics-test"
	p := "6666"
	project := "project-1"

	os.Args = []string{
		"binaryName",
		fmt.Sprintf("--metrics-path=%s", m),
		fmt.Sprintf("--port=%s", p),
		"--only-in-use",
		fmt.Sprintf("--project=%s", project),
		fmt.Sprintf("--project=%s", "project-2"),
	}
	cli := NewCLI()
	if cli.MetricsPath != m || cli.Port != p {
		t.Errorf("%s != %s || %s != %s", cli.MetricsPath, m, cli.Port, p)
	}
}