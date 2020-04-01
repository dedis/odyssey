# Cryptutil

## Purpose

Command line utility based on the golang crypto library that offers AES-GCM
symetric encryption / decryption of data.

Based on https://gist.github.com/kkirsche/e28da6754c39d5e7ea10

## Set up

If you followed the [setup instructions](setup.md#generate-the-executables) the
cryptutil executable should already be in your path. Otherwise you can do the
following:

```bash
cd cryptutil
go install
```

## Use

You can use the `-h` argument to get help. For example `cryptutil -h`.

## Example


**Encrypt a dataset**

```bash
cryptutil encrypt --key aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --initVal bbbbbbbbbbbbbbbbbbbbbbbb --readData -export < titanic.csv > titanic.csv.aes
```

**Decrypt a dataset**

```bash
cryptutil decrypt --key aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --initVal bbbbbbbbbbbbbbbbbbbbbbbb --readData -export < titanic.csv.aes > titanic.csv  
```

We can use a condensed version with the `keyAndInitVal` option: 

**Encrypt a dataset (condensed)**

```bash
cryptutil encrypt --keyAndInitVal aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbb --readData -export < titanic.csv > titanic.csv.aes
```

**Decrypt a dataset (condensed)**

```bash
cryptutil decrypt --keyAndInitVal aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbb --readData -export < titanic.csv.aes > titanic.csv  
```

## Tests

```bash
cd cryptutil
./test.sh
```