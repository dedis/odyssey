# Enclave Manager

![ENM logo](assets/enm-logo.png)

## Purpose

This component is a go/http server responsible for interfacing with VMware
vCloud REST API. It receives requests from the datascientist manager.

## Set up

1. Export the necessary environment variables present in the `variables.sh.template`
2. Build and run with `go run main.go`

## Usage

The enclave manager is a middleware components that is used throught the
DOManager and DSManager. So there is not much to do from this side.

There is a minimal debug functionality to browse and destroy all instances if
you go to http://localhost:5001/.