# Tailscale simple dns

This project aims to provide a very very simple binary which can be used on windows/mac/linux which will periodically 
call the `tailscale status`, parse the output and add each host to your local hosts file.

Tailscale is an awesome WireGaurd vpn. They have a "magic dns" feature however it doesn't seem to provide local caching
or much control about DNS over TLS etc and caused me a few problems in the past. For now im rolling my own simple wrapper around the hosts file.

```bash
➜  ~ tailscale-simple-dns --help
USAGE
  tailscale-simple-dns [flags]

FLAGS
  -cron @every 1m  controls how frequently the sync runs can be any vaild cron experssion
  -domains ...     required: domains to append to the tailscale hostname
  -dry-run=true    dry run will print the updated hosts file to os.Stdout rather than updating /etc/hosts
  -timeout 10s     set a timeout for the entire operation
```

You can also use environment variables to configure the binary `CRON="@every 1m" DRY_RUN=false DOMAINS=example.com` etc.. flags take precedence over environment variables.

# Installation
Its probably more convenient to run this as a daemon  

## Linux
Using systemd is probably the simplest way and is how im running it for now... enjoy this beautiful bash script which should install everything

```bash
export VERSION=0.0.1 # pick the latest version
export DOMAINS=example.com # set your DNS name
export DRY_RUN=false 
export CRON="@every 1m"
sudo curl -sSL "https://github.com/BradErz/tailscale-simple-dns/releases/download/v${VERSION}/tailscale-simple-dns_${VERSION}_$(uname -s)_$(uname -m).tar.gz" | tar -xzvf - 
rm -f README.md
sudo mv ./tailscale-simple-dns /usr/local/bin/tailscale-simple-dns
sudo chmod +x /usr/local/bin/tailscale-simple-dns
curl -s https://raw.githubusercontent.com/BradErz/tailscale-simple-dns/main/init/tailscale-simple-dns.service | envsubst > /tmp/tailscale-simple-dns.service
sudo mv /tmp/tailscale-simple-dns.service /etc/systemd/system/tailscale-simple-dns.service
sudo systemctl daemon-reload
sudo systemctl enable --now tailscale-simple-dns.service
```

you can check the status of the service doing the following:
```bash
systemctl status tailscale-simple-dns.service
journalctl -u tailscale-simple-dns.service -f --no-pager
```

## OTHERS (TODO)