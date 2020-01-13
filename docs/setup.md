# Setup

This section covers setup configurations that does not concern the components
and tools described bellow.

## Cloud configuration

**Upload the encrypted datasets**

```bash
# Create the 'datasets' bucket
mc mb dedis/datasets
# Copy one of the encrypted dataset
mc cp /Users/nkocher/GitHub/odyssey/secret/datasets/1_titanic.csv.aes dedis/datasets
```

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