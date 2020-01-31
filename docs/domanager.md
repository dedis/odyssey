# Data Owner Manager

This components allows a data owner to upload a dataset, as well as setting
attributes on it.

## Configuration 

The Data Scientist needs the informations about how to connect to the
cloud provider. Here is the template to fill:

```
export MINIO_ENDPOINT="..."
export MINIO_KEY="..."
export MINIO_SECRET="..."
```

The following executables are needed at the root of `domanager/app`:

- bcadmin
- catadmin
- csadmin

Then rename `app/config.toml.template` to `app/config.toml` and fill it
with the appropriate settings.

## Run

from `domanager/app` run `go run main.go`.
The app is reachable from localhost:5002.
