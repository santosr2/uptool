#!/bin/bash
# Update SECURITY.md supported versions table
# Usage: ./scripts/update-security-versions.sh <new-version>
# Example: ./scripts/update-security-versions.sh v0.3.0

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <new-version>"
    echo "Example: $0 v0.3.0"
    exit 1
fi

NEW_VERSION="$1"
SECURITY_FILE="SECURITY.md"

# Remove 'v' prefix if present
NEW_VERSION="${NEW_VERSION#v}"

# Extract major and minor version (e.g., 0.3.0 -> 0.3)
MAJOR=$(echo "$NEW_VERSION" | cut -d. -f1)
MINOR=$(echo "$NEW_VERSION" | cut -d. -f2)
NEW_MINOR_VERSION="${MAJOR}.${MINOR}"

# Calculate previous minor version (e.g., 0.3 -> 0.2)
PREV_MINOR=$((MINOR - 1))
PREV_MINOR_VERSION="${MAJOR}.${PREV_MINOR}"

# Calculate date 6 months from now for new security support window
CURRENT_DATE=$(date +"%Y-%m")
SIX_MONTHS_LATER=$(date -d "+6 months" +"%Y-%m" 2>/dev/null || date -v+6m +"%Y-%m")

echo "Updating SECURITY.md supported versions table..."
echo "  New version: ${NEW_MINOR_VERSION}.x"
echo "  Current date: $CURRENT_DATE"

# Extract current table and process each version
ACTIVE_SECURITY_VERSIONS=()
EXPIRED_VERSIONS=()

# Find all versions currently under security support (⚠️)
while IFS= read -r line; do
    # Extract version number (e.g., "| 0.1.x   | :warning: | 2025-12 (6 months) |")
    if [[ "$line" =~ \|[[:space:]]*([0-9]+\.[0-9]+)\.x[[:space:]]*\|[[:space:]]*:warning:[[:space:]]*\|[[:space:]]*([0-9]{4}-[0-9]{2}) ]]; then
        VERSION="${BASH_REMATCH[1]}"
        DEADLINE="${BASH_REMATCH[2]}"

        # Check if deadline has passed
        if [[ "$CURRENT_DATE" > "$DEADLINE" ]]; then
            echo "  Version ${VERSION}.x support expired ($DEADLINE)"
            EXPIRED_VERSIONS+=("$VERSION")
        else
            echo "  Version ${VERSION}.x still in security window (until $DEADLINE)"
            ACTIVE_SECURITY_VERSIONS+=("${VERSION}|${DEADLINE}")
        fi
    fi
done < <(grep -E "^\|.*:warning:" "$SECURITY_FILE" || true)

# Find the current stable version (✅) - this becomes security-only
CURRENT_STABLE=""
if grep -q ":white_check_mark:" "$SECURITY_FILE"; then
    CURRENT_STABLE=$(awk -F'|' '/white_check_mark/ { gsub(/[[:space:]]/, "", $2); gsub(/\.x/, "", $2); print $2; exit }' "$SECURITY_FILE")
    if [ -n "$CURRENT_STABLE" ]; then
        echo "  Moving ${CURRENT_STABLE}.x to security-only support (until $SIX_MONTHS_LATER)"
    fi
fi

# Build new table (data rows only, no header)
{
    # New stable version
    echo "| ${NEW_MINOR_VERSION}.x   | :white_check_mark: | Current stable version |"

    # Current stable becomes security-only (if it exists and isn't the same as new version)
    if [ -n "$CURRENT_STABLE" ] && [ "$CURRENT_STABLE" != "$NEW_MINOR_VERSION" ]; then
        echo "| ${CURRENT_STABLE}.x   | :warning:          | ${SIX_MONTHS_LATER} (6 months) |"
    fi

    # Keep any other versions still in their security window
    if [ "${#ACTIVE_SECURITY_VERSIONS[@]:-0}" -gt 0 ]; then
        for entry in "${ACTIVE_SECURITY_VERSIONS[@]}"; do
            VERSION="${entry%%|*}"
            DEADLINE="${entry##*|}"
            echo "| ${VERSION}.x   | :warning:          | ${DEADLINE} (6 months) |"
        done
    fi

    # Calculate what to show for unsupported versions
    # Find the oldest version that's no longer supported
    OLDEST_SUPPORTED=""
    if [ "${#ACTIVE_SECURITY_VERSIONS[@]:-0}" -gt 0 ]; then
        # Find the minimum version from active security versions
        OLDEST_SUPPORTED=$(printf '%s\n' "${ACTIVE_SECURITY_VERSIONS[@]}" | cut -d'|' -f1 | sort -V | head -n1)
    elif [ -n "$CURRENT_STABLE" ]; then
        OLDEST_SUPPORTED="$CURRENT_STABLE"
    else
        OLDEST_SUPPORTED="$NEW_MINOR_VERSION"
    fi

    echo "| < ${OLDEST_SUPPORTED}   | :x:                | No support    |"

} > /tmp/security_versions.tmp

# Use awk to replace the table in SECURITY.md
# Find the table start and replace until the empty line after the table
awk '
/\| Version \| Supported/ {
    # Print the header
    print
    # Read and print the separator line
    getline
    print
    # Insert new table rows from temp file
    while ((getline line < "/tmp/security_versions.tmp") > 0) {
        if (line ~ /^\|/) {
            print line
        }
    }
    close("/tmp/security_versions.tmp")
    # Skip old table rows until we hit an empty line
    while (getline && $0 ~ /^\|/) {
        # Skip old rows
    }
    # Print the empty line after the table
    print
    next
}
{ print }
' "$SECURITY_FILE" > /tmp/security_updated.md

# Replace the original file
mv /tmp/security_updated.md "$SECURITY_FILE"

# Also update the "currently X.Y.x" text in the support policy
sed -i.bak "s/currently [0-9]\+\.[0-9]\+\.x/currently ${NEW_MINOR_VERSION}.x/" "$SECURITY_FILE"
rm -f "${SECURITY_FILE}.bak"

# Clean up
rm -f /tmp/security_versions.tmp

echo "✅ Updated SECURITY.md successfully"
echo ""
echo "Changes:"
git diff "$SECURITY_FILE" || true
