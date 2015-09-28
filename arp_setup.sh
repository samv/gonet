#!/usr/bin/env bash

set -e

echo '00:34:45:de:ca:de' > ethernet/external_mac.static
echo '10.0.0.3' > ipv4/external_ip.static
echo '10.0.0.2' > ipv4/external_gateway.static
