package action

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"golang.org/x/crypto/ssh"
)

var gSessions map[string]*netconf.Session

func init() {
	gSessions = make(map[string]*netconf.Session)
}

// CloseAllSessions is called on exit to gracefully close the sockets
func CloseAllSessions() {
	// nolint
	for _, session := range gSessions {
		session.Close()
	}
}

func operationOrMessage(netconf *suite.Netconf) string {
	if netconf.Operation != nil {
		return *netconf.Operation
	}
	return *netconf.Message
}

// ExecuteNetconf invoked when a NETCONF Action is identified
func ExecuteNetconf(n NetconfInterface, tsStart time.Time, cID int, action suite.Action, config *suite.Sshconfig, resultChannel chan result.NetconfResult) {

	var result result.NetconfResult
	result.Client = cID
	result.Hostname = action.Netconf.Hostname
	result.Operation = operationOrMessage(action.Netconf)

	session, err := n.GetSession(cID, config.Hostname+":"+strconv.Itoa(config.Port), config.Username, config.Password, config.Reuseconnection)
	i, _ := session.(*netconf.Session) // i know its a netconf session

	if err != nil {
		fmt.Printf("E")
		result.Err = err.Error()
		resultChannel <- result
		return
	}

	// not reusing the connection, then explicitly close it
	if !config.Reuseconnection {
		// nolint
		defer session.Close()
	}

	if i != nil {
		result.SessionID = getSessionID(session)
	} else {
		fmt.Printf("E")
		result.Err = "session has expired"
		resultChannel <- result
		return
	}

	xml, err := action.Netconf.ToXMLString()
	if err != nil {
		fmt.Printf("E")
		result.Err = err.Error()
		resultChannel <- result
		return
	}

	raw := netconf.RawMethod(xml)
	start := time.Now()
	rpcReply, err := session.Exec(raw)
	if err != nil {
		if err.Error() == "WaitForFunc failed" {
			delete(gSessions, strconv.Itoa(cID)+config.Hostname+":"+strconv.Itoa(config.Port))
			result.Err = "session closed by remote side"
		} else {
			result.Err = err.Error()
		}
		fmt.Printf("e")
		resultChannel <- result
		return
	}
	elapsed := time.Since(start)
	result.When = float64(time.Since(tsStart).Nanoseconds() / int64(time.Millisecond))
	result.Latency = float64(elapsed.Nanoseconds() / int64(time.Millisecond))

	result.MessageID = rpcReply.MessageID

	if action.Netconf.Expected != nil {
		match, err := regexp.MatchString(*action.Netconf.Expected, rpcReply.Data)
		if err != nil {
			fmt.Printf("E")
			result.Err = err.Error()
			resultChannel <- result
			return
		}
		if !match {
			fmt.Printf("e")
			result.Err = "expected response did not match, expected: " + *action.Netconf.Expected + " actual: " + rpcReply.Data
			resultChannel <- result
			return
		}
	}
	resultChannel <- result
}

// GetSession returns a NETCONF Session, either a new one or a pre existing one if resuseConnection is valid for client/host
func (n *NetconfHandler) GetSession(client int, hostname, username, password string, reuseConnection bool) (SessionInterface, error) {

	// check if hostname should reuse connection
	if reuseConnection {
		// get Session from Map if present
		session, present := gSessions[strconv.Itoa(client)+hostname]
		if present {
			return session, nil
		}
		// not present in map, therefore first time its called, create a new session and store in map
		tempSession, err := n.CreateNewSession(hostname, username, password)
		session = tempSession.(*netconf.Session)

		if err == nil {
			gSessions[strconv.Itoa(client)+hostname] = session
		}
		return session, nil
	}
	return n.CreateNewSession(hostname, username, password)
}

// CreateNewSession returns a SSH session
func (n *NetconfHandler) CreateNewSession(hostname, username, password string) (SessionInterface, error) {

	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return netconf.DialSSH(hostname, sshConfig)
}

func getSessionID(s SessionInterface) int {
	switch t := s.(type) {
	case *netconf.Session:
		return t.SessionID
	default:
		return 0
	}
}

// NetconfHandler struct
type NetconfHandler struct {
}

// NetconfInterface interface
type NetconfInterface interface {
	GetSession(client int, hostname, username, password string, reuseConnection bool) (SessionInterface, error)
	CreateNewSession(hostname, username, password string) (SessionInterface, error)
}

// Netconf struct
type Netconf struct {
	Nc NetconfInterface
}

//SessionHandler handler
type SessionHandler struct {
}

//SessionInterface interface
type SessionInterface interface {
	Close() error
	Exec(methods ...netconf.RPCMethod) (*netconf.RPCReply, error)
}

//NetconfSession struct
type NetconfSession struct {
	Ns SessionInterface
}
