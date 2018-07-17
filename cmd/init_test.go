package cmd_test

import (
	"errors"
	"reflect"
	"testing"

	. "github.com/damianoneill/nc-hammer/cmd"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBuildTestSuite(t *testing.T) {

	mockPath := ""
	got := BuildTestSuite(mockPath)

	var want suite.TestSuite

	if got.File != mockPath {
		t.Errorf("got %v, want %v", got.File, mockPath)
	}
	if got.Iterations != 5 {
		t.Errorf("got %v, want %v", got.Iterations, 5)
	}
	if got.Clients != 2 {
		t.Errorf("got %v, want %v", got.Clients, 2)
	}
	if got.Rampup != 0 {
		t.Errorf("got %v, want %v", got.Rampup, 0)
	}
	if reflect.ValueOf(got.Configs).Type() != reflect.TypeOf(want.Configs) { // should I look at the values contained within here?
		t.Error("Testsuite.Configs is not of type Configs")
	}
	if reflect.ValueOf(got.Blocks).Type() != reflect.TypeOf(want.Blocks) {
		t.Error("Testsuite.Configs is not of type Blocks")
	}
}
func TestInitCmd(t *testing.T) {
	var testCmd = InitCmd
	var cmd = &cobra.Command{}

	testFunc := func(t *testing.T, args []string, want error) {
		t.Helper()

		got := testCmd.Args(cmd, args)
		if want != nil {
			if got.Error() != want.Error() {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		assert.Equal(t, got, want) // check for nil value returned
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

}
