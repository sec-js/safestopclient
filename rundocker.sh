#!/bin/bash
docker build -t "safestopclient" .
docker run -p 80:8080 -e SSC_ENV=development --name safestopclient --rm safestopclient