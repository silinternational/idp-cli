/*
Copyright © 2023 SIL International
*/
package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func SetupVersionCmd(parentCommand *cobra.Command) {
	parentCommand.AddCommand(versionCmd)
}

// version is set at build time
var version = ""

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show idp-cli version",
	Run: func(cmd *cobra.Command, args []string) {
		if version != "" {
			fmt.Printf("Version %s\n", version)
		}

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			fmt.Println("Version:", buildInfo.Main.Version)
			return
		}

		fmt.Println("Version: unknown")
	},
}
