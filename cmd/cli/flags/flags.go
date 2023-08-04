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

func NewStringFlag(command *cobra.Command, name, value, usage string) {
	var s string
	command.PersistentFlags().StringVar(&s, name, value, usage)
	if err := viper.BindPFlag(name, command.PersistentFlags().Lookup(name)); err != nil {
		log.Fatalln("Error: unable to bind flag:", err)
	}
}

func NewBoolFlag(command *cobra.Command, name, shorthand string, value bool, usage string) {
	var b bool
	command.PersistentFlags().BoolVarP(&b, name, shorthand, value, usage)
	if err := viper.BindPFlag(name, command.PersistentFlags().Lookup(name)); err != nil {
		log.Fatalln("Error: unable to bind flag:", err)
	}
}
