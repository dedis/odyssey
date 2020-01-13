#!/bin/bash
  
# This script is intended to be run at each startup of the OS. If its the first
# startup, the script creates a new pub/priv key and saves the info on the cloud
# endpoint that is stored in the properties.
#
# It further checks if a special property 'reay_to_download' is set. If so it
# starts the process of downloading the data.
#
# The minio configi should already be set for the "dedis" bucket in
# ~/.mc/config.json it can be generated with `mc config host add dedis
# $MINIO_ENDPOINT $MINIO_KEY $MINIO_SECRET`.
#
# this scrip is launched via rc.local as root:
# /home/enclave/startup.sh >> /home/enclave/rclogs.log 2>&1

# For some strange reason, the script is launched two times, where the first
# time it get exited with no reason and abruptely. From what we saw, this can
# happen between an interval of ~2min.

# We assume that the bucket has already been created

# exit when any command fails
set -e -x

# Nice to see when we capture the outputs of this script.
echo "---"
date
echo "---"

# Crontab has only the following path: /usr/bin:/bin
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games"

export BC="/home/enclave/bc-file.cfg"
export BC_CONFIG="/home/enclave/key/"
MC_CONFIG_PATH="/home/enclave/.mc"

err(){
    echo "E: $*" >>/dev/stderr
}

# This request index is used to differentiate the logs for each request. This
# field should be incremented for each different request and we then send the
# logs to {ENDPOINT}/logs/{REQUEST_INDEX}/{timestamp} so that the other apps
# (for example the enclave manager) can listen to this location in order to get
# the logs for a particular request.
REQUEST_INDEX=$(vmtoolsd --cmd "info-get guestinfo.ovfenv" | perl -n -e '/key="request_index".*value="(.*)"/ && print $1')
if [ -z "$REQUEST_INDEX" ]; then
    # in case we can't find the request index we set a default value so that the
    # other app can still check a default location and read the logs
    err "couldn't find a request index, using a default"
    REQUEST_INDEX="-1"
fi


if [ -f /home/enclave/.endpoint ];  then
    ENDPOINT=$(cat /home/enclave/.endpoint)
fi
# Check the content of the endpoint
if [ -z "$ENDPOINT" ]; then

    ENDPOINT=$(vmtoolsd --cmd "info-get guestinfo.ovfenv" | perl -n -e '/key="vm_endpoint".*value="(.*)"/ && print $1')

    # Check the content of the endpoint
    if [ -z "$ENDPOINT" ]; then
        err "endpoint is empty"
        exit 1
    fi
    printf "%s" "$ENDPOINT" > /home/enclave/.endpoint   
fi

# This function expects the message as uniq argument. The minio service must be
# proprely set up before use and the $ENDPOINT and $REQUEST_INDEX set.
logInfo() {
    echo "$1 - $2"
    escapedMsg=$(printf '%s' "$1" | /home/enclave/jq-linux64 -aRs .)
    escapedDetails=$(printf '%s' "$2" | /home/enclave/jq-linux64 -aRs .)
    timestamp=$(date '+%Y-%m-%d-%H:%M:%S.%N')
    JSON_FMT='{"type":"%s","time":"%s","message":%s,"details":%s,"source":"%s"}\n'
    JSON_STRING=$(printf "$JSON_FMT" "info" "$timestamp" "$escapedMsg" "$escapedDetails" "enclave")
    echo "$JSON_STRING" > /tmp/startup_logInfo
    # we could use a pipe to directly save the content on the cloud, but using a
    # temporary file show better error output.
    /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_logInfo "dedis/$ENDPOINT/logs/$REQUEST_INDEX/$timestamp.json"
}

# This function expects the message as uniq argument. The minio service must be
# proprely set up before use and the $ENDPOINT and $REQUEST_INDEX set.
logError() {
    echo "$1 - $2"
    escapedMsg=$(printf '%s' "$1" | /home/enclave/jq-linux64 -aRs .)
    escapedDetails=$(printf '%s' "$2" | /home/enclave/jq-linux64 -aRs .)
    timestamp=$(date '+%Y-%m-%d-%H:%M:%S.%N')
    JSON_FMT='{"type":"%s","time":"%s","message":%s,"details":%s,"source":"%s"}\n'
    JSON_STRING=$(printf "$JSON_FMT" "error" "$timestamp" "$escapedMsg" "$escapedDetails" "enclave")
    echo "$JSON_STRING" > /tmp/startup_logInfo
    # we could use a pipe to directly save the content on the cloud, but using a
    # temporary file show better error output.
    /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_logInfo "dedis/$ENDPOINT/logs/$REQUEST_INDEX/$timestamp.json"
}

