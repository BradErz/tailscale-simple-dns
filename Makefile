TAG?=0.0.1

release:
	git tag -a v$(TAG) -m "First release"
	git push origin v$(TAG)

