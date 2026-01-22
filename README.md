# OpenShift AI Lightspeed Operator

OpenShift AI Lightspeed Operator is a generative AI-based virtual assistant for
Red Hat OpenShift AI (RHOAI) users.

## Images

| Type | Quay.io Repository |
|------|--------------------|
| Operator | [quay.io/opendatahub-io/openshift-ai-lightspeed-operator](https://quay.io/repository/opendatahub-io/openshift-ai-lightspeed-operator?tab=tags) |
| Bundle | [quay.io/opendatahub-io/openshift-ai-lightspeed-operator-bundle](https://quay.io/repository/opendatahub-io/openshift-ai-lightspeed-operator-bundle?tab=tags) |
| Catalog | [quay.io/opendatahub-io/openshift-ai-lightspeed-operator-catalog](https://quay.io/repository/opendatahub-io/openshift-ai-lightspeed-operator-catalog?tab=tags) |

## Quickstart

### Prerequisites

You need access to an OpenShift cluster. For local development, you can use
[CodeReady Containers (CRC)](https://developers.redhat.com/products/openshift-local/overview).

1. Download and install CRC from the Red Hat Developer portal
2. Get your pull secret from `https://cloud.redhat.com/openshift/create/local`
3. Start CRC:

```bash
crc setup
crc start --pull-secret-file ~/pull-secret.txt --cpus 12 --memory 25600 --disk-size 100
eval $(crc oc-env)
oc login -u kubeadmin https://api.crc.testing:6443
```

### Deploy OpenShift AI Lightspeed Operator

Get the operator repository:

```bash
git clone https://github.com/opendatahub-io/openshift-ai-lightspeed-operator.git
cd openshift-ai-lightspeed-operator
```

First, deploy OpenShift AI Lightspeed Operator:

```bash
make openshift-ai-lightspeed-deploy
```

Next, verify that the OpenShift AI Lightspeed Operator pod is running:

```bash
$ oc get -n openshift-ai-lightspeed-operator-system pods
NAME                                                                  READY   STATUS    RESTARTS   AGE
openshift-ai-lightspeed-operator-controller-manager-76df7fbfb5wggr   1/1     Running   0          72s
```

### Set up the LLM endpoint along with its credentials

To access the LLM we need:
- An API Key (eg: in `LLM_KEY`)
- An URL for the server (eg: in `LLM_ENDPOINT`)
- A model (eg: in `LLM_MODEL`)
- Optionally a certificate to access the LLM endpoint (name stored in
  `CERT_SECRET_NAME`)

The API key will be stored in a `Secret`, the certificate in a `ConfigMap` and
the other 2 together with the references to the first 2 will be passed in the
`OpenShiftAILightspeed` resource that triggers the deployment.

Define the URL and model env vars, for example para Gemini:

```bash
LLM_ENDPOINT=https://generativelanguage.googleapis.com/v1beta/openai
LLM_MODEL=gemini-2.5-pro
LLM_KEY=<API TOKEN>
```

Create the LLM API key secret:

```bash
oc apply -f - <<EOF
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: openshift-ai-lightspeed-apitoken
  namespace: openshift-lightspeed
stringData:
  apitoken: $LLM_KEY
EOF
```

Not required for Gemini, but here is an example of an optional certificate:

```bash
CERT_SECRET_NAME=openshift-ai-lightspeed-certs
CERT_FILE=/path/to/cert.crt
```
```bash
oc apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
type: Opaque
metadata:
  name: $CERT_SECRET_NAME
  namespace: openshift-lightspeed
data:
  cert: |
$(sed 's/^/    /' "$CERT_FILE")
EOF
```

### Deploy

Create the `redhat-ods-applications` namespace if we haven't deployed OpenShift AI yet:

```bash
oc create namespace redhat-ods-applications
```

Deploy OpenShift AI Lightspeed, a configuration would look like this (for actual
examples look in following sections):

```bash
oc apply -f - <<EOF
apiVersion: lightspeed.openshift-ai.io/v1beta1
kind: OpenShiftAILightspeed
metadata:
  name: openshift-ai-lightspeed
  namespace: redhat-ods-applications
spec:
$(if [ -n "$RHOAI_LS_IMAGE" ]; then
  echo "  ragImage: $RHOAI_LS_IMAGE"
fi)
  llmEndpoint: $LLM_ENDPOINT
  llmEndpointType: openai
  llmCredentials: openshift-ai-lightspeed-apitoken
  modelName: $LLM_MODEL
$(if [ -n "$CERT_SECRET_NAME" ]; then
  echo "  tlsCACertBundle: $CERT_SECRET_NAME"
fi)
EOF
```

### Check deployment

Confirm the conditions are met

```bash
oc describe -n redhat-ods-applications openshiftailightspeed
oc describe -n openshift-lightspeed olsconfig
```

### Use
Now you can go to the [OpenShift web console](https://console-openshift-console.apps-crc.testing) using the `kubeadmin` username and `12345678` password and use the OpenShift Lightspeed console widget that should appear at the lower right corner.
You may need to click on `refresh` console link that appears on a message.

If you are running CRC on a different machine you can use `sshuttle` to connect to the remote system:
- Edit your local system's `/etc/hosts` (where you use the browser) and add this line verbatim (don't change the IP): `192.168.130.11 api.crc.testing canary-openshift-ingress-canary.apps-crc.testing console-openshift-console.apps-crc.testing default-route-openshift-image-registry.apps-crc.testing downloads-openshift-console.apps-crc.testing oauth-openshift.apps-crc.testing`
- In your local system run `sshuttle -r $remote_username@$remote_server 192.168.130.0/24`.
- Now the console should be accessible in your browser.

## Development

If you are making changes to the operator you can run the operator locally
(outside the cluster) using the Operator SDK make targets:

```bash
make install run
```

This will:

1. Install the CRDs into your cluster.
2. Run the operator locally, connected to your cluster.

Use this for quick development and testing.

*Attention*: In this mode RBACs are ignored, so when changing those please run
the operator in the OpenShift cluster with an image.

### Running Pre-Commit Hooks

To ensure code quality and consistency, run pre-commit hooks locally before
submitting a pull request.

Install hooks:

```bash
pre-commit install
```

Run all hooks manually:

```bash
pre-commit run --all-files
```

### Running KUTTL Tests

KUTTL (KUbernetes Test TooL) tests validate the operator's behavior in a real
OpenShift environment.

Before running the tests ensure that:
- `oc` CLI tool is available in your PATH and you can access an OpenShift cluster
(e.g., deployed with `crc`) with it
- The `openshift-lightspeed` namespace is empty or non-existing to prevent collisions

Once you are ready you can run the KUTTL tests using:

```bash
make kuttl-test-run
```

**Important Notes:**
- The tests use the `openshift-lightspeed` namespace to test in the exact namespace
where the OLS operator is expected to operate.
- The correct behavior of the OLS operator is not guaranteed outside of the
`openshift-lightspeed` namespace.
- Ensure the namespace is clean before running tests to avoid resource conflicts
or test failures.
