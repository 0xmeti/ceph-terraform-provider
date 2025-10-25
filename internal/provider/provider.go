package provider

import (
    "context"
    "os"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    // Removed the unused import: "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
    _ provider.Provider = &cephProvider{}
)

// New is a helper function to simplify provider server
func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &cephProvider{
            version: version,
        }
    }
}

// cephProvider is the provider implementation
type cephProvider struct {
    version string
}

// Metadata returns the provider type name
func (p *cephProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "ceph"
    resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *cephProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Interact with Ceph cluster via REST API.",
        Attributes: map[string]schema.Attribute{
            "endpoint": schema.StringAttribute{
                Description: "The Ceph API endpoint URL. Can also be set via CEPH_ENDPOINT environment variable.",
                Optional:    true,
            },
            "username": schema.StringAttribute{
                Description: "The username for Ceph API authentication. Can also be set via CEPH_USERNAME environment variable.",
                Optional:    true,
            },
            "password": schema.StringAttribute{
                Description: "The password for Ceph API authentication. Can also be set via CEPH_PASSWORD environment variable.",
                Optional:    true,
                Sensitive:   true,
            },
        },
    }
}

// Configure prepares a Ceph API client for data sources and resources
func (p *cephProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config CephProviderModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // If practitioner provided a configuration value for any of the
    // attributes, it must be a known value.

    if config.Endpoint.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("endpoint"),
            "Unknown Ceph API Endpoint",
            "The provider cannot create the Ceph API client as there is an unknown configuration value for the Ceph API endpoint. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the CEPH_ENDPOINT environment variable.",
        )
    }

    if config.Username.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("username"),
            "Unknown Ceph API Username",
            "The provider cannot create the Ceph API client as there is an unknown configuration value for the Ceph API username. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the CEPH_USERNAME environment variable.",
        )
    }

    if config.Password.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("password"),
            "Unknown Ceph API Password",
            "The provider cannot create the Ceph API client as there is an unknown configuration value for the Ceph API password. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the CEPH_PASSWORD environment variable.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Default values to environment variables, but override
    // with Terraform configuration value if set.

    endpoint := os.Getenv("CEPH_ENDPOINT")
    username := os.Getenv("CEPH_USERNAME")
    password := os.Getenv("CEPH_PASSWORD")

    if !config.Endpoint.IsNull() {
        endpoint = config.Endpoint.ValueString()
    }

    if !config.Username.IsNull() {
        username = config.Username.ValueString()
    }

    if !config.Password.IsNull() {
        password = config.Password.ValueString()
    }

    // If any of the expected configurations are missing, return
    // errors with provider-specific guidance.

    if endpoint == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("endpoint"),
            "Missing Ceph API Endpoint",
            "The provider cannot create the Ceph API client as there is a missing or empty value for the Ceph API endpoint. "+
                "Set the endpoint value in the configuration or use the CEPH_ENDPOINT environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if username == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("username"),
            "Missing Ceph API Username",
            "The provider cannot create the Ceph API client as there is a missing or empty value for the Ceph API username. "+
                "Set the username value in the configuration or use the CEPH_USERNAME environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if password == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("password"),
            "Missing Ceph API Password",
            "The provider cannot create the Ceph API client as there is a missing or empty value for the Ceph API password. "+
                "Set the password value in the configuration or use the CEPH_PASSWORD environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Create a new Ceph client using the configuration values
    client := NewCephClient(endpoint, username, password)

    // Make the Ceph client available during DataSource and Resource
    // type Configure methods.
    resp.DataSourceData = client
    resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider
func (p *cephProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewPoolDataSource,
    }
}

// Resources defines the resources implemented in the provider
func (p *cephProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewPoolResource,
    }
}