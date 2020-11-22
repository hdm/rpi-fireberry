#!/bin/bash


RPI=$1

if [[ "$RPI" == "" ]]; then
    echo "usage: deploy.sh [rpi-address]"
    exit 1
fi

set -x

GOOS=linux GOARCH=arm GOARM=7 go build -o fireberry.bin || exit 1

ssh root@${RPI} 'rm -f /usr/local/bin/fireberry' && \
scp ./fireberry.bin root@${RPI}:/usr/local/bin/fireberry && \
scp ./fireberry.service root@${RPI}:/lib/systemd/system/fireberry.service && \
ssh root@${RPI} 'chmod +u+x /usr/local/bin/fireberry' && \
ssh root@${RPI} 'systemctl daemon-reload; systemctl enable fireberry.service; systemctl restart fireberry'