# Terraform Provider for Ceph

This provider allows you to manage Ceph storage resources using Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21
- Ceph cluster with Dashboard API enabled

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Makefile:

```bash
make build
