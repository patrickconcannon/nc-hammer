package action

import (
	"testing"

	"github.com/Juniper/go-netconf/netconf"
)

type MockNetconf struct {
}

func getMockNetconf() *NetconfHandler {
	m := new(NetconfHandler)
	m.Nc = new(MockNetconf)
	return m
}

func (m *MockNetconf) createNewSession(hostname, username, password string) (*netconf.Session, error) {
	return nil, nil
}
func (m *MockNetconf) getSession(client int, hostname, username, password string, reuseConnection bool) (*netconf.Session, error) {
	return nil, nil
}

func TestExeceuteNetconf(t *testing.T) {

}
