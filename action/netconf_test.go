package action_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/damianoneill/nc-hammer/action"
	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/stretchr/testify/assert"
)

type MockNetconf struct {
}

func getMockNetconf() *action.Netconf {
	m := new(action.Netconf)
	m.Nc = new(MockNetconf)
	return m
}

func (m *MockNetconf) CreateNewSession(hostname, username, password string) (action.SessionInterface, error) {
	return &MockSession{}, nil
}
func (m *MockNetconf) GetSession(client int, hostname, username, password string, reuseConnection bool) (action.SessionInterface, error) {
	return m.CreateNewSession("", "", "")
}

func TestExeceuteNetconf(t *testing.T) {

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
	var sl = suite.Sleep{Duration: 0}
	//var testString = new(string)

	operationToCallString := "edit-config"
	var opString = &operationToCallString
	var mockXML = `config: <top xmlns="http://example.com/schema/1.2/config"><protocols><ospf><area><name>0.0.0.0</name><interfaces><interface
	xc:operation="delete"><name>192.0.2.4</name></interface></interfaces></area></ospf></protocols></top>`

	//var nc = suite.Netconf{Hostname: "172.26.138.91", Operation: opString, Expected: testString}
	var mockNetConf = suite.Netconf{Hostname: "10.0.0.1", Operation: opString, Config: &mockXML}
	var mockAction = suite.Action{Netconf: &mockNetConf, Sleep: &sl}
	var mockConfig = suite.Sshconfig{Hostname: "10.0.0.1", Port: 830, Username: "username", Password: "password", Reuseconnection: false}

	var want = "E"

	m := getMockNetconf()
	action.ExecuteNetconf(m.Nc, mockStartTime, 0, mockAction, &mockConfig, mockResultChan)

	close(mockResultChan)
	<-handleResultsFinished

	assert.Equal(t, got, want)
}

type MockSession struct {
	SessionID int
}

// Close close funciton
func (m *MockSession) Close() error {
	return nil
}

// Exec Exec fuction
func (m *MockSession) Exec(methods ...netconf.RPCMethod) (*netconf.RPCReply, error) {
	r := netconf.RPCReply{MessageID: "test"}
	return &r, nil
}
