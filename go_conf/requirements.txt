# RHEL Stats Collectors - System Requirements

## Build Requirements

### Build Tools
- Go compiler (1.16 or later recommended, but any recent version works)
- bash shell
- Standard UNIX utilities: cat, ls, chmod

### Build Dependencies
None - binaries are statically compiled with CGO_ENABLED=0

## Runtime Requirements

### Operating System
- RHEL 7, 8, 9, or 10
- CentOS 7, 8
- Rocky Linux 8, 9
- AlmaLinux 8, 9
- Oracle Linux 7, 8, 9
- Architecture: x86_64 (amd64)

### System Commands Required

#### For dnf_last_update
- `rpm` - Package manager query tool (pre-installed on all RHEL systems)

#### For dnf_update_check
- `dnf` or `yum` - Package manager
- `rpm` - Package manager query tool
- `uname` - Kernel version detection
- Optional: `needs-restarting` - Part of dnf-utils/yum-utils package

#### For chrony_sources
- `chronyc` - Chrony NTP client command-line tool
  - Package: `chrony`
  - Install: `sudo dnf install chrony` or `sudo yum install chrony`

### Telegraf Requirements
- Telegraf 1.0 or later
- Prometheus output plugin enabled
- Exec input plugin support (enabled by default)

## Optional Packages

### For Enhanced Functionality
- `dnf-utils` (RHEL 8+) or `yum-utils` (RHEL 7)
  - Provides `needs-restarting` command for better reboot detection
  - Install: `sudo dnf install dnf-utils` or `sudo yum install yum-utils`

### For Time Synchronization Monitoring
- `chrony` - NTP client
  - Required for chrony_sources collector
  - Install: `sudo dnf install chrony` or `sudo yum install chrony`
  - Enable: `sudo systemctl enable --now chronyd`

## Permissions

### Required Permissions
- Read access to RPM database (`/var/lib/rpm/`)
- Execute permissions for system commands (rpm, dnf/yum, uname, chronyc)
- No root/sudo required for metric collection

### Installation Permissions
- Root/sudo required to:
  - Copy binaries to `/usr/local/bin/`
  - Copy telegraf config to `/etc/telegraf/telegraf.d/`
  - Restart telegraf service

## Network Requirements
- None (all metrics collected locally)
- Prometheus scrape endpoint (Telegraf exposes metrics on configured port, typically :9273)

## Disk Space
- Binaries: ~5-10 MB total (compressed, statically linked)
- No additional disk space required for operation

## Memory
- Minimal: ~5-20 MB per binary execution
- Telegraf manages process lifecycle

## CPU
- Minimal impact: collectors run on intervals (1m-15m)
- Typical execution time: <1 second per collector

## Version Compatibility

### RHEL Version Support
- **RHEL 7**: Full support
- **RHEL 8**: Full support
- **RHEL 9**: Full support
- **RHEL 10**: Full support (tested with beta/preview releases)

### Package Manager Compatibility
- DNF (RHEL 8+): Full support
- YUM (RHEL 7): Full support with automatic fallback
- Security update checks: DNF only (gracefully skipped on YUM systems)

## Notes
- Binaries are statically compiled and have no external dependencies
- No specific Go version required on target systems
- Compatible with any RHEL-based distribution using RPM and DNF/YUM
- Collectors gracefully handle missing commands (report error metrics)

---

**Designed by Ifesinachi Osude**
Version: 1.0.0
