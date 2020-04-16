# Enclave Manager

![ENM logo](assets/enm-logo.png)

## Purpose

This component is a go/http server that offers its services to the DSManager and
the DOManager. 

The DSManager uses it as a proxy for interacting with the enclave (we do it with
VMware vCloud REST API), while the DOManager uses it to get a new DARC when
uploading a dataset (each dataset must have its own DARC).

The ENManager is a central point in our system that is trusted. It has the power
to deliver access to enclaves containing unencrypted datasets and can manipulate
the dataset DARCs.

## Set up

1. Export all environment variables present in the `variables.sh.template`
2. Copy `config.toml.template` to `config.toml` and fill it

## Usage

From `enclavem/app` run `go run main.go`. The app is reachable from:
http://localhost:5000/, but there isn't much to do there since this http server
is a middleware for the DSManager and the DOManager. There you will find a
minimal debug functionality to browse and destroy all instances.

You can exit the server with <kbd>ctrl</kbd>+<kbd>c</kbd>.