package action_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/damianoneill/nc-hammer/action"
	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/stretchr/testify/assert"
)

func TestExecuteNetconf(t *testing.T) {

	fmt.Println()
	fmt.Println("-- Starting netconf test --")

	testFunc := func(t *testing.T, mockAction suite.Action, mockConfig suite.Sshconfig, want string) {
		t.Helper()

		// make mock channels to receive data from servers
		var mockResultChan = make(chan result.NetconfResult)
		var handleResultsFinished = make(chan bool)

		var got = ""
		// run the channels
		var results = []result.NetconfResult{}
		go func(resultChannel chan result.NetconfResult, handleResultsFinished chan bool) {
			for result := range resultChannel {
				results = append(results, result)
				if result.Err == "" {
					got = "."
					fmt.Println(got)
				} else {
					got = "E"
					fmt.Println()
				}
				// change this for what you want
			}
			handleResultsFinished <- true
		}(mockResultChan, handleResultsFinished) // run channels

		var mockStartTime = time.Now()
		// feed data to channels
		action.ExecuteNetconf(mockStartTime, 0, mockAction, &mockConfig, mockResultChan)

		close(mockResultChan)
		<-handleResultsFinished

		assert.Equal(t, got, want)
	}

	t.Run("getSessions() returns err", func(t *testing.T) {
		// mock the netconf settings to test ExecuteNetconf

		var sl = suite.Sleep{Duration: 0}
		var testString = new(string)
		var nc = suite.Netconf{Hostname: "172.26.138.91", Operation: "get", Expected: testString}
		var mockAction = suite.Action{Netconf: &nc, Sleep: &sl}

		var mockConfig = suite.Sshconfig{Hostname: "172.26.138.91", Port: 830, Username: "netconf", Password: "netconf", Reuseconnection: false}

		var want = "."
		testFunc(t, mockAction, mockConfig, want)
	})

	t.Run("getSessions() returns err; no port specified", func(t *testing.T) {
		// mock the netconf settings to test ExecuteNetconf

		var sl = suite.Sleep{Duration: 0}
		var testString = new(string)
		var nc = suite.Netconf{Hostname: "172.26.138.91", Operation: "get", Expected: testString}
		var mockAction = suite.Action{Netconf: &nc, Sleep: &sl}

		var mockConfig = suite.Sshconfig{Hostname: "172.26.138.91", Username: "", Password: "netconf", Reuseconnection: false}

		var want = "E"
		testFunc(t, mockAction, mockConfig, want)
	})

	t.Run("getSessions() returns nil; incorrect login details", func(t *testing.T) {
		// mock the netconf settings to test ExecuteNetconf

		var sl = suite.Sleep{Duration: 0}
		var testString = new(string)
		var nc = suite.Netconf{Hostname: "172.26.138.91", Operation: "get", Expected: testString}
		var mockAction = suite.Action{Netconf: &nc, Sleep: &sl}

		var mockConfig = suite.Sshconfig{Hostname: "172.26.138.91", Port: 830, Username: "", Password: "netconf", Reuseconnection: false}

		var want = "E"
		testFunc(t, mockAction, mockConfig, want)
	})

	t.Run("ToXMLString() fails; No operation specified", func(t *testing.T) {
		// mock the netconf settings to test ExecuteNetconf

		var sl = suite.Sleep{Duration: 0}
		var testString = new(string)
		var nc = suite.Netconf{Hostname: "172.26.138.91", Expected: testString}
		var mockAction = suite.Action{Netconf: &nc, Sleep: &sl}

		var mockConfig = suite.Sshconfig{Hostname: "172.26.138.91", Port: 830, Username: "netconf", Password: "netconf", Reuseconnection: false}

		var want = "E"
		testFunc(t, mockAction, mockConfig, want)
	})

	t.Run("Invalid XML provided", func(t *testing.T) {
		// mock the netconf settings to test ExecuteNetconf

		var sl = suite.Sleep{Duration: 0}
		var testXML = `config: <top xmlns="http://example.com/schema/1.2/config"><protocols><ospf><area><name>0.0.0.0</name><interfaces><interface
        xc:operation="delete"><name>192.0.2.4</name></interface></interfaces></area></ospf></protocols></top>`
		var nc = suite.Netconf{Hostname: "172.26.138.91", Operation: "edit-config", Config: &testXML}
		var mockAction = suite.Action{Netconf: &nc, Sleep: &sl}

		var mockConfig = suite.Sshconfig{Hostname: "172.26.138.91", Port: 830, Username: "netconf", Password: "netconf", Reuseconnection: false}

		var want = "E"
		testFunc(t, mockAction, mockConfig, want)
	})

	t.Run("Expected value not returned", func(t *testing.T) {
		// mock the netconf settings to test ExecuteNetconf

		var sl = suite.Sleep{}
		expectedString := "=invalid"
		var nc = suite.Netconf{Hostname: "172.26.138.91", Operation: "get", Expected: &expectedString}
		var mockAction = suite.Action{Netconf: &nc, Sleep: &sl}

		var mockConfig = suite.Sshconfig{Hostname: "172.26.138.91", Port: 830, Username: "netconf", Password: "netconf", Reuseconnection: false}

		var want = "E"
		testFunc(t, mockAction, mockConfig, want)
	})

	t.Run("Invalid regexp used for expected value; err returned", func(t *testing.T) {
		// mock the netconf settings to test ExecuteNetconf

		var sl = suite.Sleep{}
		expectedString := "(<[^>]+>"
		var nc = suite.Netconf{Hostname: "172.26.138.91", Operation: "get", Expected: &expectedString}
		var mockAction = suite.Action{Netconf: &nc, Sleep: &sl}

		var mockConfig = suite.Sshconfig{Hostname: "172.26.138.91", Port: 830, Username: "netconf", Password: "netconf", Reuseconnection: false}

		var want = "E"
		testFunc(t, mockAction, mockConfig, want)
	})
}
