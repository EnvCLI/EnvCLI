package cmd

import (
	"fmt"
	"os"

	"github.com/EnvCLI/EnvCLI/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(getAllCmd)
	configCmd.AddCommand(unsetCmd)
}

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "updates the config",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

var setCmd = &cobra.Command{
	Use: "set",
	Run: func(cmd *cobra.Command, args []string) {
		// Check Parameters
		if len(args) != 2 {
			log.Fatal().Msg("Please provide the variable name and the value you want to set in this format. [envcli config set variable value]")
		}
		varName := args[0]
		varValue := args[1]

		// Set value
		config.SetPropertyConfigEntry(varName, varValue)
		fmt.Printf("Set value of %s to [%s]\n", varName, varValue)
	},
}

var getCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		// Check Parameters
		if len(args) != 1 {
			log.Fatal().Msg("Please provide the variable name you want to read. [envcli config get variable]")
		}
		varName := args[0]

		// Get Value
		fmt.Printf("%s [%s]\n", varName, config.GetPropertyConfigEntry(varName))
	},
}

var getAllCmd = &cobra.Command{
	Use: "get-all",
	Run: func(cmd *cobra.Command, args []string) {
		// Print all values
		for key, value := range propConfig.Properties {
			fmt.Printf("%s [%s]\n", key, value)
		}
	},
}

var unsetCmd = &cobra.Command{
	Use: "unset",
	Run: func(cmd *cobra.Command, args []string) {
		// Check Parameters
		if len(args) != 1 {
			log.Fatal().Msg("Please provide the variable name you want to unset. [envcli config unset variable]")
		}
		varName := args[0]

		// Unset value
		config.UnsetPropertyConfigEntry(varName)
		fmt.Printf("Value of variable %s set to [].\n", varName)
	},
}
