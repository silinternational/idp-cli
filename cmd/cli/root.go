/*
Copyright © 2023 SIL International
*/
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/silinternational/idp-cli/cmd/cli/flags"
	"github.com/silinternational/idp-cli/cmd/cli/multiregion"
)

const requiredPrefix = "required - "

var configFile string

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "idp-cli",
		Short: "idp-in-a-box CLI",
		Long: `idp is a CLI tool for the silinternational/idp-in-a-box system.
It can be used to check the status of the IdP. It can also be used to establish secondary resources
in a second AWS region, and to initiate a secondary region failover action.`,
	}

	rootCmd.PersistentFlags().StringVar(&configFile, flags.Config, "", "Config file")

	var org string
	rootCmd.PersistentFlags().StringVar(&org, flags.Org, "", requiredPrefix+"Terraform Cloud organization")
	if err := viper.BindPFlag(flags.Org, rootCmd.PersistentFlags().Lookup(flags.Org)); err != nil {
		log.Fatalln("Error: unable to bind flag:", err)
	}

	var idp string
	rootCmd.PersistentFlags().StringVar(&idp, flags.Idp, "", requiredPrefix+"IDP key (short name)")
	if err := viper.BindPFlag(flags.Idp, rootCmd.PersistentFlags().Lookup(flags.Idp)); err != nil {
		log.Fatalln("Error: unable to bind flag:", err)
	}

	var readOnly bool
	rootCmd.PersistentFlags().BoolVarP(&readOnly, flags.ReadOnlyMode, "r", false,
		`read-only mode persists no changes`,
	)
	if err := viper.BindPFlag(flags.ReadOnlyMode, rootCmd.PersistentFlags().Lookup(flags.ReadOnlyMode)); err != nil {
		log.Fatalln("Error: unable to bind flag:", err)
	}

	SetupVersionCmd(rootCmd)
	multiregion.SetupMultiregionCmd(rootCmd)

	cobra.OnInitialize(initConfig)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// initConfig reads in a Config file and ENV variables if set.
func initConfig() {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory and add ~/.config/ to the config search path
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home + "/.config")

		// Add the current directory to the search path
		viper.AddConfigPath(".")

		// Set the config file name to "idp", allowing any supported file extension, e.g. idp-cli.toml
		viper.SetConfigName("idp-cli")
	}

	// look for environment variables that match the uppercase of the viper key, prefixed with "IDP_"
	viper.SetEnvPrefix("IDP")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_, err = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		if err != nil {
			panic(err.Error())
		}
	} else {
		var vErr viper.ConfigFileNotFoundError
		if !errors.As(err, &vErr) {
			_, err = fmt.Fprintf(os.Stderr, "Problem with config file: %s %T\n", err, err)
			if err != nil {
				panic(err.Error())
			}
		}
	}
}
