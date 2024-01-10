package wireguard

import (
	"fmt"
	"net"
	"os"
)

type Config struct {
  InterfaceConfigPath string
	Subnet net.IPNet
}

func LoadConfig() (Config, error) {
  conf := Config {
    InterfaceConfigPath: "/etc/wireguard/wg0.conf",
		Subnet: net.IPNet {
			IP: net.IPv4(10, 32, 42, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
  } 

  if v, ok := os.LookupEnv("INTERFACE_CONFIG_PATH"); ok {
    conf.InterfaceConfigPath = v
  }

  if v, ok := os.LookupEnv("SUBNET"); ok {
		_, net, err := net.ParseCIDR(v)
		if err != nil {
			return conf, fmt.Errorf("unable to parse subnet %v: %w", v, err)
		}

		// Set all bits outside the mask to 0
		for i, v := range net.Mask {
			net.IP[i] &= ^v
		}

    conf.Subnet = *net
  }

  return conf, nil
}
