# Catalogc

In the `catalogc` folder you'll find the definition of the "catalog contract"
and its corresponding CLI `catadmin`.

## Catalog contract

An instance of the catalog contract stores the list of authorized owners along
with their datasets. It is also responsible for storing the attributes that can
be set on datasets. The best way to discover how data is stored on the catalog
is to have a look at `catalogc/data.go`. We allow each owner to edit its own
space on the catalog by checking that the identity of the owner corresponds to
the identity stored at the requested space.

## catadmin

The "catalog contract" has its own CLI `catadmin`. If you followed the [setup
instructions](setup.md#generate-the-executables) the catadmin executable should
already be in your path. Otherwise you can do the following:

```bash
cd catalogc/catadmin
go install
```

You can use the `-h` argument to get help on how to use the CLI. For example
`catadmin -h`.