#!/usr/bin/env bash

set -e

FPM_IMAGE_VERSION='v1.0.0'

docker pull laincloud/fpmbuilder:$FPM_IMAGE_VERSION
docker create --name lain_filebeat_builder --net=host \
    -e TRAVIS_TAG=$TRAVIS_TAG \
    -w /rpmbuilder \
    -v $(pwd):/rpmbuilder laincloud/fpmbuilder:$FPM_IMAGE_VERSION \
    /rpmbuilder/script/lain-filebeat-build.sh
