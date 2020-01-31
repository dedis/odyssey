# Setup

This section provides the instruction on global configurations such as
the cloud provider or the executables that each component need.

## Cloud configuration

Make sure you have an available cloud endpoint. Minio is a good choice.
Here are the steps we followed to install Minio on our server:

```
$ wget https://dl.min.io/server/minio/release/linux-amd64/minio
$ chmod +x minio
$ mkdir data
$ screen -S minio
$ ./minio server --address :9990 /home/JohnDoe/data
> <CTRL A> <D>
$ sudo ufw allow 9990
```

As an example, here is how we would use it with the CLI Minio client:

```
mc config host add dedis http://123.123.123.12:9990 <KEY> <SECRET>
mc mb dedis/new_bucket
echo "hello" | mc pipe "dedis/newbucket/$(date +%Y-%m-%d-%H%M-%S).txt"
```

## Generate the executables

We heavily make use of direct call to executables as a mean to interface
with cothority. As such, you will need to generate them an set up each
components with the appropriate executables. Here is a summary of the
exec dependencies of each components:

|               |catadmin|cryptutil|pcadmin|bcadmin|csadmin|
|---------------|--------|---------|-------|-------|-------|
|`domanager/app`| x      |         |       | x     | x     |
|`dsmanager/app`| x      |         | x     | x     | x     |
|enclave VM     |        | x       | x     | x     | x     |
|`enclavem/app` |        |         | x     | x     | x     |


## Pre-requist

Make sure you use a local version of cothority based on the
"odyssey-needs" branch (`git checkout odyssey-needs`), as well as the
local version of odyssey. For that, add the following replace directives
in your `go.mod` file:

```
replace go.dedis.ch/cothority/v3 => /Users/whatever/GitHub/cothority
replace github.com/dedis/odyssey => /Users/whatever/GitHub/odyssey
```

**catadmin**

```
cd catalogc/catadmin
go build
```

**cryptutil**

```
cd cryptutil
go build
```

**pcadmin**

```
cd projectc/pcadmin
go build
```

**bcadmin**

```
cd cothority/byzcoin/bcadmin
go build
```

**csadmin**

```
cd cothority/calypso/csadmin
go build
```

Then you can copy each executable to the right places, according to the
table above.

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
