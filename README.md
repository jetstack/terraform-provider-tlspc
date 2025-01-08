# Terraform Provider for Venafi TLS Protect Cloud Platform

This is an experimental Terraform Provider for the Venafi TLS Protect Cloud Platform.

It has limited support at this stage. All contributions and feedback are both welcome and appreciated.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.9.8
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to test your local build, you will need to setup a [`dev_override`](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install). The provider address is `venafi.com/dev/tlspc`.
