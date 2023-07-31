/*
Copyright © 2023 SIL International
*/

package multiregion

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DnsCommand struct {
	cfClient    *cloudflare.API
	cfZone      *cloudflare.ResourceContainer
	domainName  string
	failback    bool
	tfcOrg      string
	tfcOrgAlt   string
	tfcToken    string
	tfcTokenAlt string
	testMode    bool
}

type DnsValues struct {
	albInternal string
	albExternal string
	bot         string
	mfa         string
	twosv       string
}

func InitDnsCmd(parentCmd *cobra.Command) {
	var failback bool

	cmd := &cobra.Command{
		Use:   "dns",
		Short: "DNS Failover and Failback",
		Long:  `Configure DNS CNAME values for primary or secondary region hostnames. Default is failover, use --failback to switch back to the primary region.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDnsCommand(failback)
		},
	}
	parentCmd.AddCommand(cmd)

	cmd.PersistentFlags().BoolVar(&failback, "failback", false,
		`set DNS records to switch back to primary`,
	)
}

func runDnsCommand(failback bool) {
	pFlags := getPersistentFlags()

	d := newDnsCommand(pFlags, failback)

	values := d.getDnsValuesFromTfc(pFlags)
	d.setDnsRecordValues(pFlags.idp, values)
}

func newDnsCommand(pFlags PersistentFlags, failback bool) *DnsCommand {
	d := DnsCommand{
		testMode:   pFlags.readOnlyMode,
		domainName: viper.GetString(flagDomainName),
	}

	if d.domainName == "" {
		log.Fatalln("Cloudflare Domain Name is not configured. Use 'domain-name' parameter.")
	}

	cfToken := viper.GetString("cloudflare-token")
	if cfToken == "" {
		log.Fatalln("Cloudflare Token is not configured. Use 'cloudflare-token' parameter.")
	}

	api, err := cloudflare.NewWithAPIToken(cfToken)
	if err != nil {
		log.Fatal("failed to initialize the Cloudflare API:", err)
	}
	d.cfClient = api

	zoneID, err := d.cfClient.ZoneIDByName(d.domainName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Using domain name %s with ID %s\n", d.domainName, zoneID)
	d.cfZone = cloudflare.ZoneIdentifier(zoneID)

	d.tfcToken = pFlags.tfcToken
	d.tfcTokenAlt = pFlags.tfcTokenAlt
	d.tfcOrg = pFlags.org
	d.tfcOrgAlt = pFlags.orgAlt
	d.failback = failback

	return &d
}

func (d *DnsCommand) setDnsRecordValues(idpKey string, dnsValues DnsValues) {
	if d.failback {
		fmt.Println("Setting DNS records to primary region...")
	} else {
		fmt.Println("Setting DNS records to secondary region...")
	}

	dnsRecords := []struct {
		name      string
		valueFlag string
		tfcValue  string
	}{
		// "mfa-api" is the TOTP API, also known as serverless-mfa-api
		{"mfa-api", "mfa-api-value", dnsValues.mfa},

		// "twosv-api" is the Webauthn API, also known as serverless-mfa-api-go
		{"twosv-api", "twosv-api-value", dnsValues.twosv},

		// "support-bot" is the idp-support-bot API that is configured in the Slack API dashboard
		{"sherlock", "support-bot-value", dnsValues.bot},

		// ECS services
		{idpKey + "-email", "email-service-value", dnsValues.albInternal},
		{idpKey + "-broker", "id-broker-value", dnsValues.albInternal},
		{idpKey + "-pw-api", "pw-api-value", dnsValues.albExternal},
		{idpKey, "ssp-value", dnsValues.albExternal},
		{idpKey + "-sync", "id-sync-value", dnsValues.albExternal},
	}

	for _, record := range dnsRecords {
		value := getDnsValue(record.valueFlag, record.tfcValue)
		d.setCloudflareCname(record.name, value)
	}
}

func getDnsValue(valueFlag, tfcValue string) string {
	if tfcValue != "" {
		return tfcValue
	}
	return viper.GetString(valueFlag)
}

func (d *DnsCommand) setCloudflareCname(name, value string) {
	if value == "" {
		fmt.Printf("  skipping %s (no value provided)\n", name)
		return
	}

	fmt.Printf("  %s.%s --> %s\n", name, d.domainName, value)

	ctx := context.Background()

	r, _, err := d.cfClient.ListDNSRecords(ctx, d.cfZone, cloudflare.ListDNSRecordsParams{Name: name + "." + d.domainName})
	if err != nil {
		log.Fatalf("error finding DNS record %s: %s", name, err)
	}
	if len(r) != 1 {
		log.Fatalf("did not find DNS record %s", name)
	}

	if r[0].Content == value {
		fmt.Printf("CNAME %s is already set to %s\n", name, value)
		return
	}

	if d.testMode {
		fmt.Println("  read-only mode: skipping API call")
		return
	}

	answer := simplePrompt(`Type "yes" to set this DNS record`)
	if answer != "yes" {
		return
	}

	_, err = d.cfClient.UpdateDNSRecord(ctx, d.cfZone, cloudflare.UpdateDNSRecordParams{
		ID:      r[0].ID,
		Type:    "CNAME",
		Name:    name,
		Content: value,
	})
	if err != nil {
		log.Fatalf("error updating DNS record %s: %s", name, err)
	}
}

func (d *DnsCommand) getDnsValuesFromTfc(pFlags PersistentFlags) (values DnsValues) {
	ctx := context.Background()

	var clusterWorkspaceName string
	if d.failback {
		clusterWorkspaceName = clusterWorkspace(pFlags)
	} else {
		clusterWorkspaceName = clusterSecondaryWorkspace(pFlags)
	}

	internal, external := d.getAlbValuesFromTfc(ctx, clusterWorkspaceName)
	values.albInternal = internal
	values.albExternal = external

	bot := "idp-support-bot-prod"
	if pFlags.env != envProd {
		bot = "idp-support-bot-dev" // TODO: consider renaming the workspace name so this can be simplified
	}
	values.bot = d.getLambdaDnsValueFromTfc(ctx, bot)

	twosv := "serverless-mfa-api-go-prod"
	if pFlags.env != envProd {
		twosv = "serverless-mfa-api-go-dev" // TODO: consider renaming the workspace name so this can be simplified
	}
	values.twosv = d.getLambdaDnsValueFromTfc(ctx, twosv)

	mfa := "serverless-mfa-api-prod"
	if pFlags.env != envProd {
		mfa = "serverless-mfa-api-dev" // TODO: consider renaming the workspace name so this can be simplified
	}
	values.mfa = d.getLambdaDnsValueFromTfc(ctx, mfa)
	return
}

func (d *DnsCommand) getAlbValuesFromTfc(ctx context.Context, workspaceName string) (internal, external string) {
	workspaceID, client, err := d.findTfcWorkspace(ctx, workspaceName)
	if err != nil {
		fmt.Printf("Failed to get ALB DNS values: %s\n  Will use DNS config values if provided.\n", err)
		return
	}

	outputs, err := client.StateVersionOutputs.ReadCurrent(ctx, workspaceID)
	if err != nil {
		fmt.Printf("Error reading Terraform state outputs on workspace %s: %s", workspaceName, err)
		return
	}

	for _, item := range outputs.Items {
		itemValue, _ := item.Value.(string)
		switch item.Name {
		case "alb_dns_name":
			external = itemValue
		case "internal_alb_dns_name":
			internal = itemValue
		}
	}
	return
}

func (d *DnsCommand) getLambdaDnsValueFromTfc(ctx context.Context, workspaceName string) string {
	outputName := "secondary_region_domain_name"
	if d.failback {
		outputName = "primary_region_domain_name"
	}
	return d.getTfcOutputFromWorkspace(ctx, workspaceName, outputName)
}

func (d *DnsCommand) getTfcOutputFromWorkspace(ctx context.Context, workspaceName, outputName string) string {
	workspaceID, client, err := d.findTfcWorkspace(ctx, workspaceName)
	if err != nil {
		fmt.Printf("Failed to get DNS value from %s: %s\n  Will use config value if provided.\n", workspaceName, err)
		return ""
	}

	outputs, err := client.StateVersionOutputs.ReadCurrent(ctx, workspaceID)
	if err != nil {
		fmt.Printf("Error reading Terraform state outputs on workspace %s: %s", workspaceName, err)
		return ""
	}

	for _, item := range outputs.Items {
		if item.Name == outputName {
			if itemValue, ok := item.Value.(string); ok {
				return itemValue
			}
			break
		}
	}

	fmt.Printf("Value for %s not found in %s\n", outputName, workspaceName)
	return ""
}

// findTfcWorkspace looks for a workspace by name in two different Terraform Cloud accounts and returns
// the workspace ID and an API client for the account where the workspace was found
func (d *DnsCommand) findTfcWorkspace(ctx context.Context, workspaceName string) (id string, client *tfe.Client, err error) {
	config := &tfe.Config{
		Token:             d.tfcToken,
		RetryServerErrors: true,
	}

	client, err = tfe.NewClient(config)
	if err != nil {
		err = fmt.Errorf("error creating Terraform client: %s", err)
		return
	}

	w, err := client.Workspaces.Read(ctx, d.tfcOrg, workspaceName)
	if err == nil {
		id = w.ID
		return
	}

	if d.tfcTokenAlt == "" {
		err = fmt.Errorf("error reading Terraform workspace %s: %s", workspaceName, err)
		return
	}

	fmt.Printf("Workspace %s not found using %s, trying %s\n", workspaceName, flagTfcToken, flagTfcTokenAlternate)

	config.Token = d.tfcTokenAlt
	client, err = tfe.NewClient(config)
	if err != nil {
		err = fmt.Errorf("error creating alternate Terraform client: %s", err)
		return
	}

	w, err = client.Workspaces.Read(ctx, d.tfcOrgAlt, workspaceName)
	if err != nil {
		err = fmt.Errorf("error reading Terraform workspace %s using %s: %s", workspaceName, flagTfcTokenAlternate, err)
		return
	}

	id = w.ID
	return
}
