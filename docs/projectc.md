# Projectc 

In the `projectc` folder you'll find the definition of the "project contract"
and its corresponding CLI `pcadmin`.

## Project contract

The "project contract" is the smart contract that holds all the informations
about a project that a data scientist creates in order to request datasets.
Hence, an instance of a "project contract" is created each time a data scientist
request one or more datasets. You can have a look at `projectc/contract.go` in
order to see what informations an instance of this contract holds.

## pcadmin

The "project contract" has its own CLI `pcadmin`. If you followed the [setup
instructions](setup.md#generate-the-executables) the pcadmin executable should
already be in your path. Otherwise you can do the following:

```bash
cd projectc/pcadmin
go install
```

You can use the `-h` argument to get help on how to use the CLI. For example
`pcadmin -h`.