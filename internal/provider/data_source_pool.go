package provider

import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
    _ datasource.DataSource              = &poolDataSource{}
    _ datasource.DataSourceWithConfigure = &poolDataSource{}
)

// NewPoolDataSource is a helper function to simplify the provider implementation
func NewPoolDataSource() datasource.DataSource {
    return &poolDataSource{}
}

// poolDataSource is the data source implementation
type poolDataSource struct {
    client *CephClient
}

// Metadata returns the data source type name
func (d *poolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_pool"
}

// Schema defines the schema for the data source
func (d *poolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Fetches information about a Ceph pool.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "Pool identifier",
                Computed:    true,
            },
            "name": schema.StringAttribute{
                Description: "Name of the pool",
                Required:    true,
            },
            "pool_type": schema.StringAttribute{
                Description: "Type of the pool (replicated or erasure)",
                Computed:    true,
            },
            "pg_num": schema.Int64Attribute{
                Description: "Number of placement groups",
                Computed:    true,
            },
            "size": schema.Int64Attribute{
                Description: "Replication size",
                Computed:    true,
            },
            "application": schema.StringAttribute{
                Description: "Pool application (rbd, cephfs, rgw)",
                Computed:    true,
            },
        },
    }
}

// Configure adds the provider configured client to the data source
func (d *poolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*CephClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *CephClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    d.client = client
}

// Read refreshes the Terraform state with the latest data
func (d *poolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var state PoolDataSourceModel

    // Read Terraform configuration data into the model
    resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get pool information from Ceph
    poolName := state.Name.ValueString()
    poolData, err := d.client.GetPool(poolName)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Ceph Pool",
            fmt.Sprintf("Could not read pool %s: %s", poolName, err.Error()),
        )
        return
    }

    // Map response body to model
    state.ID = types.StringValue(poolName)
    state.Name = types.StringValue(poolName)

    // Extract pool type
    if poolType, ok := poolData["type"].(string); ok {
        state.PoolType = types.StringValue(poolType)
    }

    // Extract pg_num
    if pgNum, ok := poolData["pg_num"].(float64); ok {
        state.PgNum = types.Int64Value(int64(pgNum))
    }

    // Extract size
    if size, ok := poolData["size"].(float64); ok {
        state.Size = types.Int64Value(int64(size))
    }

    // Extract application
    if apps, ok := poolData["application_metadata"].(map[string]interface{}); ok {
        for app := range apps {
            state.Application = types.StringValue(app)
            break // Take the first application
        }
    }

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
