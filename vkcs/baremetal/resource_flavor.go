package baremetal

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vk-cs/terraform-provider-vkcs/vkcs/internal/clients"
	"github.com/vk-cs/terraform-provider-vkcs/vkcs/internal/services/baremetal/v1/flavors"
	"github.com/vk-cs/terraform-provider-vkcs/vkcs/internal/util/errutil"
)

var (
	_ resource.Resource                = (*FlavorResource)(nil)
	_ resource.ResourceWithConfigure   = (*FlavorResource)(nil)
	_ resource.ResourceWithImportState = (*FlavorResource)(nil)
)

// NewFlavorResource manages project-specific properties of an existing bare metal flavor.
func NewFlavorResource() resource.Resource {
	return &FlavorResource{}
}

type FlavorResource struct {
	config clients.Config
}

type FlavorResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Region          types.String `tfsdk:"region"`
	Name            types.String `tfsdk:"name"`
	DisplayName     types.String `tfsdk:"display_name"`
	CpuModel        types.String `tfsdk:"cpu_model"`
	CpuCores        types.Int64  `tfsdk:"cpu_cores"`
	RamSize         types.Int64  `tfsdk:"ram_size"`
	SsdSize         types.Int64  `tfsdk:"ssd_size"`
	HddSize         types.Int64  `tfsdk:"hdd_size"`
	BondVlanCapable types.Bool   `tfsdk:"bond_vlan_capable"`
}

func (r *FlavorResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "vkcs_baremetal_flavor"
}

func (r *FlavorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The UUID of the existing flavor.",
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The region to fetch the bare metal flavor from, defaults to the provider's region.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the flavor.",
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "The project-specific display name of the flavor. Destroying the resource clears it.",
			},
			"cpu_model": schema.StringAttribute{
				Computed:    true,
				Description: "The CPU model.",
			},
			"cpu_cores": schema.Int64Attribute{
				Computed:    true,
				Description: "CPU core count including hyper-threading.",
			},
			"ram_size": schema.Int64Attribute{
				Computed:    true,
				Description: "RAM in gigabytes.",
			},
			"ssd_size": schema.Int64Attribute{
				Computed:    true,
				Description: "SSD size in gigabytes.",
			},
			"hdd_size": schema.Int64Attribute{
				Computed:    true,
				Description: "HDD size in gigabytes.",
			},
			"bond_vlan_capable": schema.BoolAttribute{
				Computed:    true,
				Description: "Bond and VLAN capable.",
			},
		},
		Description: "Manages project-specific properties of an existing VKCS bare metal flavor.",
	}
}

func (r *FlavorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.config = req.ProviderData.(clients.Config)
}

func (r *FlavorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FlavorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = r.config.GetRegion()
	}

	client, err := r.config.BareMetalV1Client(region)
	if err != nil {
		resp.Diagnostics.AddError("Error creating VKCS baremetal API client", err.Error())
		return
	}

	if err := setFlavorDisplayName(client, data.ID.ValueString(), data.DisplayName.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating baremetal flavor display name", err.Error())
		return
	}

	flavor, err := flavors.Get(client, data.ID.ValueString()).Extract()
	if err != nil {
		resp.Diagnostics.AddError("Error reading baremetal flavor after update", err.Error())
		return
	}

	setFlavorResourceModel(&data, flavor, region)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FlavorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FlavorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = r.config.GetRegion()
	}

	client, err := r.config.BareMetalV1Client(region)
	if err != nil {
		resp.Diagnostics.AddError("Error creating VKCS baremetal API client", err.Error())
		return
	}

	flavor, err := flavors.Get(client, data.ID.ValueString()).Extract()
	if errutil.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading baremetal flavor", err.Error())
		return
	}

	setFlavorResourceModel(&data, flavor, region)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FlavorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FlavorResourceModel
	var state FlavorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := state.Region.ValueString()
	if region == "" {
		region = r.config.GetRegion()
	}

	client, err := r.config.BareMetalV1Client(region)
	if err != nil {
		resp.Diagnostics.AddError("Error creating VKCS baremetal API client", err.Error())
		return
	}

	if err := setFlavorDisplayName(client, state.ID.ValueString(), plan.DisplayName.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating baremetal flavor display name", err.Error())
		return
	}

	flavor, err := flavors.Get(client, state.ID.ValueString()).Extract()
	if err != nil {
		resp.Diagnostics.AddError("Error reading baremetal flavor after update", err.Error())
		return
	}

	setFlavorResourceModel(&state, flavor, region)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FlavorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FlavorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = r.config.GetRegion()
	}

	client, err := r.config.BareMetalV1Client(region)
	if err != nil {
		resp.Diagnostics.AddError("Error creating VKCS baremetal API client", err.Error())
		return
	}

	if err := setFlavorDisplayName(client, data.ID.ValueString(), ""); errutil.IsNotFound(err) {
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Error clearing baremetal flavor display name", err.Error())
	}
}

func (r *FlavorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setFlavorDisplayName(client *gophercloud.ServiceClient, flavorID, displayName string) error {
	return flavors.Update(client, flavorID, flavors.UpdateOpts{DisplayName: &displayName}).ExtractErr()
}

func setFlavorResourceModel(data *FlavorResourceModel, flavor *flavors.Flavor, region string) {
	data.ID = types.StringValue(flavor.FlavorId)
	data.Region = types.StringValue(region)
	data.Name = types.StringValue(flavor.FlavorName)
	data.DisplayName = types.StringPointerValue(flavor.DisplayName)
	data.CpuModel = types.StringValue(flavor.CpuModel)
	data.CpuCores = types.Int64Value(flavor.CpuCores)
	data.RamSize = types.Int64Value(flavor.RamGb)
	data.BondVlanCapable = types.BoolValue(flavor.BondAndVlanCapable)

	for _, disk := range flavor.Disks {
		switch disk.Type {
		case flavors.DiskTypeSSD:
			data.SsdSize = types.Int64Value(disk.Size)
		case flavors.DiskTypeHDD:
			data.HddSize = types.Int64Value(disk.Size)
		}
	}
}
