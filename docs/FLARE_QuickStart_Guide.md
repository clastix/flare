# FLARE QuickStart Guide

## Overview

This guide sets up a FLARE development environment using KinD (Kubernetes in Docker) clusters with LAN Discovery enabled. The setup creates one consumer hub cluster and two GPU provider clusters (each with 3 total nodes: 1 control-plane + 2 workers) for testing FLARE's multi-provider federation.

## Prerequisites

- Docker v28.1.1+
- kubectl v1.33.0+
- Helm v3.17.3+
- KinD v0.27.0+
- liqoctl v1.0.0+

## Development Environment Setup

### Step 1: System Preparation

```bash
# Prepare system for multiple KinD clusters
sudo swapoff -a
sudo sysctl fs.inotify.max_user_instances=8192
sudo sysctl fs.inotify.max_user_watches=524288
```

### Step 2: Create Multi-Provider KinD Clusters

```bash
# Clone FLUIDOS repository
git clone https://github.com/fluidos-project/node.git
cd node/tools/scripts

# Run setup script
./setup.sh

# Select custom setup:
# 2. Use a custom KIND environment with n consumer and m providers
# Number of consumer clusters: 1
# Number of provider clusters: 2
# Local repositories: y (for FLARE development) | n (for FLARE simulation)
# Resource auto discovery: y
# LAN node discovery: y
```

### Step 3: Configure Consumer as Non-Provider Hub

```bash
# Remove resource labels from consumer nodes (hub should not provide resources)
export KUBECONFIG=fluidos-consumer-1-config
kubectl label node fluidos-consumer-1-worker node-role.fluidos.eu/resources-
kubectl label node fluidos-consumer-1-worker2 node-role.fluidos.eu/resources-

# Delete any auto-created flavors on consumer
kubectl delete flavors --all -n fluidos

# Verify consumer has no flavors
kubectl get flavors -n fluidos
```

### Step 4: Install Fake GPU Operator on Providers

```bash
# Install on Provider 1 with H100 configuration
export KUBECONFIG=fluidos-provider-1-config
helm install fake-gpu-operator \
  oci://ghcr.io/run-ai/fake-gpu-operator/fake-gpu-operator \
  --namespace gpu-operator \
  --create-namespace \
  --version 0.0.63 \
  --set devicePlugin.image.tag="0.0.63" \
  --set statusUpdater.image.tag="0.0.63" \
  --set topologyServer.image.tag="0.0.63" \
  --set statusExporter.image.tag="0.0.63" \
  --set kwokGpuDevicePlugin.image.tag="0.0.63" \
  --set migFaker.image.tag="0.0.63" \
  --set topology.nodePools.default.gpuProduct="NVIDIA-H100" \
  --set topology.nodePools.default.gpuCount=2 \
  --set topology.nodePools.default.gpuMemory="81920" 

# Annotate nodes to simulate GPU resources
kubectl label node fluidos-provider-1-worker run.ai/simulated-gpu-node-pool=default
kubectl label node fluidos-provider-1-worker2 run.ai/simulated-gpu-node-pool=default

# Install on Provider 2 with A6000 configuration  
export KUBECONFIG=fluidos-provider-2-config
helm install fake-gpu-operator \
  oci://ghcr.io/run-ai/fake-gpu-operator/fake-gpu-operator \
  --namespace gpu-operator \
  --create-namespace \
  --version 0.0.63 \
  --set devicePlugin.image.tag="0.0.63" \
  --set statusUpdater.image.tag="0.0.63" \
  --set topologyServer.image.tag="0.0.63" \
  --set statusExporter.image.tag="0.0.63" \
  --set kwokGpuDevicePlugin.image.tag="0.0.63" \
  --set migFaker.image.tag="0.0.63" \
  --set topology.nodePools.default.gpuProduct="NVIDIA-RTX-A6000" \
  --set topology.nodePools.default.gpuCount=2 \
  --set topology.nodePools.default.gpuMemory="49140" 

# Annotate nodes to simulate GPU resources
kubectl label node fluidos-provider-2-worker run.ai/simulated-gpu-node-pool=default
kubectl label node fluidos-provider-2-worker2 run.ai/simulated-gpu-node-pool=default
```

### Step 5: Annotate Provider Nodes for FLARE

