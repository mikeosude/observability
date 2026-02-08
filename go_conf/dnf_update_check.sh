#!/usr/bin/env bash
set -euo pipefail

# ----------------------------
# Simple DNF/YUM update check
# Prometheus exposition output
# ----------------------------

# Hard timeouts to prevent hanging (adjust if needed)
TO_CHECK="30s"
TO_SEC="25s"
TO_REBOOT="10s"

# Prefer dnf; fallback to yum
PKG="dnf"
command -v dnf >/dev/null 2>&1 || PKG="yum"

# Escape for Prometheus label values
esc_label() {
  sed -e 's/\\/\\\\/g' \
      -e 's/"/\\"/g' \
      -e 's/\r//g' \
      -e 's/\n/\\n/g'
}

# Defaults
updates_available=0
pending_count=0
kernel_available=0
kernel_version=""
security_available=0
security_count=0
reboot_required=0
check_error=0
pending_list=""

# Holds the actual package list output from check-update
CHECK_OUT=""

# ----------------------------
# 1) Are updates available? + capture output for parsing
# dnf/yum check-update exit codes:
# 0 = none, 100 = updates, other = error
# ----------------------------
set +e
CHECK_OUT="$(timeout "$TO_CHECK" "$PKG" -q check-update 2>/dev/null)"
rc=$?
set -e

if [ "$rc" -eq 100 ]; then
  updates_available=1
elif [ "$rc" -eq 0 ]; then
  updates_available=0
else
  # includes timeouts (124) and other failures
  check_error=1
fi

# ----------------------------
# 2) Parse updates list from check-update output (most reliable)
# Expected (dnf + yum typical):
#   pkgname.arch   version-release   repo
# Example:
#   binutils.x86_64 2.30-128.el8_10  rhel-8-for-x86_64-baseos-rpms
#
# Stop parsing when "Obsoleting Packages" starts (dnf)
# ----------------------------
if [ "$updates_available" -eq 1 ] && [ "$check_error" -eq 0 ] && [ -n "${CHECK_OUT:-}" ]; then
  # Extract only package lines: at least 3 fields, first field contains a dot (name.arch)
  # and ignore any dnf section headers.
  # Also stop at "Obsoleting" section.
  PARSED="$(printf '%s\n' "$CHECK_OUT" \
    | awk '
        BEGIN{obs=0}
        /^Obsoleting Packages/{obs=1}
        obs==1 {next}
        NF>=3 && $1 ~ /\./ {print $1"\t"$2"\t"$3}
      ')"

  if [ -n "${PARSED:-}" ]; then
    pending_count="$(printf '%s\n' "$PARSED" | awk 'NF{c++} END{print c+0}')"

    # Build pending_list as: name.arch-version (comma separated)
    pending_list="$(printf '%s\n' "$PARSED" | awk -F'\t' '{print $1"-"$2}' | tr '\n' ',' | sed 's/,$//')"

    # Kernel version (prefer kernel-core)
    kernel_version="$(printf '%s\n' "$PARSED" | awk -F'\t' '$1 ~ /^kernel-core\./ {print $2; exit}')"
    if [ -z "${kernel_version:-}" ]; then
      kernel_version="$(printf '%s\n' "$PARSED" | awk -F'\t' '$1 ~ /^kernel\./ {print $2; exit}')"
    fi
  else
    # Updates exist but we couldn't parse them (rare). Mark error so you can alert on it.
    check_error=1
  fi
fi

# Kernel update available?
if [ -n "${kernel_version:-}" ]; then
  kernel_available=1
fi

# ----------------------------
# 3) Security updates count (dnf only; yum may not support)
# ----------------------------
if [ "$PKG" = "dnf" ]; then
  set +e
  sec_out="$(timeout "$TO_SEC" dnf -q updateinfo list security 2>/dev/null)"
  rc3=$?
  set -e

  if [ "$rc3" -eq 0 ] && [ -n "${sec_out:-}" ]; then
    security_count="$(printf '%s\n' "$sec_out" | awk 'NF{c++} END{print c+0}')"
    if [ "${security_count:-0}" -gt 0 ]; then
      security_available=1
    else
      security_available=0
      security_count=0
    fi
  else
    security_available=0
    security_count=0
  fi