# This function expects the message as uniq argument. The minio service must be
# proprely set up before use and the $ENDPOINT and $REQUEST_INDEX set
logCloseOK() {
    echo "$1 - $2"
    escapedMsg=$(printf '%s' "$1" | /home/enclave/jq-linux64 -aRs .)
    escapedDetails=$(printf '%s' "$2" | /home/enclave/jq-linux64 -aRs .)
    timestamp=$(date '+%Y-%m-%d-%H:%M:%S.%N')
    JSON_FMT='{"type":"%s","time":"%s","message":%s,"details":%s,"source":"%s"}\n'
    JSON_STRING=$(printf "$JSON_FMT" "closeOK" "$timestamp" "$escapedMsg" "$escapedDetails" "enclave")
    echo "$JSON_STRING" > /tmp/startup_logInfo
    # we could use a pipe to directly save the content on the cloud, but using a
    # temporary file show better error output.
    /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_logInfo "dedis/$ENDPOINT/logs/$REQUEST_INDEX/$timestamp.json"
}

# This function expects the message as uniq argument. The minio service must be
# proprely set up before use and the $ENDPOINT and $REQUEST_INDEX set.
logCloseError() {
    err "$1 - $2"
    escapedMsg=$(printf '%s' "$1" | /home/enclave/jq-linux64 -aRs .)
    escapedDetails=$(printf '%s' "$2" | /home/enclave/jq-linux64 -aRs .)
    timestamp=$(date '+%Y-%m-%d-%H:%M:%S.%N')
    JSON_FMT='{"type":"%s","time":"%s","message":%s,"details":%s,"source":"%s"}\n'
    JSON_STRING=$(printf "$JSON_FMT" "closeError" "$timestamp" "$escapedMsg" "$escapedDetails" "enclave")
    echo "$JSON_STRING" > /tmp/startup_logCloseError
    # we could use a pipe to directly save the content on the cloud, but using a
    # temporary file show better error output.
    /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_logCloseError "dedis/$ENDPOINT/logs/$REQUEST_INDEX/$timestamp.json"
}

# Runs "$@" and check the return code. It its not 0, then exits and logs the
# error, otherwise logs the command. Warning be careful, if the command contains
# a pipe (|), for example `runCheck echo "Hello" | wc`, then the runCheck will
# only receive the `echo "hello"` part and the pipe will receive everything
# printed in the runCheck, which is not wanted. You must avoid the pipe if you
# want to use this fucntion.
runCheck() {
    logInfo "run check" "running \"$*\""
    "$@" > /tmp/startup_runCheck_log 2>&1
    retval=$?
    if [ $retval -ne 0 ]; then
        logError "\"$*\" returned code $retval" "outputed '$(cat /tmp/startup_runCheck_log)'"
        exit 1
    fi
}

# This helps to know if the script exits well or not, although it does not give
# any details on what went wrong. This is where we send the closeOK or
# closeError on the cloud.
finish() {
    retval=$?
    if [ $retval -eq 0 ]; then
        logCloseOK "script done" "startup.sh exits with error code 0"
    elif [ $retval -eq 10 ]; then
        echo "this is an exit 10, nothing to save on the cloud. bye."
    else
        logCloseError "script errored" "startup.sh exits with error code $retval, see the previous logs for more info. (startup_runCheck_log): $(cat /tmp/startup_runCheck_log)"
    fi
}

trap finish EXIT

if [ ! -f /home/enclave/.first ];  then
    logInfo "hi from the enclave" "this is the first time I am booting"

    runCheck dmesg
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_runCheck_log "dedis/$ENDPOINT/dmesg.txt"

    runCheck env
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_runCheck_log "dedis/$ENDPOINT/env.txt"

    runCheck echo "hello"
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_runCheck_log "dedis/$ENDPOINT/hello.txt"

    runCheck ifconfig
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_runCheck_log "dedis/$ENDPOINT/network.txt"

    runCheck hostname -I
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_runCheck_log "dedis/$ENDPOINT/ip.txt"

    runCheck date
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_runCheck_log "dedis/$ENDPOINT/date.txt"

    runCheck rm -rf $BC_CONFIG
    runCheck mkdir -p $BC_CONFIG

    # Download the bc config file
    logInfo "download config file" "let's download the bc config file from 'dedis/config/bc-file.cfg'"
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp dedis/config/bc-file.cfg $BC

    if [ ! -s $BC ]; then
        logError "error with config file" "bc file is zero length"
        exit 1
    fi
    
    logInfo "getting info" "let's get the output of 'bcadmin info'"
    runCheck /home/enclave/bcadmin info
    output_res="$(cat /tmp/startup_runCheck_log)"
    logInfo "go the info" "Here is the output of 'bcadmin info': $output_res"

    logInfo "creating the key with bcadmin" "using '/home/enclave/bcadmin key'"
    runCheck /home/enclave/bcadmin key

    PUB_KEY="$(cat /tmp/startup_runCheck_log)"
    logInfo "created pub/priv key." "Here is the public key: $PUB_KEY"

    # So we can use it later
    echo "$PUB_KEY" > /home/enclave/.pub_key
    echo "$PUB_KEY" > /tmp/startup_msg
    runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" -q cp /tmp/startup_msg "dedis/$ENDPOINT/pub_key.txt"
    logInfo "public key created" "public key saved in 'dedis/$ENDPOINT/pub_key.txt'"

    logInfo "everything is alright" "the startup sequence is done. This should be done only once."

    if [ ! -s $BC ]; then
        logError "config file not present" "end of first: bc file is zero length or does not exist"
        exit 1
    fi

    # So we only do this the first time.
    runCheck touch /home/enclave/.first
