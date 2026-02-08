# RHEL System Metrics Collectors

Go-based metric collectors for RHEL systems (RHEL 7-10 compatible).

## Overview

This package contains three metric collectors that expose system information in Prometheus format:

1. **dnf_last_update** - Tracks the last package installation/update time
2. **dnf_update_check** - Monitors available system updates including kernel versions (current + pending)
3. **chrony_sources** - Monitors NTP time synchronization status

All binaries are designed to be triggered by Telegraf's exec plugin.

**Designed by Ifesinachi Osude**

## Version

Current version: `v1.0.0` (tracked in `../VERSION`)

## Building

### Prerequisites

- Go compiler (any recent version will work)
- RHEL/CentOS/Rocky/AlmaLinux 7-10
- Target architecture: x86_64 (amd64)

### Build Instructions

Each binary is built independently with static linking for maximum compatibility:

```bash
# Build all binaries
./build_all.sh

# Or build individually:
cd go_build/dnf_last_update && ./build.sh
cd go_build/dnf_update_check && ./build.sh
cd go_build/chrony_sources && ./build.sh
```

### Build Output

Binaries are created in each subdirectory:
- `go_build/dnf_last_update/dnf_last_update`
- `go_build/dnf_update_check/dnf_update_check`
- `go_build/chrony_sources/chrony_sources`

### Static Compilation

All binaries are compiled with:
- CGO disabled (pure Go, no C dependencies)
- Static linking
- Optimized for size
- RHEL 7-10 compatibility

Build command used:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -X main.version=$(cat ../../VERSION)" \
  -o <binary_name> main.go
```

## Installation

1. Build the binaries (see above)
2. Copy binaries to a suitable location (e.g., `/usr/local/bin/`)
3. Make them executable
4. Configure Telegraf (see below)

```bash
sudo cp go_build/*/dnf_* /usr/local/bin/
sudo cp go_build/chrony_sources/chrony_sources /usr/local/bin/
sudo chmod +x /usr/local/bin/{dnf_last_update,dnf_update_check,chrony_sources}
```

## Telegraf Configuration

Copy the provided telegraf configuration:

```bash
sudo cp rhel_stats.conf /etc/telegraf/telegraf.d/
sudo systemctl restart telegraf
```

The configuration uses Telegraf's exec plugin to run each binary and collect metrics in Prometheus format.

## Metrics Exposed

### dnf_last_update

- `dnf_last_update` - Unix timestamp of last package installation
- `dnf_last_update_info{time="..."}` - Human-readable time label
- `dnf_last_update_error` - 1 if unable to determine, 0 otherwise

### dnf_update_check

- `dnf_update_available` - 1 if updates available, 0 otherwise
- `dnf_update_pending_count` - Number of pending updates
- `dnf_update_kernel_available` - 1 if kernel update available
- `dnf_update_kernel_current_version_info{version="..."}` - Currently running kernel version
- `dnf_update_kernel_pending_version_info{version="..."}` - Pending kernel version (if available)
- `dnf_update_security_available` - 1 if security updates available
- `dnf_update_security_count` - Number of security updates
- `dnf_update_reboot_required` - 1 if reboot needed
- `dnf_update_check_error` - 1 if check failed
- `dnf_update_pending_pkg{name,arch,version,repo}` - Per-package details
- `dnf_update_pending_info{packages="..."}` - Truncated package list

### chrony_sources

- `chrony_sources_up` - 1 if chrony is available, 0 otherwise
- `chrony_source_selected{source,mode,stratum,poll}` - 1 if selected (*)
- `chrony_source_in_use{source,mode,stratum,poll}` - 1 if in use (+)
- `chrony_source_reachable{source,mode,stratum,poll}` - 1 if reachable
- `chrony_source_reach{source,mode,stratum,poll}` - Reach register value
- `chrony_source_last_rx_seconds{source,mode,stratum,poll}` - Seconds since last sample
- `chrony_source_offset_seconds{source,mode,stratum,poll}` - Time offset in seconds
- `chrony_source_jitter_seconds{source,mode,stratum,poll}` - Jitter in seconds

## Grafana Integration

All metrics are exposed in Prometheus format. Example PromQL queries:

### Display Current and Pending Kernel Versions
```promql
# Current kernel
dnf_update_kernel_current_version_info

# Pending kernel (if update available)
dnf_update_kernel_pending_version_info{version!=""}
```

### Updates Dashboard
```promql
# Total pending updates
dnf_update_pending_count

# Security updates
dnf_update_security_count

# Reboot required
dnf_update_reboot_required
```

### Time Since Last Update
```promql
# Days since last update
(time() - dnf_last_update) / 86400
```

## Requirements

See `requirements.txt` for system dependencies.

## Build Tracking

See `BUILD_TRACKER.md` for build history and changes.

## Compatibility

- **OS**: RHEL/CentOS/Rocky/AlmaLinux 7, 8, 9, 10
- **Architecture**: x86_64 (amd64)
- **Go Version**: Any (binaries are statically compiled)
- **Dependencies**: None (static binaries)

## Troubleshooting

### Testing Binaries

```bash
# Test each binary
/usr/local/bin/dnf_last_update
/usr/local/bin/dnf_update_check
/usr/local/bin/chrony_sources
```

### Telegraf Testing

```bash
# Test telegraf config
telegraf --test --config /etc/telegraf/telegraf.d/rhel_stats.conf

# Check logs
sudo journalctl -u telegraf -f
```

### Common Issues

1. **Permission denied**: Ensure binaries are executable
2. **Command not found**: Verify binary paths in telegraf config
3. **No metrics**: Check that Prometheus output is enabled in main telegraf.conf

## License

**Designed by Ifesinachi Osude**

Version tracking via `VERSION` file in repository root.
