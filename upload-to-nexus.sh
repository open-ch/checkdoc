#!/usr/bin/env bash

DEST_HOSTNAME="https://repo-java.open.ch"
DEST_REPOSITORY="lake-snapshots"
DEST_PATH="tools/checkdoc/checkdoc-darwin-amd64-${VERSION}"

FULL_DEST_URL="${DEST_HOSTNAME}/repository/${DEST_REPOSITORY}/${DEST_PATH}"

SOURCE_PATH=$(bazel run --run_under "echo " //tools/checkdoc:cli)

echo "Copying ${SOURCE_PATH} to ${FULL_DEST_URL}"

curl -v -u "${DEPLOY_RPM_USERNAME}:${DEPLOY_RPM_PASSWORD}" --upload-file "${SOURCE_PATH}" "${FULL_DEST_URL}"
