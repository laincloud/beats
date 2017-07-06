#!/usr/bin/env bash

set -e

FPM_IMAGE_VERSION='v1.0.0'

docker run --name lain_filebeat_builder --network=host \
    -e TRAVIS_TAG=$TRAVIS_TAG \
    -w /rpmbuilder \
    -v $(pwd):/rpmbuilder laincloud/fpmbuilder:$FPM_IMAGE_VERSION \
    /rpmbuilder/script/lain-filebeat-package.sh
