/*
Copyright © 2023 SIL International
*/

package multiregion

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/silinternational/idp-cli/cmd/cli/flags"
)

const envProd = "prod"

func SetupMultiregionCmd(parentCommand *cobra.Command) {
	multiregionCmd := &cobra.Command{
		Use:   "multiregion",
		Short: "Tools for multiregion operation",
		Long:  `Tools for multiregion setup, operation, fail-over, and fail-back`,
	}

	parentCommand.AddCommand(multiregionCmd)
	InitDnsCmd(multiregionCmd)
	InitFailoverCmd(multiregionCmd)
	InitSetupCmd(multiregionCmd)
	InitStatusCmd(multiregionCmd)

	var domainName string
	multiregionCmd.PersistentFlags().StringVar(&domainName, flags.DomainName, "", "Domain name")
	if err := viper.BindPFlag(flags.DomainName, multiregionCmd.PersistentFlags().Lookup(flags.DomainName)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var env string
	multiregionCmd.PersistentFlags().StringVar(&env, flags.Env, envProd, "Execution environment")
	if err := viper.BindPFlag(flags.Env, multiregionCmd.PersistentFlags().Lookup(flags.Env)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var region2 string
	multiregionCmd.PersistentFlags().StringVar(&region2, flags.Region2, "", "Secondary AWS region")
	if err := viper.BindPFlag(flags.Region2, multiregionCmd.PersistentFlags().Lookup(flags.Region2)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var tfcToken string
	multiregionCmd.PersistentFlags().StringVar(&tfcToken, flags.TfcToken, "", "Token for Terraform Cloud authentication")
	if err := viper.BindPFlag(flags.TfcToken, multiregionCmd.PersistentFlags().Lookup(flags.TfcToken)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var orgAlt string
	multiregionCmd.PersistentFlags().StringVar(&orgAlt, flags.OrgAlternate, "", "Alternate Terraform Cloud organization")
	if err := viper.BindPFlag(flags.OrgAlternate, multiregionCmd.PersistentFlags().Lookup(flags.OrgAlternate)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var tfcTokenAlt string
	multiregionCmd.PersistentFlags().StringVar(&tfcTokenAlt, flags.TfcTokenAlternate, "", "Alternate token for Terraform Cloud")
	if err := viper.BindPFlag(flags.TfcTokenAlternate, multiregionCmd.PersistentFlags().Lookup(flags.TfcTokenAlternate)); err != nil {
		outputFlagError(multiregionCmd, err)
	}
}

func outputFlagError(cmd *cobra.Command, err error) {
	cmd.Help()
	log.Fatalln("Error: unable to bind flag:", err)
}

type PersistentFlags struct {
	env             string
	idp             string
	org             string
	orgAlt          string
	readOnlyMode    bool
	secondaryRegion string
	tfcToken        string
	tfcTokenAlt     string
}

func getPersistentFlags() PersistentFlags {
	pFlags := PersistentFlags{
		env:             getRequiredParam(flags.Env),
		idp:             getRequiredParam("idp"),
		org:             getRequiredParam("org"),
		tfcToken:        getRequiredParam(flags.TfcToken),
		secondaryRegion: getRequiredParam(flags.Region2),
		readOnlyMode:    viper.GetBool("read-only-mode"),
		tfcTokenAlt:     getOption(flags.TfcTokenAlternate, ""),
		orgAlt:          getOption(flags.OrgAlternate, viper.GetString(flags.OrgAlternate)),
	}

	if pFlags.orgAlt != "" && pFlags.tfcTokenAlt == "" {
		log.Fatalf("%[1]s was specified without %[2]s. Please include %[2]s or remove %[1]s.",
			flags.OrgAlternate, flags.TfcTokenAlternate)
	}

	if pFlags.orgAlt == "" {
		pFlags.orgAlt = pFlags.org
	}

	return pFlags
}

func getRequiredParam(key string) string {
	value := viper.GetString(key)

	if value == "" {
		log.Fatalf("parameter %[1]s is not set, use --%[1]s on command line or include in idp-cli.toml file", key)
	}
	return value
}

func getOption(key, defaultValue string) string {
	value := viper.GetString(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

func coreWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-000-core", pFlags.idp, pFlags.env)
}

func clusterWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-010-cluster", pFlags.idp, pFlags.env)
}

func clusterSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-010-cluster-secondary", pFlags.idp, pFlags.env)
}

func databaseWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-020-database", pFlags.idp, pFlags.env)
}

func databaseSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-020-database-secondary", pFlags.idp, pFlags.env)
}

func ecrWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-022-ecr", pFlags.idp, pFlags.env)
}

func pmaWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-030-phpmyadmin", pFlags.idp, pFlags.env)
}

func pmaSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-030-phpmyadmin-secondary", pFlags.idp, pFlags.env)
}

func emailWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-031-email-service", pFlags.idp, pFlags.env)
}

func emailSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-031-email-service-secondary", pFlags.idp, pFlags.env)
}

func backupWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-032-db-backup", pFlags.idp, pFlags.env)
}

func brokerWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-040-id-broker", pFlags.idp, pFlags.env)
}

func brokerSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-040-id-broker-secondary", pFlags.idp, pFlags.env)
}

func searchWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-041-id-broker-search", pFlags.idp, pFlags.env)
}

func pwWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-050-pw-manager", pFlags.idp, pFlags.env)
}

func pwSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-050-pw-manager-secondary", pFlags.idp, pFlags.env)
}

func sspWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-060-simplesamlphp", pFlags.idp, pFlags.env)
}

func sspSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-060-simplesamlphp-secondary", pFlags.idp, pFlags.env)
}

func syncWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-070-id-sync", pFlags.idp, pFlags.env)
}

func syncSecondaryWorkspace(pFlags PersistentFlags) string {
	return fmt.Sprintf("idp-%s-%s-070-id-sync-secondary", pFlags.idp, pFlags.env)
}
