package provider

import (
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during acceptance testing
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "ceph": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
    // You can add checks here to ensure test environment is ready
    // For example: check if CEPH_ENDPOINT environment variable is set
}