```bash
# Comprehensive annotations for Provider 1 - Premium H100 Cloud
export KUBECONFIG=fluidos-provider-1-config

# Worker 1: H100 with full specifications (dedicated, non-interruptible)
kubectl annotate node fluidos-provider-1-worker \
  gpu.fluidos.eu/vendor="nvidia" \
  gpu.fluidos.eu/model="nvidia-h100" \
  gpu.fluidos.eu/count="4" \
  gpu.fluidos.eu/memory-per-gpu="80Gi" \
  gpu.fluidos.eu/tier="premium" \
  gpu.fluidos.eu/architecture="hopper" \
  gpu.fluidos.eu/cores="16896" \
  gpu.fluidos.eu/compute-capability="9.0" \
  gpu.fluidos.eu/clock-speed="1.98G" \
  gpu.fluidos.eu/fp32-tflops="83.0" \
  gpu.fluidos.eu/interconnect="nvlink" \
  gpu.fluidos.eu/interconnect-bandwidth-gbps="900" \
  gpu.fluidos.eu/topology="nvswitch" \
  gpu.fluidos.eu/multi-gpu-efficiency="0.95" \
  gpu.fluidos.eu/sharing-capable="true" \
  gpu.fluidos.eu/sharing-strategy="mig" \
  gpu.fluidos.eu/dedicated="true" \
  gpu.fluidos.eu/interruptible="false" \
  location.fluidos.eu/region="eu-west-1" \
  location.fluidos.eu/zone="zone-a" \
  cost.fluidos.eu/hourly-rate="4.5" \
  cost.fluidos.eu/currency="EUR" \
  workload.fluidos.eu/training-score="0.98" \
  workload.fluidos.eu/inference-score="0.95" \
  workload.fluidos.eu/hpc-score="0.99" \
  workload.fluidos.eu/graphics-score="0.30" \
  network.fluidos.eu/bandwidth-gbps="100" \
  network.fluidos.eu/latency-ms="1" \
  network.fluidos.eu/tier="premium" \
  provider.fluidos.eu/name="cloud-provider-1" \
  provider.fluidos.eu/preemptible="false"

# Worker 2: H100 with spot/preemptible configuration (lower cost)
kubectl annotate node fluidos-provider-1-worker2 \
  gpu.fluidos.eu/vendor="nvidia" \
  gpu.fluidos.eu/model="nvidia-h100" \
  gpu.fluidos.eu/count="4" \
  gpu.fluidos.eu/memory-per-gpu="80Gi" \
  gpu.fluidos.eu/tier="premium" \
  gpu.fluidos.eu/architecture="hopper" \
  gpu.fluidos.eu/cores="16896" \
  gpu.fluidos.eu/compute-capability="9.0" \
  gpu.fluidos.eu/clock-speed="1.98G" \
  gpu.fluidos.eu/fp32-tflops="83.0" \
  gpu.fluidos.eu/interconnect="nvlink" \
  gpu.fluidos.eu/interconnect-bandwidth-gbps="900" \
  gpu.fluidos.eu/topology="nvswitch" \
  gpu.fluidos.eu/multi-gpu-efficiency="0.95" \
  gpu.fluidos.eu/sharing-capable="true" \
  gpu.fluidos.eu/sharing-strategy="mig" \
  gpu.fluidos.eu/dedicated="false" \
  gpu.fluidos.eu/interruptible="true" \
  location.fluidos.eu/region="eu-west-1" \
  location.fluidos.eu/zone="zone-b" \
  cost.fluidos.eu/hourly-rate="1.8" \
  cost.fluidos.eu/currency="EUR" \
  workload.fluidos.eu/training-score="0.98" \
  workload.fluidos.eu/inference-score="0.95" \
  workload.fluidos.eu/hpc-score="0.99" \
  workload.fluidos.eu/graphics-score="0.30" \
  network.fluidos.eu/bandwidth-gbps="100" \
  network.fluidos.eu/latency-ms="1" \
  network.fluidos.eu/tier="premium" \
  provider.fluidos.eu/name="cloud-provider-1" \
  provider.fluidos.eu/preemptible="true"

# Comprehensive annotations for Provider 2 - Standard A6000 Cloud  
export KUBECONFIG=fluidos-provider-2-config

# Worker 1: A6000 for professional workloads (dedicated)
kubectl annotate node fluidos-provider-2-worker \
  gpu.fluidos.eu/vendor="nvidia" \
  gpu.fluidos.eu/model="nvidia-a6000" \
  gpu.fluidos.eu/count="8" \
  gpu.fluidos.eu/memory-per-gpu="48Gi" \
  gpu.fluidos.eu/tier="standard" \
  gpu.fluidos.eu/architecture="ampere" \
  gpu.fluidos.eu/cores="10752" \
  gpu.fluidos.eu/compute-capability="8.6" \
  gpu.fluidos.eu/clock-speed="1.80G" \
  gpu.fluidos.eu/fp32-tflops="38.7" \
  gpu.fluidos.eu/interconnect="nvlink" \
  gpu.fluidos.eu/interconnect-bandwidth-gbps="600" \
  gpu.fluidos.eu/topology="ring" \
  gpu.fluidos.eu/multi-gpu-efficiency="0.85" \
  gpu.fluidos.eu/sharing-capable="false" \
  gpu.fluidos.eu/sharing-strategy="none" \
  gpu.fluidos.eu/dedicated="true" \
  gpu.fluidos.eu/interruptible="false" \
  location.fluidos.eu/region="us-east-1" \
  location.fluidos.eu/zone="zone-a" \
  cost.fluidos.eu/hourly-rate="2.1" \
  cost.fluidos.eu/currency="EUR" \
  workload.fluidos.eu/training-score="0.85" \
  workload.fluidos.eu/inference-score="0.90" \
  workload.fluidos.eu/hpc-score="0.80" \
  workload.fluidos.eu/graphics-score="0.95" \
  network.fluidos.eu/bandwidth-gbps="25" \
  network.fluidos.eu/latency-ms="5" \
  network.fluidos.eu/tier="standard" \
  provider.fluidos.eu/name="cloud-provider-2" \
  provider.fluidos.eu/preemptible="false"

# Worker 2: A6000 with time-slicing for inference (interruptible)
kubectl annotate node fluidos-provider-2-worker2 \
  gpu.fluidos.eu/vendor="nvidia" \
  gpu.fluidos.eu/model="nvidia-a6000" \
  gpu.fluidos.eu/count="8" \
  gpu.fluidos.eu/memory-per-gpu="48Gi" \
  gpu.fluidos.eu/tier="standard" \
  gpu.fluidos.eu/architecture="ampere" \
  gpu.fluidos.eu/cores="10752" \
  gpu.fluidos.eu/compute-capability="8.6" \
  gpu.fluidos.eu/clock-speed="1.80G" \
  gpu.fluidos.eu/fp32-tflops="38.7" \
  gpu.fluidos.eu/interconnect="pcie" \
  gpu.fluidos.eu/interconnect-bandwidth-gbps="64" \
  gpu.fluidos.eu/topology="mesh" \
  gpu.fluidos.eu/multi-gpu-efficiency="0.70" \
  gpu.fluidos.eu/sharing-capable="true" \
  gpu.fluidos.eu/sharing-strategy="time-slicing" \
  gpu.fluidos.eu/dedicated="false" \
  gpu.fluidos.eu/interruptible="true" \
  location.fluidos.eu/region="us-east-1" \
  location.fluidos.eu/zone="zone-b" \
  cost.fluidos.eu/hourly-rate="0.8" \
  cost.fluidos.eu/currency="EUR" \
  workload.fluidos.eu/training-score="0.75" \
  workload.fluidos.eu/inference-score="0.95" \
  workload.fluidos.eu/hpc-score="0.70" \
  workload.fluidos.eu/graphics-score="0.90" \
  network.fluidos.eu/bandwidth-gbps="10" \
  network.fluidos.eu/latency-ms="10" \
  network.fluidos.eu/tier="basic" \
  provider.fluidos.eu/name="cloud-provider-2" \
  provider.fluidos.eu/preemptible="true"

# Add GPU selector labels for all GPU nodes (for easy querying)
export KUBECONFIG=fluidos-provider-1-config
kubectl label node fluidos-provider-1-worker node.fluidos.eu/gpu="true"
kubectl label node fluidos-provider-1-worker2 node.fluidos.eu/gpu="true"

export KUBECONFIG=fluidos-provider-2-config
kubectl label node fluidos-provider-2-worker node.fluidos.eu/gpu="true"
kubectl label node fluidos-provider-2-worker2 node.fluidos.eu/gpu="true"
```

