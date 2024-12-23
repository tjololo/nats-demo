ko-build:
	ko build github.com/tjololo/nats-demo/service --tarball nat-demo-tar --push=false
	podman load -i nat-demo-tar
	rm nat-demo-tar
