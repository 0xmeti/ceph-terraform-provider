package provider

import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
    _ resource.Resource                = &poolResource{}
    _ resource.ResourceWithConfigure   = &poolResource{}
    _ resource.ResourceWithImportState = &poolResource{}
)

// NewPoolResource is a helper function to simplify the provider implementation
func NewPoolResource() resource.Resource {
    return &poolResource{}
}

// poolResource is the resource implementation
type poolResource struct {
    client *CephClient
}

// Metadata returns the resource type name
func (r *poolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_pool"
}

// Schema defines the schema for the resource
func (r *poolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Manages a Ceph pool.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "Pool identifier",
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Description: "Name of the pool",
                Required:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "pool_type": schema.StringAttribute{
                Description: "Type of the pool (replicated or erasure)",
                Optional:    true,
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "pg_num": schema.Int64Attribute{
                Description: "Number of placement groups",
                Optional:    true,
                Computed:    true,
                PlanModifiers: []planmodifier.Int64{
                    int64planmodifier.UseStateForUnknown(),
                },
            },
            "pgp_num": schema.Int64Attribute{
                Description: "Number of placement groups for placement",
                Optional:    true,
                Computed:    true,
                PlanModifiers: []planmodifier.Int64{
                    int64planmodifier.UseStateForUnknown(),
                },
            },
            "size": schema.Int64Attribute{
                Description: "Replication size",
                Optional:    true,
                Computed:    true,
                PlanModifiers: []planmodifier.Int64{
                    int64planmodifier.UseStateForUnknown(),
                },
            },
            "application": schema.StringAttribute{
                Description: "Pool application (rbd, cephfs, rgw)",
                Optional:    true,
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
        },
    }
}

// Configure adds the provider configured client to the resource
func (r *poolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*CephClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *CephClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    r.client = client
}

