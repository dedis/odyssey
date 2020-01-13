# Enclave Manager

## Purpose

This component is a go/http server responsible for interfacing with VMware
vCloud REST API. It receives requests from the datascientist manager.

## Set up

1. Export the necessary environment variables present in the `variables.sh.template`
2. Build and run with `go run main.go`

## Usage

Still looking for a nice way to generate a documentation for a REST api...

## Sample code to use the REST API

```bash
# Gets and prints a valid token for vCloud REST API

# https://github.com/swisscom/dcsplus-utils/blob/master/vcd-egw-syslogsetter/setsyslogserver.sh
# https://www.vmware.com/support/vcd/doc/rest-api-doc-1.5-html/landing-user_operations.html
# https://www.onlinetool.io/xmltogo/

# Don't forget to set:
# export VCD_PASS="XXXXXX"
# export VCD_HOST="XXXXX"
# export VCD_API_USER="XXXXXX"

# Cookies need to be used, in case of a WAF doing session management in front
# of the vCloudDirector cells
COOKIEFILE=$(mktemp)
HEADERFILE=$(mktemp)

curl -s -o /dev/null -b $COOKIEFILE https://$VCD_HOST/api/versions

# Logging in
curl -s -o /dev/null -H 'Accept: application/*;version=27.0' -b $COOKIEFILE\
    -c $COOKIEFILE -X POST -D $HEADERFILE\
    -u $VCD_API_USER:$VCD_PASS https://$VCD_HOST/api/sessions
AUTHH=$(cat $HEADERFILE | grep 'x-vcloud-authorization')
AUTHH="${AUTHH%%[[:cntrl:]]}"
echo "Result is: $AUTHH"
```