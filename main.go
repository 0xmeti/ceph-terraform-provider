package main

import (
    "context"
    "flag"
    "log"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/0xmeti/terraform-provider-ceph/internal/provider"
)

// version will be set by goreleaser
var version string = "dev"

func main() {
    var debug bool

    flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
    flag.Parse()

    opts := providerserver.ServeOpts{
        Address: "registry.terraform.io/0xmeti/ceph",
        Debug:   debug,
    }

    err := providerserver.Serve(context.Background(), provider.New(version), opts)

    if err != nil {
        log.Fatal(err.Error())
    }
}
