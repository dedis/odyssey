# Cryptutil

## Purpose

Command line utility based on the golang crypto library that offers AES-GCM
symetric encryption / decryption of data.

Based on https://gist.github.com/kkirsche/e28da6754c39d5e7ea10

## Set up

**Install cryptutil**

```bash
cd /Users/nkocher/GitHub/odyssey/cryptutil
go install
```

**Encrypt the datasets**

```bash
cryptutil encrypt --key KEY --initVal INIT_VAL --readData -export < /Users/nkocher/GitHub/odyssey/secret/datasets/1_titanic.csv  > /Users/nkocher/GitHub/odyssey/secret/datasets/1_titanic.csv.aes
```

Same for the second dataset.

**Launch the tests**

```bash
cd /Users/nkocher/GitHub/odyssey/cryptutil
./test.sh
```