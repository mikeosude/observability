#! /bin/bash
# This script is used to run telegraf with the specified configuration file


cat <<EOF | sudo tee /etc/yum.repos.d/influxdata.repo
[influxdata]
name = InfluxData Repository - Stable
baseurl = https://repos.influxdata.com/stable/\$basearch/main
enabled = 1
gpgcheck = 1
gpgkey = https://repos.influxdata.com/influxdata-archive.key
EOF

yum install -y telegraf

cat << 'EOF' > /etc/sudoers.d/telegraf
Defaults:telegraf !requiretty

telegraf ALL=(root) NOPASSWD: \
    /bin/cat /proc/slabinfo, \
    /usr/local/bin/chrony_sources, \
    /usr/local/bin/dnf_last_update, \
    /usr/local/bin/dnf_update_check
EOF

chmod 0440 /etc/sudoers.d/telegraf

sudo setfacl -R -m u:telegraf:r /var/log
sudo setfacl -R -d -m u:telegraf:r /var/log
setfacl -R -m u:telegraf:rx /var/log
setfacl -R -m d:u:telegraf:rx /var/log

sleep 5

echo "press CTRL+C to stop the script would start telegraf after 30 Seconds"

sleep 30

systemctl enable --now telegraf