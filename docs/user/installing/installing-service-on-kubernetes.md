# Installing the Flight Control Service on OpenShift/Kubernetes

## Installing on Kubernetes

You can install the Flight Control Service on any certified Kubernetes distribution that supports the Gateway API. If you have an OpenShift Kubernetes cluster available, refer to [Installing on OpenShift](#installing-on-openshift) for a more streamlined experience.

It is recommended to install `cert-manager` before installing Flight Control. When the Flight Control installer detects `cert-manager`, it will use it to issue and manage required CA and server TLS certificates. Otherwise, it falls back to creating certificates using Helm's built-in functions once, but does not manage them.

### (Optional) Installing a `kind` cluster

If you do not have a Kubernetes cluster available, you can use the [installation on Linux](installing-service-on-linux.md) or create a local test cluster using `kind` (Kubernetes in Docker). This section guides you through setting up a `kind` cluster with Gateway API support for use with Flight Control.

Prerequisites:

- You have Podman or Docker installed and running.
- You have `kind` version 0.31.0 or later installed. See [kind Installation](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) for details.
- You have `kubectl` installed.

**Note:** On RHEL 9 and similar systems using rootless Podman, you must configure systemd to enable `kind` to work properly. Run the following commands:

```console
sudo mkdir -p /etc/systemd/system/user@.service.d
cat << 'EOF' | sudo tee /etc/systemd/system/user@.service.d/delegate.conf
[Service]
Delegate=yes
EOF
sudo systemctl daemon-reload
```

After running these commands, log out and log back in for the changes to take effect.

Procedure:

1. Create a `kind` cluster configuration file. This configuration maps host ports to the `NodePort`s that will be used by the Envoy Gateway:

    ```console
    cat > kind-flightctl.yaml <<EOF
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
    - role: control-plane
      extraPortMappings:
      - containerPort: 30080
        hostPort: 8080
        protocol: TCP
      - containerPort: 30443
        hostPort: 8443
        protocol: TCP
    EOF
    ```

2. Create the `kind` cluster:

    ```console
    kind create cluster --name flightctl --config kind-flightctl.yaml
    ```

3. Verify the cluster is running:

    ```console
    kubectl cluster-info
    ```

4. Install the Gateway API CRDs:

    ```console
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml
    ```

5. Install the Envoy Gateway controller using Helm:

    ```console
    helm install eg oci://docker.io/envoyproxy/gateway-helm --version v1.2.5 \
      -n envoy-gateway-system --create-namespace
    ```

6. Wait for the Envoy Gateway deployment to be ready:

    ```console
    kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
    ```

7. Create an `EnvoyProxy` resource to configure `NodePort` service type with specific ports matching the `kind` cluster configuration:

    ```console
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyProxy
    metadata:
      name: nodeport-proxy
      namespace: envoy-gateway-system
    spec:
      provider:
        type: Kubernetes
        kubernetes:
          envoyService:
            type: NodePort
            patch:
              type: StrategicMerge
              value:
                spec:
                  ports:
                  - name: http-80
                    nodePort: 30080
                    port: 80
                    protocol: TCP
                  - name: https-443
                    nodePort: 30443
                    port: 443
                    protocol: TCP
    EOF
    ```

8. Create the Envoy Gateway `GatewayClass` with a reference to the `EnvoyProxy` resource:

    ```console
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.networking.k8s.io/v1
    kind: GatewayClass
    metadata:
      name: envoy-gateway
    spec:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
      parametersRef:
        group: gateway.envoyproxy.io
        kind: EnvoyProxy
        name: nodeport-proxy
        namespace: envoy-gateway-system
    EOF
    ```

Your `kind` cluster is now ready for Flight Control installation. Flight Control will automatically create the Gateway and routing resources when deployed.

When deploying Flight Control, use:

- `global.baseDomain=<your_ip>.nip.io` where `<your_ip>` is your host's IP address (e.g., `127.0.0.1.nip.io` for local access). The `nip.io` wildcard DNS service ensures proper resolution of subdomains like `ui.<your_ip>.nip.io` and `api.<your_ip>.nip.io`. Alternatively, use a domain you control with proper DNS configuration.
- `global.gateway.gatewayClassName=envoy-gateway` to reference the correct `GatewayClass`
- `global.gateway.ports.http=80` and `global.gateway.ports.tls=443` to match the Gateway listener ports

### Installing using the CLI

Prerequisites:

- You have access to a Kubernetes cluster (v1.35+) with cluster admin permissions.
- You have `kubectl` of a matching version installed and are logged in to the Kubernetes cluster.
- You have the [Gateway API CRDs installed](https://gateway-api.sigs.k8s.io/guides/#installing-gateway-api) on your cluster.
- You have a Gateway controller installed (e.g., Envoy Gateway or Istio).
- You have `helm` version 3.17+ installed.

Procedure:

1. Define the Flight Control version you want to install.

    ```console
    FC_VERSION=1.1.0
    ```

2. Define the namespace you want to install Flight Control into.

    ```console
    FC_NAMESPACE=flightctl
    ```

3. Define the base domain for your Flight Control installation. This should be a domain or subdomain you control. If deploying on `kind`, use the `nip.io` wildcard DNS service with your host IP for proper subdomain resolution.

    ```console
    FC_BASE_DOMAIN=127.0.0.1.nip.io
    ```

    Note: Replace `127.0.0.1` with your actual host IP if accessing from other machines. For production deployments, use `FC_BASE_DOMAIN=flightctl.example.com` with a domain you control.

4. (Optional) If you have a wildcard TLS certificate for your base domain, create a Kubernetes secret with it:

    ```console
    kubectl create secret tls flightctl-tls -n ${FC_NAMESPACE} \
      --cert=<path_to_tls_cert> \
      --key=<path_to_tls_key>
    ```

5. Deploy Flight Control by running the following command, whereby `gatewayClassName` needs to be set appropriately for your gateway, e.g. `envoy-gateway`:

    ```console
    helm upgrade --install flightctl oci://quay.io/flightctl/charts/flightctl:${FC_VERSION} \
      --namespace ${FC_NAMESPACE} --create-namespace \
      --set global.baseDomain=${FC_BASE_DOMAIN} \
      --set global.gateway.gatewayClassName=envoy-gateway \
      --set global.gateway.ports.http=80 \
      --set global.gateway.ports.tls=443
    ```

    If you created a TLS secret in step 4, add:

    ```console
      --set global.baseDomainTlsSecretName=flightctl-tls
    ```

6. Wait for the pods to be in `Running` or `Completed` state:

    ```console
    kubectl get pods -n ${FC_NAMESPACE}
    ```

7. Get the external IP or hostname of your Gateway:

    ```console
    kubectl get gateway flightctl-gateway -n ${FC_NAMESPACE}
    ```

    Configure your DNS to point `*.${FC_BASE_DOMAIN}` to this address.

You can then access the Flight Control UI by navigating your web browser to:

```console
https://ui.${FC_BASE_DOMAIN}
```

Or for HTTP if you did not configure TLS:

```console
http://ui.${FC_BASE_DOMAIN}
```

To use the `flightctl` CLI with your service:

1. Get the `ServiceAccount` token for the default admin user:

    ```console
    FC_TOKEN=$(kubectl create token flightctl-admin -n ${FC_NAMESPACE} --duration=24h)
    ```

2. Log in to your Flight Control service:

    ```console
    FC_API_ENDPOINT=api.${FC_BASE_DOMAIN}

    flightctl login ${FC_API_ENDPOINT} -t ${FC_TOKEN}
    ```

## Installing on OpenShift

When you install the Flight Control Service on the OpenShift Kubernetes distribution, the installer automatically leverages OpenShift's built-in features like an OAuth2 authentication server and multi-tenancy support.

It is recommended to install the `cert-manager` Operator from the OpenShift Software Catalog before installing Flight Control. When the Flight Control installer detects `cert-manager`, it will use it to issue and manage required CA and server TLS certificates. Otherwise, it falls back to creating certificates using `openssl` once, but does not manage them.

### Installing using the CLI

Prerequisites:

- You have access to an OpenShift Kubernetes cluster (4.19+) with cluster admin permissions.
- You have `oc` of a matching version installed and are logged in to the OpenShift cluster.
- You have `helm` version 3.17+ installed.

Procedure:

1. Define the Flight Control version you want to install.

    ```console
    FC_VERSION=1.1.0
    ```

2. Define the namespace you want to install Flight Control into.

    ```console
    FC_NAMESPACE=flightctl
    ```

3. Deploy that version by running:

    ```console
    helm upgrade --install flightctl oci://quay.io/flightctl/charts/flightctl:${FC_VERSION} \
      --namespace ${FC_NAMESPACE} --create-namespace
    ```

4. Wait for the pods to be in `Running` or `Completed` state:

    ```console
    oc get pods -n ${FC_NAMESPACE}
    ```

You can then access the Flight Control UI by navigating your web browser to the URL returned by this command:

```console
oc get route flightctl-ui -n ${FC_NAMESPACE} -o jsonpath='{.spec.host}'
```

Alternatively, you can use the following commands to log the `flightctl` CLI in to your service:

```console
FC_API_ENDPOINT=$(oc get route flightctl-api -n ${FC_NAMESPACE} -o jsonpath='{.spec.host}')

flightctl login ${FC_API_ENDPOINT} -t $(oc whoami -t)
```

### Installing from the OpenShift Software Catalog

Prerequisites:

- You have access to an OpenShift Kubernetes cluster (4.19+) with cluster admin permissions.

Procedure:

1. On the OpenShift console with the **Administrator** view selected, navigate to **Home > Projects**, click the blue "Create Project" button, in the field "Name" enter "flightctl", and click the blue "Create" button.

2. Navigate to **Ecosystem > Software Catalog**. In the search bar below "All Items", enter "flightcontrol", click the "Flight Control" tile, and click the blue "Create" button.

3. Verify the selected project is "flightctl". Optionally, select a Flight Control version to install. Then click the blue "Create" button.

4. Navigate to **Workloads > Pods** and wait for the pods to be in `Running` or `Completed` state.

5. Navigate to **Networking > Routes** and click on the `flightctl-ui` route.

6. Click on the URL shown under "Location" to access the Flight Control UI with your current user account and credentials.

If you need to access your service using the `flightctl` CLI, find and click the "Copy login command" button to see the command to use to log the `flightctl` CLI in to your service.

## Certificate management

Flight Control uses mutual TLS (mTLS) for all agent-to-service communication, requiring a full certificate chain. By default, certificates are auto-generated during installation. For production deployments, you can provide certificates from your own PKI infrastructure.

For the complete certificate chain of trust, validity periods, and file locations, see [Certificate Architecture](../references/certificate-architecture.md).

> [!NOTE]
> This section covers internal mTLS certificates used for service-to-service and agent-to-service communication. These are separate from the optional wildcard TLS secret (`baseDomainTlsSecretName`) used for gateway or ingress TLS termination.

### Automatic certificate generation

Flight Control supports four certificate generation modes, configured with `global.generateCertificates` in Helm values:

| Mode | Description |
|------|-------------|
| `auto` (default) | Uses `cert-manager` if installed, otherwise falls back to `builtin` |
| `cert-manager` | Requires cert-manager to be pre-installed. Handles automatic certificate renewal |
| `builtin` | Runs a one-time Job using openssl to generate certificates. Does not handle renewal |
| `none` | User must create all required Kubernetes secrets before deployment |

### Providing all certificates

Use this approach when you have an existing PKI and want to provide every certificate yourself.

Prerequisites:

- Certificates generated from your PKI meeting the requirements below.
- `kubectl` or `oc` access with permission to create secrets in the target namespace.

Certificate requirements:

- CA certificates must have `basicConstraints = CA:TRUE` and `keyUsage = keyCertSign, cRLSign, digitalSignature`
- Server certificates must have `extendedKeyUsage = serverAuth`
- Server certificates must include the DNS SANs listed in [Certificate Architecture](../references/certificate-architecture.md#required-dns-sans-per-service)

Procedure:

1. Create the root CA secret:

    ```console
    kubectl create secret tls flightctl-ca \
      --namespace <namespace> \
      --cert=<path_to_ca_crt> \
      --key=<path_to_ca_key>
    ```

2. Create the client-signer intermediate CA secret:

    ```console
    kubectl create secret tls flightctl-client-signer-ca \
      --namespace <namespace> \
      --cert=<path_to_client_signer_crt> \
      --key=<path_to_client_signer_key>
    ```

3. Create the API server TLS secret:

    ```console
    kubectl create secret tls flightctl-api-server-tls \
      --namespace <namespace> \
      --cert=<path_to_api_server_crt> \
      --key=<path_to_api_server_key>
    ```

4. Create secrets for additional services as needed:

    | Secret name | When required |
    |-------------|---------------|
    | `flightctl-telemetry-gateway-server-tls` | If telemetry is enabled |
    | `flightctl-alertmanager-proxy-server-tls` | If Alertmanager is enabled |
    | `flightctl-imagebuilder-api-server-tls` | If imagebuilder is enabled |
    | `flightctl-ui-server-tls` | If the UI is enabled |
    | `flightctl-cli-artifacts-server-tls` | If CLI artifacts serving is enabled |

    Create each secret using the same `kubectl create secret tls` pattern shown above.

5. Create the CA bundle secret. The bundle must contain `ca.crt` concatenated with `client-signer.crt`:

    ```console
    cat <path_to_ca_crt> <path_to_client_signer_crt> > ca-bundle.crt

    kubectl create secret generic flightctl-ca-bundle \
      --namespace <namespace> \
      --from-file=ca-bundle.crt=ca-bundle.crt
    ```

6. Deploy Flight Control with certificate generation disabled:

    ```console
    helm upgrade --install flightctl oci://quay.io/flightctl/charts/flightctl:${FC_VERSION} \
      --namespace <namespace> --create-namespace \
      --set global.baseDomain=<base_domain> \
      --set global.generateCertificates=none \
      --set global.gateway.gatewayClassName=<gateway_class>
    ```

7. Verify that pods start correctly:

    ```console
    kubectl get pods -n <namespace>
    ```

> [!NOTE]
> When using `generateCertificates: none`, certificates are not automatically rotated. You must track expiration dates and manually renew certificates before they expire.

### Using an existing certificate authority

Use this approach when you have an existing root CA and want Flight Control to derive all other certificates (intermediate CA, server certificates) from it.

The Helm chart checks for an existing CA before generating a self-signed one. In `cert-manager` mode, it looks for an existing `cert-manager.io/v1 Certificate` resource named `flightctl-ca`. In `builtin` mode, it looks for an existing `flightctl-ca` secret containing `tls.crt` and `tls.key`. In both cases, if the CA already exists, the installer uses it as the root of trust and derives all other certificates automatically, including the client-signer intermediate CA and all server certificates.

> [!NOTE]
> Using `cert-manager` is recommended for this scenario because it provides automatic certificate renewal. The `builtin` mode generates certificates once and does not manage renewal.

Prerequisites:

- Your root CA certificate and private key files in PEM format.
- `kubectl` or `oc` access with permission to create secrets in the target namespace.

Procedure:

1. Create the namespace if it does not already exist:

    ```console
    kubectl create namespace <namespace>
    ```

2. Create the `flightctl-ca` secret with your root CA certificate and key:

    ```console
    kubectl create secret tls flightctl-ca \
      --namespace <namespace> \
      --cert=<path_to_ca_crt> \
      --key=<path_to_ca_key>
    ```

3. Deploy Flight Control. The installer detects the existing `flightctl-ca` secret, skips self-signed CA generation, and derives all other certificates from your CA:

    ```console
    helm upgrade --install flightctl oci://quay.io/flightctl/charts/flightctl:${FC_VERSION} \
      --namespace <namespace> --create-namespace \
      --set global.baseDomain=<base_domain> \
      --set global.gateway.gatewayClassName=<gateway_class>
    ```

    To explicitly select a generation mode, add `--set global.generateCertificates=cert-manager` or `--set global.generateCertificates=builtin`.

4. Verify that pods start correctly:

    ```console
    kubectl get pods -n <namespace>
    ```

5. Verify that all certificates were derived from your CA:

    ```console
    kubectl get secrets -n <namespace> | grep flightctl
    ```

> [!IMPORTANT]
> Back up your CA private keys. If lost, all issued certificates become unverifiable and devices must be re-enrolled.

## Installing with Advanced Cluster Management

## Installing with Ansible Automation Platform
