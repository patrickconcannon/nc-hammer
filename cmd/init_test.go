package cmd_test

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"

	. "github.com/damianoneill/nc-hammer/cmd"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBuildTestSuite(t *testing.T) { // check into test table

	mockPath := ""
	got := BuildTestSuite(mockPath)

	var want suite.TestSuite

	if got.File != mockPath {
		t.Errorf("Filename: got %v, want %v", got.File, mockPath)
	}
	if got.Iterations != 5 {
		t.Errorf("Iterations: got %v, want %v", got.Iterations, 5)
	}
	if got.Clients != 2 {
		t.Errorf("Clients: got %v, want %v", got.Clients, 2)
	}
	if got.Rampup != 0 {
		t.Errorf("Rampup: got %v, want %v", got.Rampup, 0)
	}
	if reflect.ValueOf(got.Configs).Type() != reflect.TypeOf(want.Configs) { // should I look at the values contained within here?
		t.Error("Testsuite.Configs is not of type Configs")
	}
	if reflect.ValueOf(got.Blocks).Type() != reflect.TypeOf(want.Blocks) {
		t.Error("Testsuite.Configs is not of type Blocks")
	}
}
func TestInitCmd(t *testing.T) { // check for correct return value
	var testCmd = InitCmd
	var cmd = &cobra.Command{}

	testFunc := func(t *testing.T, args []string, want error) {
		t.Helper()

		got := testCmd.Args(cmd, args)
		assert.Equal(t, got, want)
	}

	t.Run("args != 1", func(t *testing.T) {
		var a = []string{"a", "b"}
		testFunc(t, a, errors.New("init command requires a directory as an argument"))
	})

	t.Run("args == 1", func(t *testing.T) {
		var a = []string{"a"}
		testFunc(t, a, nil)
	})
}

func TestInitRun(t *testing.T) {
	var testInitCmd = InitCmd
	var testCmd = &cobra.Command{}

	args := []string{"test/string"}

	/*
		If an error is return and exits with 1, there was a problem.
		What I have to do is to test for each particular case. The
		particular variable that will define each case will be
	*/

	// Run test as subprocess when environment variable is set as 1
	if os.Getenv("RUN_SUBPROCESS") == "1" {
		testInitCmd.Run(testCmd, args)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestAnalyseRun") // create new process to run test
	cmd.Env = append(os.Environ(), "RUN_SUBPROCESS=1")          // set environmental variable
	err := cmd.Run()                                            // run
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {     // check exit status of test subprocess
		t.Errorf("\n - Exit Status 1 returned\n - File '%v' already exists", args[0])
	}
}
