# Setup

This section provides the instruction on global configurations such as
the cloud provider or the executables that each component need.

## Cloud configuration

Make sure you have an available S3-compatible cloud endpoint.

If you are using a remote cloud storage system, skip to
"CloudStorage Setup" below.

### Local Minio install

If are not using a remote cloud storage system, Minio is a good choice
for running your own. Here is how to install Minio for use in a
local development setup.

Make a copy of the `variables.sh.template` file, and edit it to set
`MINIO_SECRET_KEY` to something of your own. Source the file:

```
cp variables.sh.template variables.sh
# edit it
source variables.sh
```

Get the binaries (Linux).

```
wget https://dl.min.io/server/minio/release/linux-amd64/minio
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x minio mc
mkdir data
./minio server --address localhost:9990 `pwd`/data
```

On MacOS:

```
wget https://dl.min.io/server/minio/release/darwin-amd64/minio
wget https://dl.min.io/client/mc/release/darwin-amd64/mc
chmod +x minio mc
mkdir data
./minio server --address localhost:9990 `pwd`/data
```

### Cloud Storage Setup

If you are using Minio, use this line to set the config for the `mc` tool:

```
mc config host add dedis http://localhost:9990 $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
```

If you are using another remote cloud storage system, you'll need to check the
docs for how to configure `mc` for it.

Now you need to make a bucket for datasets, which must be named `dedis/datasets`.

```
mc mb dedis/datasets
```

## Generate the executables

We heavily make use of direct calls to executables as a means to interface
with Cothority. As such, you will need to build them and make sure they
are in the PATH.

Here is a summary
of the executables needed by each component:

|               |catadmin|cryptutil|pcadmin|bcadmin|csadmin|
|---------------|--------|---------|-------|-------|-------|
|`domanager/app`| x      |         |       | x     | x     |
|`dsmanager/app`| x      |         | x     | x     | x     |
|enclave VM     |        | x       | x     | x     | x     |
|`enclavem/app` |        |         | x     | x     | x     |


You can build all the required binaries and install them into $GOPATH/bin by
using the Makefile:

```make
make
```

Note that this target compile "bcadmin" and "csadmin" with the v3.4.4 of
cothority, which could erase a more recent version that is already in your
$GOPATH. If you do not want that you can manually select the needed executables:

```make
make catadmin cryptutil pcadmin
```

There are additional setup steps for each component that you will find
in their associated documentation (chapter "Components").

## Run a set of local conodes, and start a ledger

### Using a local ledger

If you are not using a public ledger, you will need to run your own local one.

#### Run the nodes

Build the conode binary and use the `run_nodes.sh` script in order to get a set
of conodes up and running. This command will put the conodes data under the
"cothority_data" folder (and create it if needed):

```bash
cd ledger/conode
go build
./run_nodes.sh -d cothority_data -v 2
```

This command will run 3 nodes, saving their databases and credentials in the
`ledger/conode/cothority_data` folder.

#### Initialize a new skipchain

Now that the conodes are runing, we must initialize the ledger. The following
steps use `bcadmin`. If you followed the steps in "Generate the executables",
then the `bcadmin` command line utility should already be in your path.

Make sure to run `varialbes.sh` if you set a custom `BC_CONFIG` variables. If
set, this variable tells `bcadmin` where to store and get the configuration
files.

Create a new skipchain:

```bash
# This file was created by the run_nodes.sh script
bcadmin create ledger/conode/cothority_data/public.toml 
# The output of the command offers you to export the BC variable. Copy/past the
# last line into to your terminal
export BC="path/to/folder/bc-BYZCOIN_ID.cfg"
```

It is also a good idea to save the `export BC=...` command in the
`variables.sh`.

Upon creation, bcadmin created for you the admin identity and DARC, which are
described by the config file saved in the BC variable. You can display the
content of this configuration file with the following command:

```bash
bcadmin info
```

#### Initialize Calypso

