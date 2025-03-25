#!/usr/bin/env bash

set -eo pipefail

# Load secret generation functions
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
source "${SCRIPT_DIR}"/secrets.sh

write_default_base_domain() {
    # Write base domain to the config file
    base_domain="$(ip route get 1.1.1.1 | grep -oP 'src \K\S+')"
    echo "Setting base domain to: ${base_domain}"
    VALUES_FILE="/etc/flightctl/values.yaml"
    sed -i "s/^\(\s*baseDomain\s*\):\s*.*$/\1: ${base_domain}/" "${VALUES_FILE}"
}

main() {
    echo "Configuring Flight Control post install"

    ensure_secrets
    write_default_base_domain

    echo "Post install configuration complete"
}

main