else
    logInfo "init sequence already done" "not the first time, doing nothing here"
fi

READY=$(vmtoolsd --cmd "info-get guestinfo.ovfenv" | perl -n -e '/key="ready_to_download".*value="(.*)"/ && print $1')
if [ ! -z "$READY" ]; then
    logInfo "ready to download" "we are ready to download, let's get the read and write instance IDs"
    # Parse the properties for the read and write instance ids
    # each property should contain a list of instance ids separated by a comma
    READ_INST_IDS_STR=$(vmtoolsd --cmd "info-get guestinfo.ovfenv" | perl -n -e '/key="read_instance_ids".*value="(.*)"/ && print $1')
    WRITE_INST_IDS_STR=$(vmtoolsd --cmd "info-get guestinfo.ovfenv" | perl -n -e '/key="write_instance_ids".*value="(.*)"/ && print $1')
    
    # Check the content of the properties
    if [ -z "$READ_INST_IDS_STR" ] || [ -z "$WRITE_INST_IDS_STR" ]; then
        logError "did not find any instace ids." "READ_INST_IDS_STR: $READ_INST_IDS_STR,  WRITE_INST_IDS_STR: $WRITE_INST_IDS_STR"
        exit 1
    fi

    logInfo "got the write ids" "found those write ids: $WRITE_INST_IDS_STR"
    logInfo "got the read ids" "found those read ids: $READ_INST_IDS_STR"

    logInfo "parsing" "parsing now those write ids and read ids strings into arrays"

    # Parse the properties into arrays
    IFS=',' read -ra READ_INST_IDS_ARRAY <<< "$READ_INST_IDS_STR"
    IFS=',' read -ra WRITE_INST_IDS_ARRAY <<< "$WRITE_INST_IDS_STR"

    logInfo "sanity check of the array" "checking the arrays"

    # Check the arrays
    lenread=${#READ_INST_IDS_ARRAY[@]}
    lenwrite=${#WRITE_INST_IDS_ARRAY[@]}

    if [ ! "$lenread" -eq "$lenwrite" ]; then
        logError "lenght of instance arrays are not equal." "len(READ)=$lenread, len(write)=$lenwrite"
        exit 1
    fi

    logInfo "arrays checked" "arrays checked, we can then ask to reencrypt the symmetric keys"

    keypath="/home/enclave/key/$(ls /home/enclave/key/ | head -1)"
    logInfo "got the key" "using this key in /home/enclave/key/: $keypath"

    runCheck rm -rf /home/enclave/replies
    runCheck mkdir -p /home/enclave/replies

    if [ ! -f "/home/enclave/bc-file.cfg" ]; then
        logError "config file not present" "didn't find '/home/enclave/bc-file.cfg'"
        exit 1
    fi

    logInfo "getting info" "let's get the output of 'bcadmin info'"
    runCheck /home/enclave/bcadmin info
    output_res="$(cat /tmp/startup_runCheck_log)"
    logInfo "got the info" "Here is the output of 'bcadmin info': $output_res"

    logInfo "creating folder" "creating the 'datasets' folder"
    runCheck rm -rf /home/enclave/datasets
    runCheck mkdir -p /home/enclave/datasets

    logInfo "loop over the datasets" "iterating over the instance IDs"

    # Iterate over the instance ids
    for (( i=0; i<lenread; i++ )); do
        rid=${READ_INST_IDS_ARRAY[$i]}
        wid=${WRITE_INST_IDS_ARRAY[$i]}
        logInfo "got the read and write ids" "read id: $rid, write id: $wid"

        logInfo "getting the write instance" "trying to get the write instance with 'csadmin contract write get -i $wid'"
        WRITE_DATA=$(/home/enclave/csadmin contract write get -i "$wid")

        logInfo "got the write data" "Here are the write data: $WRITE_DATA"
        CLOUD_URL=$(echo "$WRITE_DATA" | perl -n -e '/"CloudURL": "(.*?)",/ && print $1')
        logInfo "got the cloud url" "here is the extracted cloud url: $CLOUD_URL"

        DATASET_TITLE=$(echo "$WRITE_DATA" | perl -n -e '/"Title": "(.*?)",/ && print $1')
        logInfo "got the dataset title" "here is the extracted dataset title: $DATASET_TITLE"

        DATASET_FILENAME=$(basename -- "$CLOUD_URL")
        logInfo "got the dataset file name" "here is the dataset filename: $DATASET_FILENAME"

        logInfo "now reencrypting" "executing 'csadmin reecrypt' to get the reencrypted secret"
        /home/enclave/csadmin reencrypt --writeid "$wid" --readid "$rid" -x > "/home/enclave/replies/$i.bin"
        
        logInfo "now decrypting" "executing 'csadmin decrypt'"
        secret=$(/home/enclave/csadmin decrypt --key "$keypath" < "/home/enclave/replies/$i.bin" -x | xxd -p)

        logInfo "dowload the dataset" "so let's download dataset $CLOUD_URL"
        runCheck /home/enclave/mc --config-dir "$MC_CONFIG_PATH" cp "$CLOUD_URL" "/home/enclave/datasets/$DATASET_FILENAME"
        if [ ! -s "/home/enclave/datasets/$DATASET_FILENAME" ]; then
            logError "dataset not found" "no file downloaded, or it is zero length"
            exit 1
        fi

        # The basename should end with the .aes extension. So the new filename
        # will be the basename without the .aes extension.
        NEW_FILENAME=$(basename "$CLOUD_URL" .aes)
        logInfo "decrypting" "now let's decrypt the dataset with 'cryptutil'"
        # we cannot use 'runCheck' with stdin/out operations
        # WARNING the cryptutil will load all the dataset into memory in order
        # decrypt it. With big datasets (over 800Mb) that might thus not work
        # since the enclave has only 1Gb of RAM.
        /home/enclave/cryptutil decrypt --keyAndInitVal "$secret" --readData < "/home/enclave/datasets/$DATASET_FILENAME" -x > "/home/scientist/python_project/datasets/$NEW_FILENAME"
        logInfo "dataset decrypted" "dataset decrypted and saved in $NEW_FILENAME"
    done

    logInfo "getting the project instance id" "from the VmWare tool"
    PROJECT_INST_ID=$(vmtoolsd --cmd "info-get guestinfo.ovfenv" | perl -n -e '/key="project_instance_id".*value="(.*)"/ && print $1')
    logInfo "got the project instance id" "here is the project instance id: $PROJECT_INST_ID"

    # Note: ideally we should set the status of the contract at the very end.
    # However, we need the private key in order to update the contract, this is
    # why we have to do it now. In case the key folder has not been deleted, we
    # still try to set the status as unlockedErrored.
    OUR_KEY="$(cat /home/enclave/.pub_key)"
    logInfo "setting the contract to unlocked" "we set the contract to unlocked now before deleting the key that allows us to update the contract"
    runCheck /home/enclave/pcadmin contract project invoke updateStatus --status "unlockedOK" -i "$PROJECT_INST_ID" -s "$OUR_KEY"

    ## Cleaning operation...
    ## ...
    ## ...
    logInfo "cleaning" "removing the key folder"
    runCheck rm -rf /home/enclave/key
    logInfo "sanity check" "checking if the key folder has been deleted"
    if [ -d /home/enclave/key ]; then
        logError "the key folder is still there, aborting!" "found something in /home/enclave/key"
        runCheck /home/enclave/pcadmin contract project invoke updateStatus --status "unlockedErrored" -i "$PROJECT_INST_ID" -s "$OUR_KEY"
        exit 1
    fi
    
    logInfo "getting the public key to be authorized" "from the VmWare tool"
    DS_PUB_KEY=$(vmtoolsd --cmd "info-get guestinfo.ovfenv" | perl -n -e '/key="authorized_key".*value="(.*)"/ && print $1')
    logInfo "got the public key" "here is the public key to be authorized: $DS_PUB_KEY"
    logInfo "update the firewall" "we now deny by default outgoing trafic"
    logCloseOK "this will be our last message since we now block in/out trafic. bye." "since we now block all in/out trafic except ssh we can't save the log on the cloud. You will have to ssh to the enclave to have more infos."
    sudo ufw default deny outgoing
    # if the firewall update produces and error, the script exits and the pub
    # key is not added
    echo "$DS_PUB_KEY" | sudo tee -a /home/scientist/.ssh/authorized_keys
    # this a special exit that won't save logs on the cloud because everything
    # is locked down by the firewall.
    exit 10
else
    logInfo "not ready" "did not find the ready signal, doing nothing here"
fi

logInfo "things are ok" "startup.sh ends correctly."