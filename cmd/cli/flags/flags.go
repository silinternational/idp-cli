package flags

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
