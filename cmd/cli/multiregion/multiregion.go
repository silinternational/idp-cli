/*
Copyright © 2023 SIL International
*/

package multiregion

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagDomainName        = "domain-name"
	flagEnv               = "env"
	flagRegion2           = "region2"
	flagTfcToken          = "tfc-token"
	flagOrgAlternate      = "org-alternate"
	flagTfcTokenAlternate = "tfc-token-alternate"
)

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
	multiregionCmd.PersistentFlags().StringVar(&domainName, flagDomainName, "", "Domain name")
	if err := viper.BindPFlag(flagDomainName, multiregionCmd.PersistentFlags().Lookup(flagDomainName)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var env string
	multiregionCmd.PersistentFlags().StringVar(&env, flagEnv, "prod", "Execution environment")
	if err := viper.BindPFlag(flagEnv, multiregionCmd.PersistentFlags().Lookup(flagEnv)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var region2 string
	multiregionCmd.PersistentFlags().StringVar(&region2, flagRegion2, "", "Secondary AWS region")
	if err := viper.BindPFlag(flagRegion2, multiregionCmd.PersistentFlags().Lookup(flagRegion2)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var tfcToken string
	multiregionCmd.PersistentFlags().StringVar(&tfcToken, flagTfcToken, "", "Token for Terraform Cloud authentication")
	if err := viper.BindPFlag(flagTfcToken, multiregionCmd.PersistentFlags().Lookup(flagTfcToken)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var orgAlt string
	multiregionCmd.PersistentFlags().StringVar(&orgAlt, flagOrgAlternate, "", "Alternate Terraform Cloud organization")
	if err := viper.BindPFlag(flagOrgAlternate, multiregionCmd.PersistentFlags().Lookup(flagOrgAlternate)); err != nil {
		outputFlagError(multiregionCmd, err)
	}

	var tfcTokenAlt string
	multiregionCmd.PersistentFlags().StringVar(&tfcTokenAlt, flagTfcTokenAlternate, "", "Alternate token for Terraform Cloud")
	if err := viper.BindPFlag(flagTfcTokenAlternate, multiregionCmd.PersistentFlags().Lookup(flagTfcTokenAlternate)); err != nil {
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
		env:             getRequiredParam(flagEnv),
		idp:             getRequiredParam("idp"),
		org:             getRequiredParam("org"),
		tfcToken:        getRequiredParam(flagTfcToken),
		secondaryRegion: getRequiredParam(flagRegion2),
		readOnlyMode:    viper.GetBool("read-only-mode"),
		tfcTokenAlt:     getOption(flagTfcTokenAlternate, ""),
		orgAlt:          getOption(flagOrgAlternate, viper.GetString("org")),
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
