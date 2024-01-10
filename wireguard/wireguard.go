package wireguard

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/inconshreveable/log15"
	"github.com/pmorjan/kmod"
)

const interfaceName = "wg0"

// RunServer starts the wireguard server and watches the config file for changes.
func RunServer() error {
	conf, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("unable to load the config: %w", err)
	}

	// Ensure the module is loaded
	err = loadModule()
	if err != nil {
		return fmt.Errorf("unable to load the wireguard module: %w", err)
	}

	// Create the interface
	err = createInterface(&conf)
	if err != nil {
		return fmt.Errorf("unable to setup the interface: %w", err)
	}

	// Setup the iptable rules for forwarding
	err = setupIPTables(&conf)
	if err != nil {
		return fmt.Errorf("unable to configure the iptables: %w", err)
	}

	// Apply the initial config to the interface
	err = initInterfaceConfig(&conf)
	if err != nil {
		return fmt.Errorf("unable to initialize the config: %w", err)
	}

	err = applyInterfaceConfig(&conf)
	if err != nil {
		return fmt.Errorf("unable to apply the initial config: %w", err)
	}

	// Set the interface to be up
	err = setInterfaceUp()
	if err != nil {
		return fmt.Errorf("unable to set the interface to be up: %w", err)
	}

	log15.Info("initialization complete, starting the config monitoring")

	// Monitor the config for changes and apply them
	err = monitorConfig(&conf)
	if err != nil {
		return fmt.Errorf("an error occured while watching for interface changes: %w", err)
	}

	return nil
}

func loadModule() error {
	isLoaded, err := isWireguardLoaded()
	if err != nil {
		return err
	}

	if isLoaded {
		log15.Info("The wireguard module is already loaded")
		return nil
	}

	k, err := kmod.New()
	if err != nil {
		return fmt.Errorf("unable to connect to the kernel's modules interface: %w", err)
	}

	err = k.Load("wireguard", "", 0)
	if err != nil {
		return fmt.Errorf("unable to load the wireguard kernel module: %w", err)
	}

	log15.Info("Loaded the wireguard kernel module")

	return nil
}

func isWireguardLoaded() (bool, error) {
	b, err := os.ReadFile("/proc/modules")
	if err != nil {
		return false, fmt.Errorf("unable to read modules from /proc/modules: %w", err)
	}

	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "wireguard ") {
			return true, nil
		}
	}

	return false, nil
}

func createInterface(conf *Config) error {
	cmd := exec.Command("ip", "link", "add", interfaceName, "type", "wireguard")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to create the interface: %w", err)
	}

	// Turn the subnet specification into the first ip of the subnet
	// This assumes all non masked bits are 0, which the config does at load time
	maskSize, _ := conf.Subnet.Mask.Size()
	ip := conf.Subnet.IP
	ip[len(ip)-1] += 1

	// Print the ip and mask in CIDR notation
	addr := fmt.Sprintf("%v/%v", ip.String(), maskSize)

	cmd = exec.Command("ip", "addr", "add", addr, "dev", interfaceName)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to set an ip on the interface: %w", err)
	}

	return nil
}

func setupIPTables(conf *Config) error {
	cmd := exec.Command("iptables", "-P", "FORWARD", "ACCEPT")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to enable forwarding in the iptables: %w", err)
	}

	// maskSize, _ := conf.Subnet.Mask.Size()
	// addr := fmt.Sprintf("%v/%v", conf.Subnet.IP.String(), maskSize)

	// // Ensure requests to ips on the vpn subnet are masqueraded, for the forwarding to work properly.
	// cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-d", addr, "-j", "MASQUERADE")
	// err = cmd.Run()
	// if err != nil {
	// 	return fmt.Errorf("unable to enable masquerading in the iptables: %w", err)
	// }

	return nil
}

//go:embed default.conf
var defaultConfig string

func initInterfaceConfig(conf *Config) error {
	_, err := os.Stat(conf.InterfaceConfigPath)
	if err == nil {
		// The config file already exists
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to access the config file at %v: %w", conf.InterfaceConfigPath, err)
	}

	// generate a private key
	cmd := exec.Command("wg", "genkey")
	keyBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("unable to generate a new private key: %w", err)
	}

	interfaceConfig := fmt.Sprintf(defaultConfig, string(keyBytes))

	// Write the default config to disk
	err = os.WriteFile(conf.InterfaceConfigPath, []byte(interfaceConfig), 0755)
	if err != nil {
		return fmt.Errorf("unable to write the default config to the file at %v: %w", conf.InterfaceConfigPath, err)
	}

	log15.Info("Generated a new server config.")

	return nil
}

// applyInterfaceConfig reads the current configuration from disk and applies it to the interface
func applyInterfaceConfig(conf *Config) error {
	cmd := exec.Command("wg-quick", "strip", conf.InterfaceConfigPath)
	outp, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("unable to strip the wireguard config: %w", err)
	}

	const tmpPath = "/tmp/wg0.conf"
	os.WriteFile(tmpPath, outp, 0755)

	cmd = exec.Command("wg", "syncconf", interfaceName, tmpPath)
	outp, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to sync the wireguard config: %v: %w", string(outp), err)
	}

	return nil
}

func setInterfaceUp() error {
	cmd := exec.Command("ip", "link", "set", interfaceName, "up")
	outp, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to set the interface up: %v - %w", string(outp), err)
	}

	return nil
}

func monitorConfig(conf *Config) error {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("unable to initialize the file watching: %w", err)
	}

	err = watcher.Add(conf.InterfaceConfigPath)
	if err != nil {
		return fmt.Errorf("cannot watch %v: %w", conf.InterfaceConfigPath, err)
	}

	for {
		event := <-watcher.Events

		if event.Has(fsnotify.Write) {
			log15.Info("Updating the wireguard interface config.")
			err = applyInterfaceConfig(conf)
			if err != nil {
				return fmt.Errorf("unable to update the config on the interface: %w")
			}
		}

		// In case the tool that changed the file did some funky stuff like a write and rename to avoid partial writes
		err = watcher.Add(conf.InterfaceConfigPath)
		if err != nil {
			return fmt.Errorf("cannot watch %v: %w", conf.InterfaceConfigPath, err)
		}
	}
}