### Step 6: Verify Setup from Liqo and FLUIDOS Perspective

```bash
# Check Liqo status on all clusters
export KUBECONFIG=fluidos-consumer-1-config
liqoctl info

export KUBECONFIG=fluidos-provider-1-config
liqoctl info

export KUBECONFIG=fluidos-provider-2-config
liqoctl info

# Check FLUIDOS discovery from consumer perspective
export KUBECONFIG=fluidos-consumer-1-config
kubectl get knownclusters -n fluidos

# Check GPU resources on providers
export KUBECONFIG=fluidos-provider-1-config
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable."nvidia\.com/gpu"
kubectl get flavors -n fluidos

export KUBECONFIG=fluidos-provider-2-config
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable."nvidia\.com/gpu"
kubectl get flavors -n fluidos

# Verify consumer has no flavors (hub only)
export KUBECONFIG=fluidos-consumer-1-config
kubectl get flavors -n fluidos
```

### Step 7: Install FLARE Components

```bash
# Switch to consumer cluster
export KUBECONFIG=fluidos-consumer-1-config
# Install Capsule for Multi Tenancy
helm upgrade --install capsule oci://ghcr.io/projectcapsule/charts/capsule --version 0.10.5  \
  --namespace=capsule-system --create-namespace \
  --set "manager.options.capsuleUserGroups[0]=system:serviceaccounts:tenants" \
  --set "manager.options.forceTenantPrefix=true"
# The Namespace where ServiceAccount will be available
kubectl create namespace tenants --dry-run=client -o yaml | kubectl apply -f -
kubectl -n tenants create serviceaccount solar --dry-run=client -o yaml | kubectl apply -f -
# Create a Tenant: enrich with your Capsule policies
cat <<EOF | kubectl apply -f -
apiVersion: capsule.clastix.io/v1beta2
kind: Tenant
metadata:
  name: solar
spec:
  owners:
  - kind: ServiceAccount
    name: system:serviceaccount:tenants:solar
EOF
# Following steps are expected to run in kind:
# YMMV regarding container registry.
#
# Build and load FLARE container images
make load
# Install FLARE API Server and controllers
make install
# Generate a ServiceAccount token as `solar` Tenant to interact with the API
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: solar
  namespace: tenants
  annotations:
    kubernetes.io/service-account.name: solar
type: kubernetes.io/service-account-token
EOF
# Extract the token from such a Secret
kubectl -n tenants get secret solar -ojsonpath='{.data.token}' | base64 -d
# Port-forward to access the API locally
kubectl -n flare-system port-forward service/flare-server 8080:8080
```

