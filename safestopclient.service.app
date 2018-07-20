[Unit]
Description=safestopclient
# this unit will only start after docker.service
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
# per https://www.digitalocean.com/community/tutorials/how-to-create-and-run-a-service-on-a-coreos-cluster
EnvironmentFile=/etc/environment
# before starting make sure it doesn't exist
# '=-' means it can fail
ExecStartPre=-/usr/bin/docker rm safestopclient
ExecStart=/usr/bin/docker run --rm -p 443:8443 -p 80:8080 -e SSC_MAIL_PASSWORD=${SSC_MAIL_PASSWORD} -e SSC_DB_PASSWORD=${SSC_DB_PASSWORD} --name safestopclient safestopclient:latest
ExecStop=/usr/bin/docker stop safestopclient
# restart if it crashes or is killed e.g. by oom
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target