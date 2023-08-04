package flags

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Root-level persistent flags
const (
	Config       = "config"
	Org          = "org"
	Idp          = "idp"
	Region       = "region"
	ReadOnlyMode = "read-only-mode"
)

// Persistent flags for multiregion commands
const (
	DomainName        = "domain-name"
	Env               = "env"
	Region2           = "region2"
	TfcToken          = "tfc-token"
	OrgAlternate      = "org-alternate"
	TfcTokenAlternate = "tfc-token-alternate"
)

func NewStringFlag(command *cobra.Command, name, shorthand string, value, usage string) {
	var s string

	if shorthand == "" {
		command.PersistentFlags().StringVar(&s, name, value, usage)
	} else {
		command.PersistentFlags().StringVarP(&s, name, shorthand, value, usage)
	}
	if err := viper.BindPFlag(name, command.PersistentFlags().Lookup(name)); err != nil {
		log.Fatalln("Error: unable to bind flag:", err)
	}
}

func NewBoolFlag(command *cobra.Command, name, shorthand string, value bool, usage string) {
	var b bool
	if shorthand == "" {
		command.PersistentFlags().BoolVar(&b, name, value, usage)
	} else {
		command.PersistentFlags().BoolVarP(&b, name, shorthand, value, usage)
	}
	if err := viper.BindPFlag(name, command.PersistentFlags().Lookup(name)); err != nil {
		log.Fatalln("Error: unable to bind flag:", err)
	}
}