### Step 8: (Optional) Install FLUIDOS Node Dashboard

```bash
# Install FLUIDOS dashboard on consumer cluster for visual monitoring
export KUBECONFIG=fluidos-consumer-1-config

# Clone the dashboard repository
git clone https://github.com/fluidos-project/fluidos-node-dashboard.git
cd fluidos-node-dashboard

# Install the dashboard
./install.sh kind fluidos-consumer-1

# Get the dashboard service details
kubectl get svc | grep dashboard

# Port-forward to access the dashboard locally
kubectl port-forward service/frontend-service 8080:80 &

# Access dashboard at http://localhost:8080
```

## Configuration Summary

| Cluster | Role | Resources | GPU Type | GPUs/Node |
|---------|------|-----------|----------|-----------|
| fluidos-consumer-1 | FLARE Hub | None (labels removed) | - | - |
| fluidos-provider-1 | GPU Provider | 2 worker nodes | H100 | 4 |
| fluidos-provider-2 | GPU Provider | 2 worker nodes | A6000 | 8 |

**Key Configuration Steps Applied:**
1. Consumer nodes de-labeled (no `node-role.fluidos.eu/resources`)
2. Consumer flavors deleted (hub-only mode)
3. Fake GPU operator installed on providers only
4. Provider nodes annotated with `gpu.fluidos.eu/*` labels per FLARE spec
5. Different GPU models configured per provider (H100 vs A6000)

## Next Steps

- For production deployment, see [FLARE Admin Guide](FLARE_Admin_Guide.md)
- For real GPU hardware setup, configure actual GPU nodes
- For testing FLARE intents, deploy FLARE API Server components

## Cleanup

```bash
# Clean up all KinD clusters and resources
cd node/tools/scripts
./clean-dev-env.sh
```

This multi-provider quickstart environment provides comprehensive testing of FLARE's federation capabilities across multiple simulated GPU providers.