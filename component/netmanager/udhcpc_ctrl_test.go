package netmanager

import (
	"strings"
	"testing"
)

// mockRunner is a mock implementation of CommandRunner
type mockRunner struct {
	Commands []string
	Output   map[string]string
	Err      map[string]error
}

func (m *mockRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := name + " " + strings.Join(args, " ")
	m.Commands = append(m.Commands, cmd)

	if err, exists := m.Err[cmd]; exists {
		return nil, err
	}
	if out, exists := m.Output[cmd]; exists {
		return []byte(out), nil
	}
	return []byte(""), nil
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		Output: make(map[string]string),
		Err:    make(map[string]error),
	}
}

func TestStart_Success(t *testing.T) {
	mock := newMockRunner()
	ctrl := &UDHCPCController{
		Config: Config{
			Interface: "eth0",
			Timeout:   5,
		},
	}

	mock.Output["udhcpc -i eth0 -T 5"] = "bound"

	err := ctrl.Start()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestStart_MissingInterface(t *testing.T) {
	ctrl := &UDHCPCController{
		Config: Config{},
	}

	err := ctrl.Start()
	if err == nil {
		t.Fatal("expected error for missing interface")
	}
}

func TestRelease_Success(t *testing.T) {
	mock := newMockRunner()
	ctrl := &UDHCPCController{
		Config: Config{Interface: "eth0"},
	}
	mock.Output["udhcpc -i eth0 -n -q -R"] = "released"

	err := ctrl.Release()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestGetLeaseInfo_Success(t *testing.T) {
	mock := newMockRunner()
	ctrl := &UDHCPCController{
		Config: Config{Interface: "eth0"},
	}

	mock.Output["ip -4 addr show dev eth0"] = `
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
    inet 192.168.1.100/24 brd 192.168.1.255 scope global eth0
`
	mock.Output["ip route show default dev eth0"] = `default via 192.168.1.1 dev eth0`
	mock.Output["cat /etc/resolv.conf"] = `
nameserver 8.8.8.8
nameserver 1.1.1.1
`

	lease, err := ctrl.GetLeaseInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lease.IPAddress != "192.168.1.100" {
		t.Errorf("unexpected IP: %s", lease.IPAddress)
	}
	if lease.Gateway != "192.168.1.1" {
		t.Errorf("unexpected gateway: %s", lease.Gateway)
	}
	if len(lease.DNS) != 2 || lease.DNS[0] != "8.8.8.8" {
		t.Errorf("unexpected DNS: %v", lease.DNS)
	}
}

func TestGetLeaseInfo_NoIP(t *testing.T) {
	mock := newMockRunner()
	ctrl := &UDHCPCController{
		Config: Config{Interface: "eth0"},
	}

	mock.Output["ip -4 addr show dev eth0"] = `
2: eth0: <NO-CARRIER> mtu 1500
    inet6 fe80::1/64 scope link
`

	_, err := ctrl.GetLeaseInfo()
	if err == nil {
		t.Fatal("expected error due to missing IP")
	}
}
