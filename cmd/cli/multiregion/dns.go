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
	cfClient   *cloudflare.API
	cfZone     *cloudflare.ResourceContainer
	domainName string
	tfcOrg     string
	tfcToken   string
	testMode   bool
}

type AlbDnsValues struct {
	internal string
	external string
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

	d := newDnsCommand(pFlags)

	var clusterWorkspaceName string
	if failback {
		clusterWorkspaceName = clusterWorkspace(pFlags)
	} else {
		clusterWorkspaceName = clusterSecondaryWorkspace(pFlags)
	}

	dnsValues := d.getAlbDnsValuesFromTfc(clusterWorkspaceName)

	d.setDnsRecordValues(pFlags.idp, dnsValues, failback)
}

func newDnsCommand(pFlags PersistentFlags) *DnsCommand {
	d := DnsCommand{
		testMode:   pFlags.readOnlyMode,
		domainName: viper.GetString("domain-name"),
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
	d.tfcOrg = pFlags.org

	return &d
}

func (d *DnsCommand) setDnsRecordValues(idpKey string, dnsValues AlbDnsValues, failback bool) {
	if failback {
		fmt.Println("Setting DNS records to primary region...")
	} else {
		fmt.Println("Setting DNS records to secondary region...")
	}

	dnsRecords := []struct {
		name         string
		optionValue  string
		defaultValue string
	}{
		// "mfa-api" is the TOTP API, also known as serverless-mfa-api
		{"mfa-api", "mfa-api-value", ""},

		// "twosv-api" is the Webauthn API, also known as serverless-mfa-api-go
		{"twosv-api", "twosv-api-value", ""},

		// "support-bot" is the idp-support-bot API that is configured in the Slack API dashboard
		{"sherlock", "support-bot-value", ""},

		// ECS services
		{idpKey + "-email", "email-service-value", dnsValues.internal},
		{idpKey + "-broker", "id-broker-value", dnsValues.internal},
		{idpKey + "-pw-api", "pw-api-value", dnsValues.external},
		{idpKey, "ssp-value", dnsValues.external},
		{idpKey + "-sync", "id-sync-value", dnsValues.external},
	}

	for _, record := range dnsRecords {
		value := getOption(record.optionValue, record.defaultValue)
		d.setCloudflareCname(record.name, value)
	}
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

func (d *DnsCommand) getAlbDnsValuesFromTfc(workspaceName string) (values AlbDnsValues) {
	config := &tfe.Config{
		Token:             d.tfcToken,
		RetryServerErrors: true,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		fmt.Printf("Error creating Terraform client: %s", err)
		return
	}

	ctx := context.Background()

	w, err := client.Workspaces.Read(ctx, d.tfcOrg, workspaceName)
	if err != nil {
		fmt.Printf("Error reading Terraform workspace %s: %s", workspaceName, err)
		return
	}

	outputs, err := client.StateVersionOutputs.ReadCurrent(ctx, w.ID)
	if err != nil {
		fmt.Printf("Error reading Terraform state outputs on workspace %s: %s", workspaceName, err)
		return
	}

	for _, item := range outputs.Items {
		itemValue, _ := item.Value.(string)
		switch item.Name {
		case "alb_dns_name":
			values.external = itemValue
		case "internal_alb_dns_name":
			values.internal = itemValue
		}
	}
	return
}
