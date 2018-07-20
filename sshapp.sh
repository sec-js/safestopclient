#!/bin/bash
. ./prodappip.sh
ssh -i "~/.ssh/hermes_prod" -v core@${SSC_PROD_APP_IP}