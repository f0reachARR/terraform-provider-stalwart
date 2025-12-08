package provider

import (
	"context"
	"os"

	"github.com/f0reachARR/terraform-provider-stalwart/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &stalwartProvider{}
)

// stalwartProvider is the provider implementation.
type stalwartProvider struct {
	version string
}

// stalwartProviderModel describes the provider data model.
type stalwartProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Metadata returns the provider type name.
func (p *stalwartProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "stalwart"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *stalwartProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Stalwart Email Server API.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The API endpoint for Stalwart server. May also be provided via STALWART_ENDPOINT environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for Stalwart API authentication. May also be provided via STALWART_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Stalwart API authentication. May also be provided via STALWART_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Stalwart API client for data sources and resources.
func (p *stalwartProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config stalwartProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Stalwart API Endpoint",
			"The provider cannot create the Stalwart API client as there is an unknown configuration value for the Stalwart API endpoint.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Stalwart API Username",
			"The provider cannot create the Stalwart API client as there is an unknown configuration value for the Stalwart API username.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Stalwart API Password",
			"The provider cannot create the Stalwart API client as there is an unknown configuration value for the Stalwart API password.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("STALWART_ENDPOINT")
	username := os.Getenv("STALWART_USERNAME")
	password := os.Getenv("STALWART_PASSWORD")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Stalwart API Endpoint",
			"The provider cannot create the Stalwart API client as there is a missing or empty value for the Stalwart API endpoint.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing Stalwart API Username",
			"The provider cannot create the Stalwart API client as there is a missing or empty value for the Stalwart API username.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Missing Stalwart API Password",
			"The provider cannot create the Stalwart API client as there is a missing or empty value for the Stalwart API password.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create authenticated client using generated code
	apiClient, err := client.NewAuthenticatedClient(endpoint, username, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Stalwart API Client",
			"An unexpected error occurred when creating the Stalwart API client. "+
				"Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

// DataSources defines the data sources implemented in the provider.
func (p *stalwartProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPrincipalsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *stalwartProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPrincipalResource,
		NewDKIMSignatureResource,
	}
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &stalwartProvider{
			version: version,
		}
	}
}
