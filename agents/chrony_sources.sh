#!/usr/bin/env bash
set -euo pipefail

# Prometheus exposition for chrony sources
# Requires: chronyc

# Escape for Prometheus label values
esc() {
  sed -e 's/\\/\\\\/g' \
      -e 's/"/\\"/g' \
      -e 's/\r//g' \
      -e 's/\n/\\n/g'
}

# If chronyc isn't available, emit a single "up" metric and exit
if ! command -v chronyc >/dev/null 2>&1; then
  echo "# TYPE chrony_sources_up gauge"
  echo "chrony_sources_up 0"
  exit 0
fi

# Get sources (suppress chronyc banners)
# We parse the "chronyc sources -v" table.
out="$(chronyc -n sources -v 2>/dev/null || true)"

echo "# TYPE chrony_sources_up gauge"
if [ -z "$out" ]; then
  echo "chrony_sources_up 0"
  exit 0
fi
echo "chrony_sources_up 1"

echo "# HELP chrony_source_selected 1 if this source is selected (*), else 0"
echo "# TYPE chrony_source_selected gauge"
echo "# HELP chrony_source_in_use 1 if this source is being combined/used (+), else 0"
echo "# TYPE chrony_source_in_use gauge"
echo "# HELP chrony_source_reachable 1 if reachable (reach>0), else 0"
echo "# TYPE chrony_source_reachable gauge"
echo "# HELP chrony_source_reach Reach register (octal shown by chrony, emitted as integer 0-377)"
echo "# TYPE chrony_source_reach gauge"
echo "# HELP chrony_source_last_rx_seconds Seconds since last sample"
echo "# TYPE chrony_source_last_rx_seconds gauge"
echo "# HELP chrony_source_offset_seconds Reported offset in seconds"
echo "# TYPE chrony_source_offset_seconds gauge"
echo "# HELP chrony_source_jitter_seconds Reported jitter in seconds"
echo "# TYPE chrony_source_jitter_seconds gauge"

# Parse lines that look like:
# ^* 10.0.0.1  2  10  377   32  -120us  500us
# First field: MS (2 chars), then: Name/IP, Stratum, Poll, Reach, LastRx, Last sample (offset), +/- (optional), RMS offset/jitter
#
# We'll focus on: MS, name, stratum, poll, reach, lastRx, offset, jitter
# Offset/jitter unit can be ns/us/ms/s; convert to seconds.
printf '%s\n' "$out" | awk '
function unit_to_seconds(v,  num, unit) {
  # v like -120us or 1.2ms or 0.0003s
  num=v
  unit=""
  if (match(v, /[a-z]+$/)) { unit=substr(v, RSTART, RLENGTH); num=substr(v, 1, RSTART-1) }
  if (num == "") num=0
  # handle possible "+/-" tokens already removed upstream
  if (unit=="ns") return num/1000000000.0
  if (unit=="us") return num/1000000.0
  if (unit=="ms") return num/1000.0
  # s or empty
  return num+0.0
}
BEGIN { OFS="\t" }
# lines starting with ^ or = are the source table
$1 ~ /^[\^=]/ {
  ms=$1
  name=$2
  stratum=$3
  poll=$4
  reach=$5
  lastrx=$6

  # offset usually at $7, jitter often at last column
  # Some chrony versions include "Last sample" with two tokens like: "-120us[ -200us] +/- 500us"
  # With -v output, columns usually simplify to offset and jitter; but we make it robust:
  offset_raw=$7
  jitter_raw=$(NF)

  # Clean common noise: remove brackets if present
  gsub(/\[/,"",offset_raw); gsub(/\]/,"",offset_raw)
  gsub(/\[/,"",jitter_raw); gsub(/\]/,"",jitter_raw)

  # Selected/in use flags
  sel = (ms ~ /\*/) ? 1 : 0
  use = (ms ~ /\+/) ? 1 : 0

  # Reachability
  reach_i = reach + 0
  reachable = (reach_i > 0) ? 1 : 0

  # Convert offset/jitter to seconds
  offset_s = unit_to_seconds(offset_raw)
  jitter_s = unit_to_seconds(jitter_raw)

  print ms, name, stratum, poll, reach_i, lastrx+0, offset_s, jitter_s, sel, use, reachable
}' | while IFS=$'\t' read -r ms name stratum poll reach lastrx offset_s jitter_s sel use reachable; do
  name_esc="$(printf '%s' "$name" | esc)"
  ms_esc="$(printf '%s' "$ms" | esc)"

  # Labels
  labels="source=\"${name_esc}\",mode=\"${ms_esc}\",stratum=\"${stratum}\",poll=\"${poll}\""

  echo "chrony_source_selected{${labels}} ${sel}"
  echo "chrony_source_in_use{${labels}} ${use}"
  echo "chrony_source_reachable{${labels}} ${reachable}"
  echo "chrony_source_reach{${labels}} ${reach}"
  echo "chrony_source_last_rx_seconds{${labels}} ${lastrx}"
  echo "chrony_source_offset_seconds{${labels}} ${offset_s}"
  echo "chrony_source_jitter_seconds{${labels}} ${jitter_s}"
done