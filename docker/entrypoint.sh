#!/bin/sh
set -e

echo "Starting Unbound DNS server..."

# Setup unbound-control keys if not exist
if [ ! -f /etc/unbound/unbound_control.key ]; then
    echo "Generating unbound-control keys..."
    unbound-control-setup
fi

# Start unbound in background
unbound -d &
UNBOUND_PID=$!

# Wait for unbound to be ready
sleep 2

echo "Starting Unbound UI..."
/usr/local/bin/unbound-ui &
UI_PID=$!

# Handle shutdown
trap "kill $UNBOUND_PID $UI_PID; exit 0" SIGTERM SIGINT

# Wait for either process to exit
wait -n $UNBOUND_PID $UI_PID
exit $?
