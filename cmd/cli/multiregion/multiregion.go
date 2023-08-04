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

	flags.NewStringFlag(multiregionCmd, flags.DomainName, "", "Domain name")
	flags.NewStringFlag(multiregionCmd, flags.Env, envProd, "Execution environment")
	flags.NewStringFlag(multiregionCmd, flags.Region2, "", "Secondary AWS region")
	flags.NewStringFlag(multiregionCmd, flags.TfcToken, "", "Token for Terraform Cloud authentication")
	flags.NewStringFlag(multiregionCmd, flags.OrgAlternate, "", "Alternate Terraform Cloud organization")
	flags.NewStringFlag(multiregionCmd, flags.TfcTokenAlternate, "", "Alternate token for Terraform Cloud")
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
		idp:             getRequiredParam(flags.Idp),
		org:             getRequiredParam(flags.Org),
		tfcToken:        getRequiredParam(flags.TfcToken),
		secondaryRegion: getRequiredParam(flags.Region2),
		readOnlyMode:    viper.GetBool(flags.ReadOnlyMode),
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
