#!/usr/bin/env bash
set -euo pipefail

# Prometheus label escaping
esc_label() {
  sed -e 's/\\/\\\\/g' \
      -e 's/"/\\"/g' \
      -e 's/\r//g' \
      -e 's/\n/\\n/g'
}

# rpm -qa --last prints newest first on RHEL (most recent installs first)
# Example line:
#   bash-4.4.20-...   Mon 13 Jan 2026 10:22:11 AM EST
#
# We'll take the FIRST package line and parse the date portion.
line="$(rpm -qa --last 2>/dev/null | head -n 1 || true)"

# If rpm output is empty/unexpected, export 0 and error=1
if [ -z "${line:-}" ]; then
  echo "# TYPE dnf_last_update gauge"
  echo "dnf_last_update 0"
  echo "# TYPE dnf_last_update_error gauge"
  echo "dnf_last_update_error 1"
  exit 0
fi

# Remove the package name (first column) and normalize whitespace
datestr="$(printf '%s\n' "$line" | awk '{$1=""; sub(/^[ \t]+/,""); print}')"

# Convert to epoch seconds (works on GNU date)
epoch="0"
if epoch="$(date -d "$datestr" +%s 2>/dev/null)"; then
  : # ok
else
  epoch="0"
fi

err=0
if [ "$epoch" = "0" ]; then
  err=1
fi

datestr_esc="$(printf '%s' "$datestr" | esc_label)"

echo "# TYPE dnf_last_update gauge"
echo "dnf_last_update $epoch"

# Optional human-readable info metric
echo "# TYPE dnf_last_update_info gauge"
echo "dnf_last_update_info{time=\"${datestr_esc}\"} 1"

echo "# TYPE dnf_last_update_error gauge"
echo "dnf_last_update_error $err"