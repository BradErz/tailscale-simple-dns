TAG?=0.0.1

release:
	git tag -a v$(TAG) -m "First release"
	git push origin v$(TAG)

update-local:
	go build -o tailscale-simple-dns ./cmd/tailscale-simple-dns/
	sudo mv ./tailscale-simple-dns /usr/local/bin/tailscale-simple-dns
	sudo chmod +x /usr/local/bin/tailscale-simple-dns
	sudo systemctl restart tailscale-simple-dns

logs:
	journalctl -u tailscale-simple-dns.service --no-pager -f