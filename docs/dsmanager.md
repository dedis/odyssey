# Data Scientist Manager

## Purpose

This component is a go/http server that offer the upfront interface (understand
"GUI" by that) to the data scientist. This components mainly talks to the
enclave manager

## Set up

1. Generate an ssh keypair (in the `/app` folder): `ssh-keygen -t rsa -b 4096 -f ./id_rsa -C "dsmanager-key"`

2. Update the `config.toml`

3. Export the needed variables:

```
export MINIO_ENDPOINT=
export MINIO_KEY=
export MINIO_SECRET=
```

4. Build and run with `go run main.go`

## Usage

Go to `http://localhost:5001`. From there you can browse the list of datasets
and create a project.