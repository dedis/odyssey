#!/bin/sh
# this file is a utility script that we launch before creating a new template
# image. This script ensures no unwanted files are left behind.

rm /home/enclave/.first
rm /home/enclave/.endpoint
rm /home/enclave/bc-file.cfg
rm /home/enclave/cron.log
rm -rf /home/enclave/key
rm /home/enclave/rclogs.log
rm /tmp/startup_logInfo
rm /tmp/startup_msg
rm /tmp/startup_runCheck_log
rm /home/enclave/.pub_key