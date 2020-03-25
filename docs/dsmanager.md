# Data Scientist Manager

![DOM logo](assets/dsm-logo.png)

This component allows a data scientist to request and use datasets based on the
attributes of a project. The system only releases datasets if the attributes of
the project comply with the access control rules set on each dataset by their
respective data owners.

The datasets are never directly delivered to the data scientist, but only via an
encalve that controls the lifecycle of the data. Once the data are no longer
needed by the data scientist, the enclave is entirely destroyed and no trace of
the data can be recovered. The lifecycle of the data is witnessed by the nodes
of the blokchain.

## Executables

The following executables are needed at the root of `dsmanager/app` :

- bcadmin
- pcadmin
- catadmin
- csadmin

If you followed the [setup instructions](setup.md#generate-the-executables)
those executables should already be on your gopath. Put them there with:

```
cd dsmanager/app
cp `go env GOPATH`/bin/{bcadmin,catadmin,csadmin,pcadmin} .
```

## Configuration

You must have the MINO_* variables from `variables.sh` loaded in your shell.
(See section "Setup, Cloud configuration".)

Then rename `app/config.toml.template` to `app/config.toml` and fill it
with the appropriate settings, following the instructions below to find
each of them.

## SSH keypair

The enclave grants access via a ssh keypair that it find at the `PubKeyPath`
variable that you must set in the configuration file. Put your public key there
or generate a new keypair:

```bash
# dsmanager/app
ssh-keygen -t rsa -b 4096 -f ./id_rsa -C "dsmanager-key"`
```

## Credentials creation (DARC and Key)

An admin of the ledger could use those commands in order to create a DARC and a
key for a data scientist. The key will allow the data scientist to authenticate
on the ledger and the DARC contains the authorized actions.

```bash
# we assume that the admin key is available in "./darc_key.txt"
mkdir data_scientist
bcadmin -c . darc add --out_id data_scientist/darc_id.txt --out_key data_scientist/darc_key.txt --desc "DARC for the data scientist"  --unrestricted

bcadmin darc rule -rule "spawn:odysseyproject" -darc $(cat data_scientist/darc_id.txt) -sign $(cat data_scientist/darc_key.txt) -identity $(cat data_scientist/darc_key.txt)
bcadmin darc rule -rule "invoke:odysseyproject.updateMetadata" -darc $(cat data_scientist/darc_id.txt) -sign $(cat data_scientist/darc_key.txt) -identity $(cat data_scientist/darc_key.txt)
# We need both the enclave manager and the data scientist to update the status
# because the data scientist sets the status when it updates the attributes of
# the project and the enclave manager sets all the other statuses (preparing,
# unlocking, destroying).
bcadmin darc rule -rule "invoke:odysseyproject.updateStatus" -darc $(cat data_scientist/darc_id.txt) -sign $(cat data_scientist/darc_key.txt) -identity "$(cat darc_key.txt) | $(cat data_scientist/darc_key.txt)" --replace
bcadmin darc rule -rule "invoke:odysseyproject.setURL" -darc $(cat data_scientist/darc_id.txt) -sign $(cat data_scientist/darc_key.txt) -identity $(cat darc_key.txt)
# the access pub key is set during the spawn
bcadmin darc rule -rule "invoke:odysseyproject.setEnclavePubKey" -darc $(cat data_scientist/darc_id.txt) -sign $(cat data_scientist/darc_key.txt) -identity $(cat darc_key.txt)
```

## Enclave manager

In order to upload new datasets, the enclave manager must be running on
`localhost:5000`. See its corresponding documentation in order to run
it.


## Run

From `dsmanager/app` run `go run main.go`. The app is reachable from:
http://localhost:5001/