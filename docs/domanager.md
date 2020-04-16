# Data Owner Manager

![DOM logo](assets/dom-logo.png)

This components allows a data owner to upload a dataset, as well as setting
attributes on it.

## Executables

The following executables are needed at the root of `domanager/app`:

- bcadmin
- catadmin
- csadmin

If you followed the [setup instructions](setup.md#generate-the-executables)
those executables should already be on your gopath. Put them at the root of this
module with:

```
cd domanager/app
cp `go env GOPATH`/bin/{bcadmin,catadmin,csadmin} .
```

## Configuration

You must have the MINO_* variables from `variables.sh` loaded in your shell (see
section "Setup, Cloud configuration").

Then rename `app/config.toml.template` to `app/config.toml` and fill it
with the appropriate settings, following the instructions below to find
each of them.

## Catalog

To create a new catalog (ie. an instance of the catalog smart contract), ensure
that your Darc has the spawn:catalog rule on it, and then use `catadmin`. You
will need the `BC` variable set, which should be set in your `variables.sh`
file.

```bash
source ../../variables.sh
# This command outputs the CATALOG_INSTANCE_ID
catadmin contract catalog spawn
```

Each data owner, represented as an identity string and DARC, must be first added
to the catalog before being able to upload a dataset. The following commads add
the current identity set in the BC config file to the catalog:

```bash
# Spot your identity with the following, something like "ed25519:aef123..."
bcadmin info
# Add your identity to the catalog
catadmin contract catalog invoke addOwner -i CATALOG_INSTANCE_ID --firstname John --lastname Doe --identityStr IDENTITY
# If you print the catalog, you can notice the change
catadmin contract catalog get -i CATALOG_INSTANCE_ID
```

If you are not in a standalone mode then you might want to set the attributes on
the catalog. You can use the attributes described in the [definition section](https://dedis.github.io/odyssey/#/attributes?id=definition) of the attribute documentation.

```bash
read -r -d '' JSONATTR << EOM
{
...your JSON attributes...
}
EOM
catadmin contract catalog invoke updateMetadata -i CATALOG_INSTANCE_ID --metadataJSON "$JSONATTR"
```

## Enclave manager

In order to upload new datasets, the enclave manager must be running on
`localhost:5000`. See its corresponding documentation in order to run
it.

If you are not using the Enclave manager, you can edit `config.toml`
to set `Standalone = true`.

## Run

From `domanager/app` run `go run main.go`. The app is reachable from: http://localhost:5002/.

You can exit the server with <kbd>ctrl</kbd>+<kbd>c</kbd>.

## Login

When you are able to open the DOManager, you then have to log in to manage your
datasets (the link is top right). To login you need to provide a credential
file. This file is made with `bcadmin`. For this step you need the `roster.toml`
file which contains the definition of the ledger and the ByzcoinID. Given that
you are in the same folder that contains your private key, you can create this
credential file with this command:

```bash
bcadmin -c . link ../roster.toml BYZCOIN_ID --darc DARC_ID --id DARC_KEY
```