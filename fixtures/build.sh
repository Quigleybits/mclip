#!/usr/bin/env bash
# Build all 9 fixture servers (POSIX).
#
# Flag rationale:
#   -trimpath           : reproducible builds, strips local path leaks.
#   -ldflags "-s -w"    : strips symbol + DWARF tables (~30% smaller binaries).
set -euo pipefail
servers=(fx-echo fx-flag-collisions fx-destructive fx-resources-watch
         fx-http-auth fx-http-error-data fx-pagination fx-prompts fx-progress)
root="$(cd "$(dirname "$0")" && pwd)"
for s in "${servers[@]}"; do
    echo "==> $s"
    ( cd "$root/servers/$s" && go build -trimpath -ldflags="-s -w" -o "$s" . )
done
echo "All 9 fixture servers built."
