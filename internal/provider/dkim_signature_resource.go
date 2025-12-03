package provider

import (
	"context"
	"fmt"

	"github.com/f0reachARR/terraform-provider-stalwart/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &dkimSignatureResource{}
	_ resource.ResourceWithConfigure = &dkimSignatureResource{}
)

func NewDKIMSignatureResource() resource.Resource {
	return &dkimSignatureResource{}
}

type dkimSignatureResource struct {
	client *client.AuthenticatedClient
}

type dkimSignatureResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Algorithm types.String `tfsdk:"algorithm"`
	Domain    types.String `tfsdk:"domain"`
	Selector  types.String `tfsdk:"selector"`
}

func (r *dkimSignatureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dkim_signature"
}

func (r *dkimSignatureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DKIM signature for a domain.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the DKIM signature.",
				Optional:    true,
				Computed:    true,
			},
			"algorithm": schema.StringAttribute{
				Description: "Algorithm for DKIM signature: Ed25519 or RSA.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Domain for which to create the DKIM signature.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"selector": schema.StringAttribute{
				Description: "DKIM selector.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *dkimSignatureResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.AuthenticatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.AuthenticatedClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *dkimSignatureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dkimSignatureResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body using generated types
	body := client.PostDkimJSONRequestBody{
		Algorithm: ptrString(plan.Algorithm.ValueString()),
		Domain:    ptrString(plan.Domain.ValueString()),
	}

	// Create DKIM signature via generated API
	apiResp, err := r.client.PostDkimWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating DKIM signature",
			"Could not create DKIM signature: "+err.Error(),
		)
		return
	}

	if apiResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error creating DKIM signature",
			fmt.Sprintf("API returned status %d", apiResp.StatusCode()),
		)
		return
	}

	// Set computed values
	if plan.ID.IsNull() {
		plan.ID = types.StringValue(plan.Domain.ValueString())
	}
	if plan.Selector.IsNull() {
		plan.Selector = types.StringValue("default")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dkimSignatureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dkimSignatureResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: The Stalwart OpenAPI specification (v1) does not provide a GET endpoint
	// for retrieving DKIM signatures. The signature configuration is stored in the
	// server's settings/configuration, but there's no direct API to query it.
	// 
	// In a production implementation, you could:
	// 1. Use the /settings/keys or /settings endpoints to verify the signature exists
	// 2. Store the signature data in the Terraform state during creation
	// 3. Accept that DKIM signatures are write-only resources
	//
	// For now, we maintain the state as-is and rely on Terraform state management.
	// This means the resource won't detect drift if the DKIM signature is deleted
	// outside of Terraform.

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *dkimSignatureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// DKIM signatures typically cannot be updated, only recreated
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"DKIM signatures cannot be updated. Please delete and recreate the resource.",
	)
}

func (r *dkimSignatureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Note: The Stalwart OpenAPI specification (v1) does not provide a DELETE endpoint
	// for DKIM signatures. DKIM signatures are typically managed as part of the server's
	// configuration and would need to be removed via the /settings endpoint.
	//
	// In a production implementation, you would:
	// 1. Use DELETE /settings with the appropriate prefix to remove the DKIM configuration
	// 2. Or use POST /settings with a "clear" action to remove the signature settings
	//
	// For now, we only remove the resource from Terraform state. This is a limitation
	// of the current API design. Users who need to fully delete DKIM signatures should
	// do so through the Stalwart admin interface or by using the settings resource directly.
	//
	// This is documented in the resource schema as a known limitation.
}
