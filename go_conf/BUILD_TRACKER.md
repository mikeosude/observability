# Build Tracker - RHEL Stats Collectors

**Designed by Ifesinachi Osude**

This document tracks all builds, changes, and versions for the RHEL stats collectors.

---

## Version History

### v1.0.0 - 2026-01-14

**Initial Release**

#### Components Built
1. **dnf_last_update** - Tracks last package installation time
2. **dnf_update_check** - Monitors system updates and kernel versions
3. **chrony_sources** - Monitors NTP time synchronization

#### Features
- ✅ Static compilation (CGO_ENABLED=0)
- ✅ RHEL 7-10 compatibility
- ✅ Prometheus metric format
- ✅ Current AND pending kernel version tracking
- ✅ Telegraf exec plugin integration
- ✅ Version tracking via VERSION file
- ✅ Signature embedded in output

#### Build Configuration
```
OS: linux
Architecture: amd64
CGO: Disabled
Linker Flags: -s -w (strip symbols, reduce size)
Version Injection: -X main.version=$(VERSION)
```

#### Files Created
```
go_conf/
├── VERSION                          # Version tracking
├── README.md                        # Build and usage documentation
├── requirements.txt                 # System requirements
├── rhel_stats.conf                 # Telegraf configuration
├── build_all.sh                    # Master build script
├── BUILD_TRACKER.md                # This file
└── go_build/
    ├── dnf_last_update/
    │   ├── main.go
    │   └── build.sh
    ├── dnf_update_check/
    │   ├── main.go
    │   └── build.sh
    └── chrony_sources/
        ├── main.go
        └── build.sh
```

#### Metrics Exposed

**dnf_last_update (3 metrics)**
- dnf_last_update
- dnf_last_update_info{time}
- dnf_last_update_error

**dnf_update_check (11 metrics)**
- dnf_update_available
- dnf_update_pending_count
- dnf_update_kernel_available
- dnf_update_kernel_current_version_info{version}  ← NEW
- dnf_update_kernel_pending_version_info{version}  ← NEW
- dnf_update_security_available
- dnf_update_security_count
- dnf_update_reboot_required
- dnf_update_check_error
- dnf_update_pending_pkg{name,arch,version,repo}
- dnf_update_pending_info{packages}

**chrony_sources (8 metrics per source)**
- chrony_sources_up
- chrony_source_selected{source,mode,stratum,poll}
- chrony_source_in_use{source,mode,stratum,poll}
- chrony_source_reachable{source,mode,stratum,poll}
- chrony_source_reach{source,mode,stratum,poll}
- chrony_source_last_rx_seconds{source,mode,stratum,poll}
- chrony_source_offset_seconds{source,mode,stratum,poll}
- chrony_source_jitter_seconds{source,mode,stratum,poll}

#### Key Design Decisions
1. **Static Compilation**: Ensures compatibility across RHEL 7-10 without Go installation
2. **No External Dependencies**: Pure Go implementation, no cgo
3. **Version Injection**: Build-time version from VERSION file
4. **Dual Kernel Metrics**: Both current and pending kernel versions for Grafana panels
5. **Prometheus Format**: Native support for Telegraf exec plugin
6. **Error Handling**: Graceful degradation with error metrics

#### Testing Targets
- RHEL 7.x (CentOS 7)
- RHEL 8.x (Rocky Linux 8, AlmaLinux 8)
- RHEL 9.x (Rocky Linux 9, AlmaLinux 9)
- RHEL 10.x (preview/beta)

---

## Build Instructions

### Quick Build
```bash
cd go_conf
./build_all.sh
```

### Individual Builds
```bash
cd go_conf/go_build/<component>
./build.sh
```

### Version Update
```bash
# Update version
echo "v1.0.1" > go_conf/VERSION

# Rebuild all
cd go_conf
./build_all.sh
```

---

## Change Log Template

### vX.Y.Z - YYYY-MM-DD

**Type**: [Feature/Bugfix/Security/Performance]

#### Changes
- [ ] Description of change
- [ ] Related issue/ticket

#### Metrics Added/Modified
- metric_name: Description

#### Files Modified
- path/to/file: Description

#### Build Configuration Changes
- Configuration change description

#### Breaking Changes
- Description of breaking changes (if any)

#### Testing
- Test scenarios covered
- RHEL versions tested

#### Migration Notes
- Steps required for upgrade (if any)

---

## Maintenance Notes

### Version Bumping
1. Update `VERSION` file in go_conf/ directory
2. Run `build_all.sh` to rebuild with new version
3. Update BUILD_TRACKER.md with changes
4. Test on target RHEL versions
5. Update README.md if needed

### Adding New Metrics
1. Modify main.go in respective component
2. Update README.md metrics section
3. Update BUILD_TRACKER.md
4. Rebuild and test
5. Update requirements.txt if new system commands needed

### Code Signing
Currently not implemented. Future consideration:
- GPG signing of binaries
- Checksum generation
- Release artifact management

---

## Build Verification

### Verification Steps
1. Build completes without errors
2. Binary size reasonable (<5MB per binary)
3. Binaries are statically linked (ldd shows "not a dynamic executable")
4. Version string embedded correctly
5. Signature appears in output
6. Metrics format valid Prometheus

### Verification Commands
```bash
# Check static linking
ldd go_build/dnf_last_update/dnf_last_update

# Check version
strings go_build/dnf_last_update/dnf_last_update | grep -E "v[0-9]+\.[0-9]+\.[0-9]+"

# Check signature
strings go_build/dnf_last_update/dnf_last_update | grep "Designed by"

# Test execution
go_build/dnf_last_update/dnf_last_update

# Validate Prometheus format
go_build/dnf_last_update/dnf_last_update | promtool check metrics
```

---

## Distribution

### Binary Distribution
- Binaries should be distributed via internal artifact repository
- Include VERSION file with distribution
- Include rhel_stats.conf telegraf configuration
- Include README.md for deployment instructions

### Deployment
1. Copy binaries to `/usr/local/bin/`
2. Set executable permissions
3. Copy telegraf config to `/etc/telegraf/telegraf.d/`
4. Restart telegraf service
5. Verify metrics in Prometheus

---

## Support Matrix

| RHEL Version | DNF Last Update | DNF Update Check | Chrony Sources | Status |
|--------------|----------------|------------------|----------------|---------|
| RHEL 7       | ✅             | ✅              | ✅            | Tested  |
| RHEL 8       | ✅             | ✅              | ✅            | Tested  |
| RHEL 9       | ✅             | ✅              | ✅            | Tested  |
| RHEL 10      | ✅             | ✅              | ✅            | Beta    |

---

## Known Issues

None at initial release.

---

## Future Enhancements

### Planned
- [ ] Add systemd service status metrics
- [ ] Add disk usage metrics
- [ ] Add process monitoring
- [ ] Add SELinux status metrics

### Under Consideration
- [ ] Configuration file support
- [ ] Custom metric intervals
- [ ] Metric filtering options
- [ ] JSON output format option

---

**Last Updated**: 2026-01-14  
**Maintainer**: Ifesinachi Osude  
**Current Version**: v1.0.0
