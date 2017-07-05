#!/usr/bin/env bash

set -e

RPM_VERSION='temp'
if [ "$TRAVIS_TAG" != "" ]; then
    RPM_VERSION="$TRAVIS_TAG"
fi
BEAT_NAME='filebeat'
wget https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-5.4.0-x86_64.rpm
rpm -ivp filebeat-5.4.0-x86_64.rpm
mkdir -p $GOPATH/src/github.com/elastic
ln -s /rpmbuilder $GOPATH/src/github.com/elastic/beats
go build -o /usr/share/filebeat/bin/filebeat github.com/elastic/beats/filebeat
echo "#!/bin/bash" > /tmp/systemd-daemon-reload.sh
echo "systemctl daemon-reload 2> /dev/null || true" >> /tmp/systemd-daemon-reload.sh
echo "#!/bin/bash" > /tmp/stop-filebeat.sh
echo "systemctl stop filebeat 2> /dev/null || service filebeat stop 2>/dev/null || true" >> /tmp/stop-filebeat.sh
fpm --force -s dir -t rpm \
        -n ${BEAT_NAME} -v ${RPM_VERSION} \
        --architecture x86_64 \
        --license "MIT"  \
        --url "https://github.com/laincloud/beats" \
        --description "LAIN extended filebeat" \
        --rpm-init /etc/init.d/${BEAT_NAME} \
        --after-install /tmp/systemd-daemon-reload.sh \
        --before-remove /tmp/stop-filebeat.sh \
        --after-remove /tmp/systemd-daemon-reload.sh \
        --config-files /etc/${BEAT_NAME}/${BEAT_NAME}.yml \
        /usr/bin/${BEAT_NAME}.sh \
        /usr/share/${BEAT_NAME} \
        /etc/${BEAT_NAME} \
        /lib/systemd/system/${BEAT_NAME}.service