fi

# ----------------------------
# 4) Reboot required?
# Prefer needs-restarting -r (dnf-utils/yum-utils). Protect with timeout.
# ----------------------------
if command -v needs-restarting >/dev/null 2>&1; then
  set +e
  timeout "$TO_REBOOT" needs-restarting -r >/dev/null 2>&1
  rcr=$?
  set -e
  if [ "$rcr" -ne 0 ]; then
    reboot_required=1
  fi
else
  # Fallback: running kernel != newest installed kernel-core
  running_k="$(uname -r || true)"
  newest_k="$(rpm -q --last kernel-core 2>/dev/null | head -1 | awk '{print $1}' | sed 's/^kernel-core-//' || true)"
  if [ -n "${newest_k:-}" ] && [ -n "${running_k:-}" ] && [[ "$running_k" != "$newest_k"* ]]; then
    reboot_required=1
  fi
fi

# ----------------------------
# Prometheus exposition (dnf_update_*)
# ----------------------------
echo "# TYPE dnf_update_available gauge"
echo "dnf_update_available $updates_available"

echo "# TYPE dnf_update_pending_count gauge"
echo "dnf_update_pending_count $pending_count"

echo "# TYPE dnf_update_kernel_available gauge"
echo "dnf_update_kernel_available $kernel_available"

# Kernel version info (label). Value=1 when kernel update exists else 0.
kv_esc="$(printf '%s' "$kernel_version" | esc_label)"
echo "# TYPE dnf_update_kernel_version_info gauge"
if [ "$kernel_available" -eq 1 ]; then
  echo "dnf_update_kernel_version_info{version=\"${kv_esc}\"} 1"
else
  echo "dnf_update_kernel_version_info{version=\"\"} 0"
fi

echo "# TYPE dnf_update_security_available gauge"
echo "dnf_update_security_available $security_available"

echo "# TYPE dnf_update_security_count gauge"
echo "dnf_update_security_count $security_count"

echo "# TYPE dnf_update_reboot_required gauge"
echo "dnf_update_reboot_required $reboot_required"

echo "# TYPE dnf_update_check_error gauge"
echo "dnf_update_check_error $check_error"

# ----------------------------
# Table-friendly per-package metric:
# dnf_update_pending_pkg{name="",arch="",version="",repo=""} 1
# ----------------------------
echo "# TYPE dnf_update_pending_pkg gauge"
if [ "$updates_available" -eq 1 ] && [ "$check_error" -eq 0 ] && [ -n "${PARSED:-}" ]; then
  printf '%s\n' "$PARSED" | while IFS=$'\t' read -r namearch ver repo; do
    [ -z "${namearch:-}" ] && continue
    name="${namearch%.*}"
    arch="${namearch##*.}"

    n_esc="$(printf '%s' "$name" | esc_label)"
    a_esc="$(printf '%s' "$arch" | esc_label)"
    v_esc="$(printf '%s' "$ver"  | esc_label)"
    r_esc="$(printf '%s' "$repo" | esc_label)"

    echo "dnf_update_pending_pkg{name=\"${n_esc}\",arch=\"${a_esc}\",version=\"${v_esc}\",repo=\"${r_esc}\"} 1"
  done
fi

# ----------------------------
# Info metric with full list (still useful for quick viewing)
# NOTE: truncate to reduce label churn
# ----------------------------
pending_trunc="$(printf '%s' "$pending_list" | cut -c1-1200 | esc_label)"
echo "# TYPE dnf_update_pending_info gauge"
echo "dnf_update_pending_info{packages=\"${pending_trunc}\"} 1"