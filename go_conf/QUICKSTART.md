# Quick Start Guide - RHEL Stats Collectors

**Designed by Ifesinachi Osude**

## 1. Build Binaries

```bash
cd go_conf
./build_all.sh
```

This creates three binaries:
- `go_build/dnf_last_update/dnf_last_update`
- `go_build/dnf_update_check/dnf_update_check`
- `go_build/chrony_sources/chrony_sources`

## 2. Install Binaries

```bash
sudo cp go_build/dnf_last_update/dnf_last_update /usr/local/bin/
sudo cp go_build/dnf_update_check/dnf_update_check /usr/local/bin/
sudo cp go_build/chrony_sources/chrony_sources /usr/local/bin/
sudo chmod +x /usr/local/bin/{dnf_last_update,dnf_update_check,chrony_sources}
```

## 3. Test Binaries

```bash
/usr/local/bin/dnf_last_update
/usr/local/bin/dnf_update_check
/usr/local/bin/chrony_sources
```

Expected output: Prometheus format metrics

## 4. Configure Telegraf

```bash
sudo cp rhel_stats.conf /etc/telegraf/telegraf.d/
sudo systemctl restart telegraf
```

## 5. Verify Metrics

```bash
# Check telegraf is collecting
telegraf --test --config /etc/telegraf/telegraf.d/rhel_stats.conf

# Check prometheus metrics
curl http://localhost:9273/metrics | grep -E "(dnf_|chrony_)"
```

## 6. Grafana Dashboard

Use these PromQL queries:

### Current Kernel
```promql
dnf_update_kernel_current_version_info
```

### Pending Kernel (if update available)
```promql
dnf_update_kernel_pending_version_info{version!=""}
```

### Pending Updates Count
```promql
rhel_dnf_update_pending_count
```

### Days Since Last Update
```promql
(time() - rhel_dnf_last_update) / 86400
```

---

**Version**: v1.0.0  
**For detailed documentation, see README.md**
