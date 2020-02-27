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
$ cp variables.sh.template variables.sh
# edit it
$ source variables.sh
```

Get the binaries (Linux).

```
$ wget https://dl.min.io/server/minio/release/linux-amd64/minio
$ wget https://dl.min.io/client/mc/release/linux-amd64/mc
$ chmod +x minio mc
$ mkdir data
$ ./minio server --address localhost:9990 `pwd`/data
```

On MacOS:

```
$ wget https://dl.min.io/server/minio/release/darwin-amd64/minio
$ wget https://dl.min.io/client/mc/release/darwin-amd64/mc
$ chmod +x minio mc
$ mkdir data
$ ./minio server --address localhost:9990 `pwd`/data
```

### Cloud Storage Setup

If you are using Minio, use this line to set the config for the `mc` tool:

```
$ mc config host add dedis http://localhost:9990 $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
```

If you are using another remote cloud storage system, you'll need to check the
docs for how to configure `mc` for it.

Now you need to make a bucket for datasets, which must be named `dedis/datasets`.

```
$ mc mb dedis/datasets
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


You can build all the required binaries and install them into $GOPATH/bin like this:

```
git clone https://github.com/dedis/cothority
cd cothority
go install ./...
cd ..
git clone https://github.com/dedis/odyssey
cd odyssey
go install ./...
```

There are additional setup steps for each component that you will find
in their associated documentation (chapter "Components").

## Run a set of local conodes, and start a ledger

### Using a local ledger

If you are not using a public ledger, you will need to run your own local one.

...run conodes, make ledger, get bc file, set BC env variable in variables.sh, etc.

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
