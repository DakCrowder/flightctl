
services-container: rpm
	sudo podman build -t flightctl-services:latest -f test/scripts/services-images/Containerfile.services .

run-services-container:
	sudo podman run -d --privileged --replace \
	--name flightctl-services \
	-p 8080:443 \
	-p 3443:3443 \
	-p 8090:8090 \
	localhost/flightctl-services:latest

clean-services-container:
	sudo podman stop flightctl-services || true
	sudo podman rm flightctl-services || true

build-services-qcow:
	mkdir -p bin/output && \
	sudo podman run --rm -it --privileged --pull=newer \
		--security-opt label=type:unconfined_t \
		-v "${PWD}/bin/output":/output \
		-v /var/lib/containers/storage:/var/lib/containers/storage \
		quay.io/centos-bootc/bootc-image-builder:latest \
		--type qcow2 \
		--local \
		localhost/flightctl-services:latest

.PHONY: services-container run-services-container clean-services-container
