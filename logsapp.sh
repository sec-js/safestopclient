#!/bin/bash

. ./prodappip.sh

ssh -i ~/.ssh/hermes_prod core@${SSC_PROD_APP_IP} <<'ENDSSH'
cd /home/core
docker logs -f safestopclient
ENDSSH