# OpenShift AI Lightspeed Operator

OpenShift AI Lightspeed Operator is a Kubernetes Operator that deploys and configures a generative AI-based virtual assistant for Red Hat OpenShift AI (RHOAI) users.

## Overview

The assistant helps users with:

- Natural-language questions about OpenShift AI
- Troubleshooting AI/ML workloads
- Configuration guidance and best practices
- Understanding RHOAI features and capabilities

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    OpenShift Cluster                            │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  OpenShift AI Lightspeed Operator                      │     │
│  │  (This Repository)                                     │     │
│  │                                                        │     │
│  │  Manages:                                              │     │
│  │  ├── OpenShift Lightspeed Operator (OLS) installation  │     │
│  │  ├── OLSConfig (LLM + RAG configuration)               │     │
│  │  └── RAG content container (RHOAI documentation)       │     │
│  └────────────────────────────────────────────────────────┘     │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  OpenShift Lightspeed Operator (OLS)                   │     │
│  │  (Installed automatically via OLM)                     │     │
│  │                                                        │     │
│  │  Provides:                                             │     │
│  │  ├── Chat UI widget in OpenShift Console               │     │
│  │  ├── LLM integration layer                             │     │
│  │  └── RAG (Retrieval Augmented Generation) pipeline     │     │
│  └────────────────────────────────────────────────────────┘     │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  External LLM Provider                                 │     │
│  │  (OpenAI, Azure OpenAI, WatsonX, RHOAI vLLM, etc.)    │     │
│  └────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

## Key Components

| Component | Purpose |
|-----------|---------|
| `OpenShiftAILightspeed` CRD | User-facing custom resource for configuration |
| Controller | Reconciles desired state, manages OLS operator lifecycle |
| RAG Content | Pre-built vector database with RHOAI documentation |
| OLSConfig | Configuration passed to the underlying OLS operator |

## End-User Experience

1. **Administrator deploys the operator** to the cluster
2. **Administrator creates an `OpenShiftAILightspeed` resource** with LLM configuration
3. **Operator automatically**:
   - Installs the OpenShift Lightspeed operator via OLM
   - Configures it with the LLM endpoint and RAG content
   - Sets up the chat widget in the OpenShift Console
4. **End users** see a chat widget in the lower-right corner of the OpenShift web console and can ask questions about OpenShift AI

## Supported LLM Providers

| Provider | `llmEndpointType` value |
|----------|-------------------------|
| OpenAI / OpenAI-compatible | `openai` |
| Azure OpenAI | `azure_openai` |
| IBM WatsonX | `watsonx` |
| IBM BAM | `bam` |
| RHOAI vLLM | `rhoai_vllm` |
| RHEL AI vLLM | `rhelai_vllm` |

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

## CRD Reference

### OpenShiftAILightspeed Spec

| Field | Required | Description |
|-------|----------|-------------|
| `llmEndpoint` | Yes | URL pointing to the LLM provider |
| `llmEndpointType` | Yes | Provider type (see supported providers above) |
| `modelName` | Yes | Name of the model to use at the LLM endpoint |
| `llmCredentials` | Yes | Secret name containing API token (key: `apitoken`) |
| `ragImage` | No | Container image for RAG content (defaults to RHOAI docs) |
| `tlsCACertBundle` | No | ConfigMap name containing CA certificates |
| `maxTokensForResponse` | No | Maximum tokens for response generation (default: 2048) |
| `catalogSourceNamespace` | No | Namespace for OLS CatalogSource (default: `openshift-marketplace`) |
| `catalogSourceName` | No | Name of CatalogSource for OLS (default: `redhat-operators`) |
| `llmProjectID` | No | Project ID for providers like WatsonX |
| `llmDeploymentName` | No | Deployment name for Azure OpenAI |
| `llmAPIVersion` | No | API version for Azure OpenAI |
| `feedbackDisabled` | No | Disable feedback collection |
| `transcriptsDisabled` | No | Disable conversation transcripts collection |

### Status Conditions

| Condition | Description |
|-----------|-------------|
| `OpenShiftAILightspeedReady` | Instance is configured and operational |
| `OpenShiftLightspeedOperatorReady` | OLS operator is installed and operational |

## Repository Structure

```
openshift-ai-lightspeed-operator/
├── api/v1beta1/           # CRD type definitions
├── cmd/main.go            # Operator entry point
├── config/                # Kubernetes manifests (CRD, RBAC, deployment)
├── internal/controller/   # Reconciliation logic
│   ├── openshiftailightspeed_controller.go  # Main reconciler
│   ├── funcs.go           # OLSConfig management helpers
│   └── ols_install.go     # OLS operator installation via OLM
├── pkg/common/            # Shared utilities
└── test/                  # KUTTL and E2E tests
```

## Key Design Decisions

1. **Automatic OLS Operator Installation**: The operator installs OpenShift Lightspeed via OLM with manual InstallPlan approval to prevent unexpected upgrades

2. **Ownership Tracking**: Uses Kubernetes OwnerReferences and labels to ensure proper cleanup on deletion

3. **RAG Pre-packaging**: Ships with a pre-built vector database containing RHOAI documentation (`quay.io/opendatahub-io/openshift-ai-lightspeed-rag-content:rhoai-docs-2025.1`)

4. **Singleton OLSConfig**: Only one OLSConfig named "cluster" exists; the operator patches it with the user's configuration

## Technology Stack

- **Language**: Go 1.24
- **Framework**: Kubebuilder v4 + controller-runtime v0.22
- **Kubernetes**: Compatible with v1.34+
- **Testing**: Ginkgo/Gomega + KUTTL
- **Packaging**: OLM (Operator Lifecycle Manager) bundles

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