In the following steps we will set up Calypso. Calypso is the technology that
allows secrets sharing on the blockchain. In our case, we use it to store the
symmetric keys of the encrypted datasets. The steps are performed with the
`csadmin` command line utility. If you followed the steps in "Generate the
executables", then the `csadmin` command line utility should already be in your
path.

The first step is to authorize each node in participating to a secret sharing
protocol. You must run one command for each server, giving the path to its
private key file and the Byzcoin ID. The Byzcoin ID can be retreived with the
`bcadmin info` command:

```bash
# Display the configuration info: spot the Byzcoin ID
bcadmin info
# Authorize each node:
csadmin authorize ledger/conode/cothority_data/co1/private.toml BYZCOIN_ID
csadmin authorize ledger/conode/cothority_data/co2/private.toml BYZCOIN_ID
csadmin authorize ledger/conode/cothority_data/co3/private.toml BYZCOIN_ID
```

The next step is to setup a long term secret and then launch a distributed key
generation protocol. Do those two steps with the following command. You will
need the output of the first command to launch the second one:

```bash
# setup a long term secret, note the instance id
csadmin contract lts spawn
# start a distributed key generation protocol
csadmin dkg start -i INSTANCE_ID
```

The instance id that outputs the first command correponds to the needed `LtsID`
configuration entry in `domanager/app/config.toml`, while the "X" that is
displayed by the second command corresponds to the `LtsKey` entry.

#### Add rules to the DARC

In order to use the custom Odyssey smart contracts, we need to allow actions on
those smart contract in our DARC (the one specified by the BC config file).

Launch the following commands or only those containing the rules you are
interested in:

```bash
# Firstly, spot and note your identity (something like "ed25519:aef123...")
bcadmin info
# Save the identity into a variable
id=IDENTITY
# Note: do not run those commands into one single batch in order to let the 
# signer counter to be refreshed.
bcadmin darc rule -rule "spawn:calypsoWrite" -id $id
bcadmin darc rule -rule "spawn:calypsoRead" -id $id
bcadmin darc rule -rule "spawn:odysseyproject" -id $id
bcadmin darc rule -rule "invoke:odysseyproject.update" -id $id
bcadmin darc rule -rule "invoke:odysseyproject.updateStatus" -id $id
bcadmin darc rule -rule "invoke:odysseyproject.setURL" -id $id
bcadmin darc rule -rule "invoke:odysseyproject.setAccessPubKey" -id $id
bcadmin darc rule -rule "invoke:odysseyproject.setEnclavePubKey" -id $id
bcadmin darc rule -rule "spawn:odysseycatalog" -id $id
bcadmin darc rule -rule "invoke:odysseycatalog.addOwner" -id $id
bcadmin darc rule -rule "invoke:odysseycatalog.updateMetadata" -id $id
bcadmin darc rule -rule "invoke:odysseycatalog.deleteDataset" -id $id
bcadmin darc rule -rule "invoke:odysseycatalog.archiveDataset" -id $id
# You can print your darc and notice the new rules added
bcadmin darc show
```

### Using a public ledger

If you are using a public ledger, the admin will ask for your public
key and then give you instructions on how to use `bcadmin link` to connect
to the ledger. The resulting `bc-*.cfg` file will be used instead of the
one created in the previous section.

## Generate doc

You can generate the REST documentation with

```bash
swag init
```

then you can launch the data scientist manager and navigate to `docs/`.

## Skipchain Explorer

Ensure you have `yarn` installed

```
brew install yarn
```

Begin by cloning skipchain explorer and switching to the odyssey branch

```
git clone https://github.com/gnarula/student_18_explorer.git
git checkout odyssey
```

Run the development server

```
make build
yarn run serve
```

Click on Roster on the top right corner. Add the contents of your `roster.toml` in the dialog and click save.

Select the skipchain from the dropdown menu. Use the `Status` tab to see the list of conodes and the `Graph` tab to see a visualisation of the blocks.