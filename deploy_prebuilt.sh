#!/bin/bash

#https://blog.kowalczyk.info/article/5/Blueprint-for-deploying-web-apps-on-CoreOS.html
#https://coderwall.com/p/fkfaqq/safer-bash-scripts-with-set-euxo-pipefail

set -u -e -o pipefail

. ./prodappip.sh

dir=`pwd`
blog_dir=${GOPATH}/src/github.com/schoolwheels/safestopclient

echo "building"
go get golang.org/x/sys/unix
GOOS=linux GOARCH=amd64 go build -o safestopclient
docker build --no-cache --tag safestopclient:latest -f Dockerfile-prebuilt .
rm safestopclient
cd "${dir}"

echo "docker save"
docker save safestopclient:latest | bzip2 > safestopcliente-latest.tar.bz2
ls -lah safestopclient-latest.tar.bz2

echo "uploading to the server"
scp -i ~/.ssh/hermes_prod safestopclient-latest.tar.bz2 core@${SSC_PROD_APP_IP}:/home/core/safestopclient-latest.tar.bz2

echo "extracting on the server"
ssh -i ~/.ssh/hermes_prod core@${SSC_PROD_APP_IP} <<'ENDSSH'
cd /home/core
bunzip2 --stdout safestopclient-latest.tar.bz2 | docker load
rm safestopclient-latest.tar.bz2
sudo systemctl restart safestopclient
ENDSSH

rm -rf safestopclient-latest.tar.bz2