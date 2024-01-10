package cmd

import (
	"os"

	"github.com/floriankramer/wireguard_gaming/wireguard"
	"github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wireguard_gaming",
	Short: "A wireguard-ui compatible wireguard server configuration tool",
	Long: `This tool creates and manages a wireguard interface with the goal
of simulating a lan for a set of machines. It can be used e.g. for gaming.
Its designed to be compatible with wireguard-ui and supports hot-reloading
of the config when the config file changes using inotify.
https://github.com/ngoduykhanh/wireguard-ui

Environment Variables:
  INTERFACE_CONFIG_PATH: The path to the wireguard interfaces config. Defaults to /etc/wireguard/wg0.conf
	SUBNET: The subnet the server used. The first ip is assigned to the server's interface. Defaults to 10.32.42.0/24`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		println(`

 _      _  ____  _____ _____ _     ____  ____  ____    _____ ____  _      _  _      _____
/ \  /|/ \/  __\/  __//  __// \ /\/  _ \/  __\/  _ \  /  __//  _ \/ \__/|/ \/ \  /|/  __/
| |  ||| ||  \/||  \  | |  _| | ||| / \||  \/|| | \|  | |  _| / \|| |\/||| || |\ ||| |  _
| |/\||| ||    /|  /_ | |_//| \_/|| |-|||    /| |_/|  | |_//| |-||| |  ||| || | \||| |_//
\_/  \|\_/\_/\_\\____\\____\\____/\_/ \|\_/\_\\____/  \____\\_/ \|\_/  \|\_/\_/  \|\____\
                                                                                         

`)

		err := wireguard.RunServer()
		if err != nil {
			log15.Error("A critical error occurred while running the wireguard server", "erro", err)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wireguard_gaming.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
