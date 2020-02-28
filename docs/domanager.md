# Data Owner Manager

This components allows a data owner to upload a dataset, as well as setting
attributes on it.

## Executables

The following executables are needed at the root of `domanager/app`:

- bcadmin
- catadmin
- csadmin

Put them there with:

```
$ cd domanager/app
$ cp `go env GOPATH`/bin/{bcadmin,catadmin,csadmin} .
```

## Configuration

You must have the variables from `variables.sh` loaded in your shell.
(See section "Setup, Cloud configuration".)

Then rename `app/config.toml.template` to `app/config.toml` and fill it
with the appropriate settings, following the instructions below to find
each of them.

## Catalog

To create a new catalog, ensure that your Darc has the spawn:catalog rule on it,
and then use `catadmin`. You will need the `BC` variable set, which should be
set in your `variables.sh` file.

```
$ source ../../variables.sh
$ catadmin contract catalog spawn
```

## Enclave manager

In order to upload new datasets, the enclave manager must be running on
`localhost:5000`. See its corresponding documentation in order to run
it.

If you are not using the Enclave manager, you can edit `config.toml`
to set `Standalone = true`.

## Run

From `domanager/app` run `go run main.go`. The app is reachable from: http://localhost:5002/
