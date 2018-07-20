#!/bin/bash
set -u -e -o pipefail

. ./prodappip.sh

scp -i ~/.ssh/hermes_prod ./safestopclient.service core@${SSC_PROD_APP_IP}:/home/core/safestopclient.service

ssh -i ~/.ssh/hermes_prod core@${SSCC_PROD_APP_IP} <<'ENDSSH'
cd /home/core
sudo cp safestopclient.service /etc/systemd/system
sudo systemctl enable /etc/systemd/system/safestopclient.service
rm safestopclient.service
ENDSSH