#!/usr/bin/env bash

set -eo pipefail

echo "Initializing flightctl-cli-artifacts certificates"

source "/utils/init_utils.sh"

CERTS_SOURCE_PATH="/certs-source"
CERTS_DEST_PATH="/certs-destination"

# Wait for certificates
wait_for_files "$CERTS_SOURCE_PATH/server.crt" "$CERTS_SOURCE_PATH/server.key"

# Handle server certificates
#
# The CLI artifacts container runs as user 1001 by default,
# so we need to ensure that the server certificate and key files
# can be read by this user.
if [ -f "$CERTS_SOURCE_PATH/server.crt" ]; then
  cp "$CERTS_SOURCE_PATH/server.crt" "$CERTS_DEST_PATH/server.crt"
  chown 1001:0 "$CERTS_DEST_PATH/server.crt"
  chmod 0440 "$CERTS_DEST_PATH/server.crt"
else
  echo "Error: Server certificate not found at $CERTS_SOURCE_PATH/server.crt"
  exit 1
fi
if [ -f "$CERTS_SOURCE_PATH/server.key" ]; then
  cp "$CERTS_SOURCE_PATH/server.key" "$CERTS_DEST_PATH/server.key"
  chown 1001:0 "$CERTS_DEST_PATH/server.key"
  chmod 0440 "$CERTS_DEST_PATH/server.key"
else
  echo "Error: Server key not found at $CERTS_SOURCE_PATH/server.key"
  exit 1
fi

echo "Certificate initialization complete"