// Create creates the resource and sets the initial Terraform state
func (r *poolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan PoolResourceModel

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create the pool
    poolName := plan.Name.ValueString()
    poolType := "replicated"
    if !plan.PoolType.IsNull() {
        poolType = plan.PoolType.ValueString()
    }

    pgNum := int64(32)
    if !plan.PgNum.IsNull() {
        pgNum = plan.PgNum.ValueInt64()
    }
    
    pgpNum := pgNum
    if !plan.PgpNum.IsNull() {
        pgpNum = plan.PgpNum.ValueInt64()
    }

    size := int64(3)
    if !plan.Size.IsNull() {
        size = plan.Size.ValueInt64()
    }

    application := ""
    if !plan.Application.IsNull() {
        application = plan.Application.ValueString()
    }

    // Create pool request
    poolReq := PoolCreateRequest{
        Pool:        poolName,
        PoolType:    poolType,
        PgNum:       int(pgNum),
        PgpNum:      int(pgpNum),
        Size:        int(size),
        Application: application,
    }

    err := r.client.CreatePool(poolReq)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating Ceph Pool",
            fmt.Sprintf("Could not create pool %s: %s", poolName, err.Error()),
        )
        return
    }

    // Set application if provided and not already set during pool creation
    if application != "" {
        err = r.client.SetApplication(poolName, application)
        if err != nil {
            resp.Diagnostics.AddWarning(
                "Error Setting Pool Application",
                fmt.Sprintf("Pool created but could not set application %s: %s", application, err.Error()),
            )
        }
    }

    // Set the ID and computed values
    plan.ID = types.StringValue(poolName)
    plan.PoolType = types.StringValue(poolType)
    plan.PgNum = types.Int64Value(pgNum)
    plan.PgpNum = types.Int64Value(pgpNum)
    plan.Size = types.Int64Value(size)
    if application != "" {
        plan.Application = types.StringValue(application)
    }

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *poolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state PoolResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get pool information from Ceph
    poolName := state.Name.ValueString()
    poolData, err := r.client.GetPool(poolName)
    if err != nil {
        // If the pool is not found, remove it from state
        if err.Error() == "pool not found" {
            resp.State.RemoveResource(ctx)
            return
        }
        
        resp.Diagnostics.AddError(
            "Error Reading Ceph Pool",
            fmt.Sprintf("Could not read pool %s: %s", poolName, err.Error()),
        )
        return
    }

    // Update the state with the latest data
    if poolType, ok := poolData["type"].(string); ok {
        state.PoolType = types.StringValue(poolType)
    }

    if pgNum, ok := poolData["pg_num"].(float64); ok {
        state.PgNum = types.Int64Value(int64(pgNum))
    }
    
    if pgpNum, ok := poolData["pgp_num"].(float64); ok {
        state.PgpNum = types.Int64Value(int64(pgpNum))
    }

    if size, ok := poolData["size"].(float64); ok {
        state.Size = types.Int64Value(int64(size))
    }

    if apps, ok := poolData["application_metadata"].(map[string]interface{}); ok {
        for app := range apps {
            state.Application = types.StringValue(app)
            break
        }
    }

    // Save updated data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *poolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan PoolResourceModel
    var state PoolResourceModel

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    poolName := plan.Name.ValueString()
    changed := false

    // Update pg_num if changed
    if !plan.PgNum.IsNull() && plan.PgNum.ValueInt64() != state.PgNum.ValueInt64() {
        pgNum := int(plan.PgNum.ValueInt64())
        err := r.client.SetPoolProperty(poolName, "pg_num", pgNum)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Updating Pool PG Num",
                fmt.Sprintf("Could not update pg_num for pool %s: %s", poolName, err.Error()),
            )
            return
        }
        changed = true
    }
    
    // Update pgp_num if changed
    if !plan.PgpNum.IsNull() && plan.PgpNum.ValueInt64() != state.PgpNum.ValueInt64() {
        pgpNum := int(plan.PgpNum.ValueInt64())
        err := r.client.SetPoolProperty(poolName, "pgp_num", pgpNum)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Updating Pool PGP Num",
                fmt.Sprintf("Could not update pgp_num for pool %s: %s", poolName, err.Error()),
            )
            return
        }
        changed = true
    }

    // Update size if changed
    if !plan.Size.IsNull() && plan.Size.ValueInt64() != state.Size.ValueInt64() {
        size := int(plan.Size.ValueInt64())
        err := r.client.SetPoolProperty(poolName, "size", size)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Updating Pool Size",
                fmt.Sprintf("Could not update size for pool %s: %s", poolName, err.Error()),
            )
            return
        }
        changed = true
    }

    // Update application if changed
    if !plan.Application.IsNull() && 
       (state.Application.IsNull() || plan.Application.ValueString() != state.Application.ValueString()) {
        application := plan.Application.ValueString()
        err := r.client.SetApplication(poolName, application)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Updating Pool Application",
                fmt.Sprintf("Could not update application for pool %s: %s", poolName, err.Error()),
            )
            return
        }
        changed = true
    }

    // If something changed, refresh the state
    if changed {
        poolData, err := r.client.GetPool(poolName)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading Updated Ceph Pool",
                fmt.Sprintf("Could not read updated pool %s: %s", poolName, err.Error()),
            )
            return
        }
        
        // Update the state with the latest data
        if poolType, ok := poolData["type"].(string); ok {
            plan.PoolType = types.StringValue(poolType)
        }

        if pgNum, ok := poolData["pg_num"].(float64); ok {
            plan.PgNum = types.Int64Value(int64(pgNum))
        }
        
        if pgpNum, ok := poolData["pgp_num"].(float64); ok {
            plan.PgpNum = types.Int64Value(int64(pgpNum))
        }

        if size, ok := poolData["size"].(float64); ok {
            plan.Size = types.Int64Value(int64(size))
        }

        if apps, ok := poolData["application_metadata"].(map[string]interface{}); ok {
            for app := range apps {
                plan.Application = types.StringValue(app)
                break
            }
        }
    }

    // Save updated data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *poolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state PoolResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Delete the pool
    poolName := state.Name.ValueString()
    err := r.client.DeletePool(poolName)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting Ceph Pool",
            fmt.Sprintf("Could not delete pool %s: %s", poolName, err.Error()),
        )
        return
    }
}

// ImportState imports the resource state
func (r *poolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Use the ID (pool name) as the import identifier
    resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}