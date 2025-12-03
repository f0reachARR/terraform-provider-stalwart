package provider

import (
	"context"
	"fmt"

	"github.com/f0reachARR/terraform-provider-stalwart/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &principalsDataSource{}
	_ datasource.DataSourceWithConfigure = &principalsDataSource{}
)

func NewPrincipalsDataSource() datasource.DataSource {
	return &principalsDataSource{}
}

type principalsDataSource struct {
	client *client.AuthenticatedClient
}

type principalsDataSourceModel struct {
	Types      types.String      `tfsdk:"types"`
	Principals []principalModel  `tfsdk:"principals"`
}

type principalModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Quota       types.Int64  `tfsdk:"quota"`
}

func (d *principalsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_principals"
}

func (d *principalsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of principals from Stalwart server.",
		Attributes: map[string]schema.Attribute{
			"types": schema.StringAttribute{
				Description: "Filter principals by type (comma-separated).",
				Optional:    true,
			},
			"principals": schema.ListNestedAttribute{
				Description: "List of principals.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Principal ID.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Principal type.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Principal name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Principal description.",
							Computed:    true,
						},
						"quota": schema.Int64Attribute{
							Description: "Storage quota.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *principalsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.AuthenticatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.AuthenticatedClient, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *principalsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config principalsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare query parameters
	params := &client.GetPrincipalParams{}
	if !config.Types.IsNull() {
		types := config.Types.ValueString()
		params.Types = &types
	}

	// Get principals using generated API
	apiResp, err := d.client.GetPrincipalWithResponse(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading principals",
			"Could not read principals: "+err.Error(),
		)
		return
	}

	if apiResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error reading principals",
			fmt.Sprintf("API returned status %d", apiResp.StatusCode()),
		)
		return
	}

	// Process response
	if apiResp.JSON200 != nil && apiResp.JSON200.Data != nil && apiResp.JSON200.Data.Items != nil {
		items := *apiResp.JSON200.Data.Items
		principals := make([]principalModel, 0, len(items))
		
		for _, item := range items {
			if p, ok := item.(map[string]interface{}); ok {
				principal := principalModel{}
				
				if id, ok := p["id"].(float64); ok {
					principal.ID = types.Int64Value(int64(id))
				}
				
				if typeVal, ok := p["type"].(string); ok {
					principal.Type = types.StringValue(typeVal)
				}
				
				if name, ok := p["name"].(string); ok {
					principal.Name = types.StringValue(name)
				}
				
				if desc, ok := p["description"].(string); ok {
					principal.Description = types.StringValue(desc)
				} else {
					principal.Description = types.StringValue("")
				}
				
				if quota, ok := p["quota"].(float64); ok {
					principal.Quota = types.Int64Value(int64(quota))
				} else {
					principal.Quota = types.Int64Value(0)
				}
				
				principals = append(principals, principal)
			}
		}
		
		config.Principals = principals
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
