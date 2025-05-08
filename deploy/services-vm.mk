
services-container: rpm
	podman build -t flightctl-services-image -f test/scripts/services-images/Containerfile.services .

run-services-container:
	podman run -d --network host --name flightctl-services localhost/flightctl-services-image

clean-services-container:
	podman stop flightctl-services-container || true
	podman rm flightctl-services-container || true

.PHONY: services-container run-services-container clean-services-container
