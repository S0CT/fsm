#!/bin/bash
# entrypoint.sh - Dynamic PUID/PGID mapping for Unraid/Docker

PUID=${PUID:-1000}
PGID=${PGID:-1000}

echo "Starting Factorio Server Manager"
echo "Enforcing ownership with PUID: $PUID and PGID: $PGID"

# Modify the fsm user and group to match the environment variables
groupmod -o -g "$PGID" fsm
usermod -o -u "$PUID" fsm

# Ensure the /data directory is correctly owned by the target user
chown -R fsm:fsm /data

# Drop root privileges and execute the application
exec gosu fsm "$@"
