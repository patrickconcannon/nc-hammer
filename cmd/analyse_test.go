package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/damianoneill/nc-hammer/cmd"
	"github.com/damianoneill/nc-hammer/result"
	. "github.com/damianoneill/nc-hammer/suite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat"
)

// Test variables used to populate []result.NetconfResult used throughout
var (
	ts1 = result.NetconfResult{Client: 5, SessionID: 2318, Hostname: "172.26.138.91", Operation: "edit-config", When: 55282, Err: "", Latency: 288}
	ts2 = result.NetconfResult{Client: 6, SessionID: 859, Hostname: "172.26.138.92", Operation: "get-config", When: 55943, Err: "", Latency: 176}
	ts3 = result.NetconfResult{Client: 4, SessionID: 601, Hostname: "172.26.138.93", Operation: "get", When: 9840, Err: "", Latency: 3320}
	ts4 = result.NetconfResult{Client: 4, SessionID: 2322, Hostname: "172.26.138.91", Operation: "get", When: 56967, Err: "", Latency: 420}
	ts5 = result.NetconfResult{Client: 4, SessionID: 860, Hostname: "172.26.138.92", Operation: "kill-session", When: 0, Err: "kill-session is not a supported operation", Latency: 0}
)

func TestSortResults(t *testing.T) {

	testSort := func(t *testing.T, unsortedSlice []result.NetconfResult, want []result.NetconfResult) {
		t.Helper()

		SortResults(unsortedSlice)
		got := unsortedSlice

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	t.Run("Sort by Hostname", func(t *testing.T) {
		unsortedSlice := []result.NetconfResult{ts3, ts1, ts2}
		want := []result.NetconfResult{ts1, ts2, ts3}

		testSort(t, unsortedSlice, want)
	})

	t.Run("Sort by Operation", func(t *testing.T) {
		unsortedSlice := []result.NetconfResult{ts3, ts4, ts2, ts5, ts1}
		want := []result.NetconfResult{ts1, ts4, ts2, ts5, ts3}

		testSort(t, unsortedSlice, want)
	})
}

func TestOrderAndExcludeErrValues(t *testing.T) {
	testResults := []result.NetconfResult{ts1, ts2, ts3, ts4, ts5}
	testLatencies := make(map[string]map[string][]float64)

	got := OrderAndExcludeErrValues(testResults, testLatencies)

	if got != 1 {
		t.Errorf("got %v, want %v", 1, got)
	}
}

func TestAnalyseResults(t *testing.T) {
	var latencies = make(map[string]map[string][]float64)
	var mockCmd *cobra.Command
	var mockResults = []result.NetconfResult{ts1, ts2, ts3}
	var mockTs = TestSuite{}
	mockTs.File = "testdata/emptytestsuite.yml"

	// Capture StdErr
	var lOut = new(bytes.Buffer)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamps
	log.SetOutput(lOut)                                  // log output

	// Capture StdOut
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	AnalyseResults(mockCmd, &mockTs, mockResults)

	out := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		defer r.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		out <- buf.String()
		buf.Reset()
	}()

	w.Close()
	os.Stdout = old
	cOut := <-out // console output

	testRun := func(t *testing.T, got, want string) {
		t.Helper()

		assert.Equal(t, got, want)
	}

	t.Run("Log check", func(t *testing.T) {

		var logBuffer bytes.Buffer

		logBuffer.WriteString("Testsuite executed at " + strings.Split(mockTs.File, string(filepath.Separator))[1] + " Suite defined the following hosts: ")
		logBuffer.WriteString("[")
		for _, config := range mockTs.Configs {
			logBuffer.WriteString(config.Hostname + " ")
		}
		logBuffer.WriteString("] ")

		errCount := OrderAndExcludeErrValues(mockResults, latencies)

		var when float64
		for _, result := range mockResults {
			if result.When > when {
				when = result.When
			}
		}
		executionTime := time.Duration(when) * time.Millisecond

		logBuffer.WriteString(strconv.Itoa(mockTs.Clients) + " client(s) started, " + strconv.Itoa(mockTs.Iterations) + " iterations per client, " + strconv.Itoa(mockTs.Rampup) + " seconds wait between starting each client ")
		logBuffer.WriteString(" Total execution time: " + executionTime.String() + ", Suite execution contained " + strconv.Itoa(errCount) + " errors")

		// Format logString

		re := regexp.MustCompile(`\r?\n`)
		got := strings.Trim(re.ReplaceAllString(lOut.String(), " "), " ")

		testRun(t, got, logBuffer.String())
	})

	t.Run("Console check", func(t *testing.T) {
		// TODO: Add test cases to capture op and hostname test cases --
		/*
			mockCmd.SetArgs([]string{ // sets flags
				"test",
				//fmt.Sprintf("--some-string=%s", value),
				//fmt.Sprintf("--some-integer=%d", integer),
			})
		*/
		op := ""
		hostname := ""

		// sort latencies Map by key
		var keys []string
		for k := range latencies {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		consoleBuffer := new(bytes.Buffer)
		consoleBuffer.WriteString("HOST OPERATION REUSE CONNECTION REQUESTS TPS MEAN VARIANCE STD DEVIATION ")
		for _, k := range keys {
			host := k
			operations := latencies[k]
			//	for host, operations := range latencies {
			for operation, latencies := range operations {
				if op != "" && op != operation {
					continue
				}
				if hostname != "" && hostname != host {
					continue
				}
				mean := stat.Mean(latencies, nil)
				tps := 1000 / mean
				variance := stat.Variance(latencies, nil)
				stddev := math.Sqrt(variance)
				consoleBuffer.WriteString(host + " " + operation + " " + strconv.FormatBool(mockTs.Configs.IsReuseConnection(host)) + " " + strconv.Itoa(len(latencies)) + " " + fmt.Sprintf("%.2f", tps) + " " + fmt.Sprintf("%.2f", mean) + " " + fmt.Sprintf("%.2f", variance) + " " + fmt.Sprintf("%.2f", stddev) + " ")
			}
		}
		// remove formating from test table
		removeWhtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`) // remove whitespace outside required string
		want := removeWhtsp.ReplaceAllString(cOut, "")
		removeWhtsp = regexp.MustCompile(`[\s\p{Zs}]{2,}`) // remove whitespace inside required string
		want = removeWhtsp.ReplaceAllString(want, " ")

		got := strings.Trim(consoleBuffer.String(), " ")

		testRun(t, got, want)
	})
}

func TestAnalyseArgs(t *testing.T) {

	var testCmd = AnalyseCmd
	var cmd = &cobra.Command{}

	testStruct := func(t *testing.T, args []string, got error) {
		t.Helper()

		want := testCmd.Args(cmd, args) // args = 1 or != 1
		assert.Equal(t, got, want)
	}

	t.Run("args == 1", func(t *testing.T) {
		var a = []string{"a"}

		testStruct(t, a, nil)
	})

	t.Run("args != 1", func(t *testing.T) {
		b := []string{"a", "b", "c"}
		rstr := errors.New("analyse command requires a test results directory as an argument")

		testStruct(t, b, rstr)
	})
}

func TestAnalyseRun(t *testing.T) {

	tests := []struct {
		name     string
		testCmd  *cobra.Command
		testArgs []string
	}{
		// TODO: add more use cases
		{name: "single valid yaml file", testArgs: []string{"../suite/testdata/"}},
	}

	for _, tt := range tests {

		// Run test as subprocess when environment variable is set as 1
		if os.Getenv("RUN_SUBPROCESS") == "1" {
			AnalyseCmd.Run(tt.testCmd, tt.testArgs)
			return
		}

		cmd := exec.Command(os.Args[0], "-test.run=TestAnalyseRun") // create new process to run test
		cmd.Env = append(os.Environ(), "RUN_SUBPROCESS=1")          // set environmental variable
		err := cmd.Run()                                            // run
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {     // check exit status of test subprocess
			t.Fatalf("Program failed to load file -- os.Exit(1)")
		}
	}
}
