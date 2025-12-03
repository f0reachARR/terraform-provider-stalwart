package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/f0reachARR/terraform-provider-stalwart/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &principalResource{}
	_ resource.ResourceWithConfigure   = &principalResource{}
	_ resource.ResourceWithImportState = &principalResource{}
)

func NewPrincipalResource() resource.Resource {
	return &principalResource{}
}

type principalResource struct {
	client *client.AuthenticatedClient
}

type principalResourceModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	Type                types.String `tfsdk:"type"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Quota               types.Int64  `tfsdk:"quota"`
	Secrets             types.List   `tfsdk:"secrets"`
	Emails              types.List   `tfsdk:"emails"`
	URLs                types.List   `tfsdk:"urls"`
	MemberOf            types.List   `tfsdk:"member_of"`
	Roles               types.List   `tfsdk:"roles"`
	Lists               types.List   `tfsdk:"lists"`
	Members             types.List   `tfsdk:"members"`
	EnabledPermissions  types.List   `tfsdk:"enabled_permissions"`
	DisabledPermissions types.List   `tfsdk:"disabled_permissions"`
	ExternalMembers     types.List   `tfsdk:"external_members"`
}

func (r *principalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_principal"
}

func (r *principalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Stalwart principal (user, group, domain, or list).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the principal.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of principal: individual, group, domain, or list.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the principal.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the principal.",
				Optional:    true,
			},
			"quota": schema.Int64Attribute{
				Description: "Storage quota for the principal in bytes.",
				Optional:    true,
			},
			"secrets": schema.ListAttribute{
				Description: "List of password hashes for authentication.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"emails": schema.ListAttribute{
				Description: "List of email addresses associated with the principal.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"urls": schema.ListAttribute{
				Description: "List of URLs associated with the principal.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"member_of": schema.ListAttribute{
				Description: "List of groups this principal is a member of.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"roles": schema.ListAttribute{
				Description: "List of roles assigned to the principal.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"lists": schema.ListAttribute{
				Description: "List of mailing lists this principal belongs to.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"members": schema.ListAttribute{
				Description: "List of members (for group or list principals).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"enabled_permissions": schema.ListAttribute{
				Description: "List of enabled permissions for the principal.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"disabled_permissions": schema.ListAttribute{
				Description: "List of disabled permissions for the principal.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"external_members": schema.ListAttribute{
				Description: "List of external members (for list principals).",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *principalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *principalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan principalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body using generated types
	body := client.PostPrincipalJSONRequestBody{
		Type:        ptrString(plan.Type.ValueString()),
		Name:        ptrString(plan.Name.ValueString()),
		Description: ptrString(plan.Description.ValueString()),
	}

	if !plan.Quota.IsNull() {
		quota := float32(plan.Quota.ValueInt64())
		body.Quota = &quota
	}

	// Convert lists to interface arrays as expected by generated code
	if !plan.Secrets.IsNull() {
		var secrets []string
		diags = plan.Secrets.ElementsAs(ctx, &secrets, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		secretsInterface := make([]interface{}, len(secrets))
		for i, s := range secrets {
			secretsInterface[i] = s
		}
		body.Secrets = &secretsInterface
	}

	if !plan.Emails.IsNull() {
		var emails []string
		diags = plan.Emails.ElementsAs(ctx, &emails, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		emailsInterface := make([]interface{}, len(emails))
		for i, e := range emails {
			emailsInterface[i] = e
		}
		body.Emails = &emailsInterface
	}

	if !plan.Roles.IsNull() {
		var roles []string
		diags = plan.Roles.ElementsAs(ctx, &roles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		rolesInterface := make([]interface{}, len(roles))
		for i, r := range roles {
			rolesInterface[i] = r
		}
		body.Roles = &rolesInterface
	}

	// Create principal via generated API
	apiResp, err := r.client.PostPrincipalWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating principal",
			"Could not create principal: "+err.Error(),
		)
		return
	}

	if apiResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error creating principal",
			fmt.Sprintf("API returned status %d", apiResp.StatusCode()),
		)
		return
	}

	// Extract ID from response
	if apiResp.JSON200 != nil && apiResp.JSON200.Data != nil {
		id := int64(*apiResp.JSON200.Data)
		plan.ID = types.Int64Value(id)
	} else {
		resp.Diagnostics.AddError(
			"Error parsing response",
			"Could not extract principal ID from response",
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *principalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state principalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get principal using generated API
	principalID := fmt.Sprintf("%d", state.ID.ValueInt64())
	apiResp, err := r.client.GetPrincipalPrincipalIdWithResponse(ctx, principalID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading principal",
			"Could not read principal: "+err.Error(),
		)
		return
	}

	if apiResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error reading principal",
			fmt.Sprintf("API returned status %d", apiResp.StatusCode()),
		)
		return
	}

	// Update state from response
	if apiResp.JSON200 != nil && apiResp.JSON200.Data != nil {
		data := apiResp.JSON200.Data
		
		if data.Type != nil {
			state.Type = types.StringValue(*data.Type)
		}
		
		if data.Name != nil {
			state.Name = types.StringValue(*data.Name)
		}
		
		if data.Description != nil {
			state.Description = types.StringValue(*data.Description)
		}

		if data.Quota != nil {
			state.Quota = types.Int64Value(int64(*data.Quota))
		}

		// Handle emails - According to the OpenAPI spec, the emails field is returned as a string
		// (not an array as one might expect). This appears to be how the API represents the data.
		// We convert it to a list for consistency with the Terraform resource schema.
		if data.Emails != nil && *data.Emails != "" {
			// Treat as a single email string from the API
			emailsList, diags := types.ListValueFrom(ctx, types.StringType, []string{*data.Emails})
			resp.Diagnostics.Append(diags...)
			state.Emails = emailsList
		} else {
			state.Emails = types.ListNull(types.StringType)
		}

		if data.Roles != nil && len(*data.Roles) > 0 {
			rolesList, diags := types.ListValueFrom(ctx, types.StringType, *data.Roles)
			resp.Diagnostics.Append(diags...)
			state.Roles = rolesList
		} else {
			state.Roles = types.ListNull(types.StringType)
		}

		if data.Lists != nil && len(*data.Lists) > 0 {
			listsList, diags := types.ListValueFrom(ctx, types.StringType, *data.Lists)
			resp.Diagnostics.Append(diags...)
			state.Lists = listsList
		} else {
			state.Lists = types.ListNull(types.StringType)
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *principalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan principalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state principalResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update operations using generated types
	var updates client.PatchPrincipalPrincipalIdJSONRequestBody

	if !plan.Name.Equal(state.Name) {
		updates = append(updates, struct {
			Action *string `json:"action,omitempty"`
			Field  *string `json:"field,omitempty"`
			Value  *string `json:"value,omitempty"`
		}{
			Action: ptrString("set"),
			Field:  ptrString("name"),
			Value:  ptrString(plan.Name.ValueString()),
		})
	}

	if !plan.Description.Equal(state.Description) {
		updates = append(updates, struct {
			Action *string `json:"action,omitempty"`
			Field  *string `json:"field,omitempty"`
			Value  *string `json:"value,omitempty"`
		}{
			Action: ptrString("set"),
			Field:  ptrString("description"),
			Value:  ptrString(plan.Description.ValueString()),
		})
	}

	if len(updates) > 0 {
		principalID := fmt.Sprintf("%d", plan.ID.ValueInt64())
		apiResp, err := r.client.PatchPrincipalPrincipalIdWithResponse(ctx, principalID, updates)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating principal",
				"Could not update principal: "+err.Error(),
			)
			return
		}

		if apiResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(
				"Error updating principal",
				fmt.Sprintf("API returned status %d", apiResp.StatusCode()),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *principalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state principalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	principalID := fmt.Sprintf("%d", state.ID.ValueInt64())
	apiResp, err := r.client.DeletePrincipalPrincipalIdWithResponse(ctx, principalID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting principal",
			"Could not delete principal: "+err.Error(),
		)
		return
	}

	if apiResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error deleting principal",
			fmt.Sprintf("API returned status %d", apiResp.StatusCode()),
		)
		return
	}
}

func (r *principalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing principal",
			"Could not parse principal ID: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func ptrString(s string) *string {
	return &s
}
