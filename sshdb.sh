#!/bin/bash
. ./proddbip.sh
ssh -i "~/.ssh/hermes_prod" -v root@${SSC_PROD_DB_IP}