#!/usr/bin/env bash
#
# Build the self-contained SEP submission file from the split working sources.
#
# Inputs (relative to repo root):
#   sep-standards-mclip-profile.md  — front matter (preamble through "Open Issues")
#                                     and back matter ("References")
#   profile-v0.md                   — full §0–§16 normative body + Appendices
#
# Output:
#   build/sep-mclip-standards-track.md — the single self-contained SEP submission.
#
# See sep-standards-mclip-profile.md "Filing form" section for the contract.

set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p build

SEP_FRONT_AND_BACK="sep-standards-mclip-profile.md"
PROFILE_BODY="profile-v0.md"
OUT="build/sep-mclip-standards-track.md"

# Split the SEP file at the "## References" header into front and back.
FRONT_END_LINE=$(grep -n '^## References$' "$SEP_FRONT_AND_BACK" | head -n1 | cut -d: -f1)
if [ -z "$FRONT_END_LINE" ]; then
  echo "ERROR: could not locate '## References' header in $SEP_FRONT_AND_BACK" >&2
  exit 1
fi

# Front matter: lines 1 .. (FRONT_END_LINE - 1)
head -n "$((FRONT_END_LINE - 1))" "$SEP_FRONT_AND_BACK" > "$OUT"
echo "" >> "$OUT"

# Normative body: full profile-v0.md.
# Strip its document-level H1 since the SEP already has one.
sed -e '1{/^# /d;}' "$PROFILE_BODY" >> "$OUT"
echo "" >> "$OUT"

# Back matter: from "## References" to end of SEP file.
tail -n "+$FRONT_END_LINE" "$SEP_FRONT_AND_BACK" >> "$OUT"

echo "Built: $OUT"
wc -l "$OUT"
