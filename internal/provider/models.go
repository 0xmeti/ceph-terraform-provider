package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// CephProviderModel describes the provider configuration
type CephProviderModel struct {
    Endpoint types.String `tfsdk:"endpoint"`
    Username types.String `tfsdk:"username"`
    Password types.String `tfsdk:"password"`
}

// PoolResourceModel describes the pool resource
type PoolResourceModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    PoolType    types.String `tfsdk:"pool_type"`
    PgNum       types.Int64  `tfsdk:"pg_num"`
    PgpNum      types.Int64  `tfsdk:"pgp_num"`
    Size        types.Int64  `tfsdk:"size"`
    Application types.String `tfsdk:"application"`
}

// PoolDataSourceModel describes the pool data source
type PoolDataSourceModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    PoolType    types.String `tfsdk:"pool_type"`
    PgNum       types.Int64  `tfsdk:"pg_num"`
    Size        types.Int64  `tfsdk:"size"`
    Application types.String `tfsdk:"application"`
}
