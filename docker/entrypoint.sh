#!/usr/bin/env bash
set -e

# Create data/config dirs if they don't exist (handles fresh volumes)
mkdir -p "$(dirname "$GOSUKI_DB")" "$GOSUKI_CONFIG_DIR"

# Generate default config if it doesn't exist
if [ ! -f "$GOSUKI_CONFIG" ]; then
    echo "generating default config to $GOSUKI_CONFIG"
    gosuki config gen > "$GOSUKI_CONFIG" 2>/dev/null || true
fi

exec gosuki -c "$GOSUKI_CONFIG" --db="$GOSUKI_DB" -l 0.0.0.0:2025 "$@"
