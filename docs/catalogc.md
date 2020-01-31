# Catalogc

The catalog contract stores the list of authorized owners along with
their dataset. The best way to discover how data is stored on the
catalog is to have a look at `catalogc/data.go`. We allow each owner to
edit its own space on the catalog by checking that the identity of the
owner corresponds to the identity stored at the requested space.

The best way to discover how to use the catalog is to use its CLI
`catalogc help`.
