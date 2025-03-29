package netmanager

type NetworkInterfaceParam struct {
	Type        string
	Interface   string
	Address     string
	Netmask     string
	Gateway     string
	DHCPEnabled bool
	SSID        string
	Password    string
	SecureType  string
}
