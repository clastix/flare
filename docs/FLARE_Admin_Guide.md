# FLARE Administration Guide

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Hub Cluster Setup (FLARE Consumer)](#hub-cluster-setup-flare-consumer)
4. [GPU Provider Setup](#gpu-provider-setup)
5. [Broker Requirements](#broker-requirements)
6. [Verification](#verification)
7. [Monitoring](#monitoring)
8. [Troubleshooting](#troubleshooting)

## Overview

This guide covers FLARE deployment across multiple independent GPU providers using FLUIDOS broker-based federation for WAN scenarios, 

- **Hub Cluster**: FLUIDOS consumer cluster hosting FLARE API Gateway
- **Provider Clusters**: Independent GPU providers 
- **WAN Discovery**: FLUIDOS Broker CRs for cross-network communication


## Prerequisites

**Remote Broker**: External broker service with certificates provided by broker administrator:

- Broker server address 
- CA root certificate (.pem file)
- Client certificates and private keys for each cluster (.pem files)

**Clusters**: Kubernetes 1.28+

**Tools**: kubectl, helm v3.8+, liqoctl v1.0.0+

**Node Labeling**: At least one node must be labeled `node-role.fluidos.eu/worker: "true"`

## Hub Cluster Setup (FLARE Consumer)

### Step 1: Install Liqo

```bash
# Install Liqo first (required by FLUIDOS)
cd fluidos/node/tools/scripts
./install_liqo.sh kubeadm flare-hub-cluster $KUBECONFIG /usr/local/bin/liqoctl
```

### Step 2: Label Nodes

```bash
# Label nodes for FLUIDOS (replace <node-name> with actual node names)
kubectl get nodes  # List available nodes
kubectl label node <node-name> node-role.fluidos.eu/worker=true

# Verify labeling was successful
kubectl get nodes --show-labels | grep fluidos
```

### Step 3: Install FLUIDOS (Consumer Mode)

```bash
helm repo add fluidos https://fluidos-project.github.io/node/

# Determine your public IP (example methods):
# HUB_PUBLIC_IP=$(curl -s ifconfig.me)
# Or manually set: HUB_PUBLIC_IP="your.external.ip"
# Choose an available port for REAR protocol (default: 30000-32767 range)
# REAR_PORT="30080"

helm upgrade --install node fluidos/node \
  -n fluidos --create-namespace \
  --set localResourceManager.config.enableAutoDiscovery=false \
  --set networkManager.config.enableLocalDiscovery=false \
  --set networkManager.configMaps.nodeIdentity.ip="<HUB_PUBLIC_IP>" \
  --set rearController.service.gateway.nodePort.port="<REAR_PORT>" \
  --wait
```

### Step 4: Install FLARE Components

```bash
# Install FLARE API Gateway
kubectl apply -f flare-deployment.yaml
```

### Step 5: Configure Hub Broker Connection

```bash
# Use FLUIDOS broker creation script
cd fluidos/node/tools/scripts
./broker-creation.sh

# Script will interactively prompt for:
# - Broker name: flare-federation-broker
# - Server address: <provided by broker administrator>
# - Root cert file: /path/to/ca.pem
# - Client cert file: /path/to/hub-client.pem
# - Private key file: /path/to/hub-client-key.pem
# - Role: subscriber
# - Rule JSON file: /path/to/hub-rule.json (see JSON examples below)
# - Metric JSON file: /path/to/hub-metric.json (see JSON examples below)
```

### Step 6: Verify Hub Installation

```bash
kubectl get pods -n fluidos
kubectl get brokers -n fluidos
kubectl get knownclusters -n fluidos  # Discovered providers
```

## GPU Provider Setup

### Step 1: Label Nodes

```bash
# Label nodes for FLUIDOS (replace <node-name> with actual node names)
kubectl get nodes  # List available nodes
kubectl label node <node-name> node-role.fluidos.eu/worker=true

# For resource-providing nodes
kubectl label node <node-name> node-role.fluidos.eu/resources=true

# Verify labeling was successful
kubectl get nodes --show-labels | grep fluidos
```

### Step 2: Install Liqo

```bash
# Install Liqo first (required by FLUIDOS)
cd fluidos/node/tools/scripts
./install_liqo.sh kubeadm gpu-provider-cluster $KUBECONFIG /usr/local/bin/liqoctl
```

### Step 3: Install FLUIDOS (Provider Mode)

```bash
helm repo add fluidos https://fluidos-project.github.io/node/

helm upgrade --install node fluidos/node \
  -n fluidos --create-namespace \
  --set localResourceManager.config.enableAutoDiscovery=true \
  --set networkManager.config.enableLocalDiscovery=false \
  --set networkManager.configMaps.nodeIdentity.ip="<PROVIDER_PUBLIC_IP>" \
  --set rearController.service.gateway.nodePort.port="<REAR_PORT>" \
  --wait
```

### Step 4: Annotate GPU Nodes

```bash
# Annotate GPU nodes with specifications (for auto-discovery)
kubectl annotate node <gpu-node> \
  gpu.fluidos.eu/vendor="nvidia" \
  gpu.fluidos.eu/model="nvidia-h100" \
  gpu.fluidos.eu/count="8" \
  gpu.fluidos.eu/memory="80Gi" \
  location.fluidos.eu/region="us-east-1" \
  cost.fluidos.eu/hourly-rate="4.50"
```

### Step 5: Configure Provider Broker Connection

```bash
./broker-creation.sh

# Script will interactively prompt for:
# - Broker name: flare-federation-broker
# - Server address: <same as hub, from broker administrator>
# - Root cert file: /path/to/ca.pem (same CA as hub)
# - Client cert file: /path/to/provider-client.pem
# - Private key file: /path/to/provider-client-key.pem
# - Role: publisher
# - Rule JSON file: /path/to/provider-rule.json (see JSON examples below)
# - Metric JSON file: /path/to/provider-metric.json (see JSON examples below)
```

### Step 6: Verify Provider Installation

```bash
kubectl get pods -n fluidos
kubectl get flavors -n fluidos    # Auto-created GPU flavors
kubectl get brokers -n fluidos
```

## Broker Requirements

**Required from broker administrator**:

- Broker server address (must match certificate CN)
- CA root certificate (.pem file)
- Client certificates (.pem files) and private keys for each cluster
- Use FLUIDOS `broker-creation.sh` script for interactive configuration

**Prepare JSON files** (required by script):

Example hub-rule.json:

```json
{
  "gpu": true,
  "federation": "flare"
}
```

Example hub-metric.json:

```json
{
  "consumer_hub": true,
  "flare_enabled": true
}
```

Example provider-rule.json:

```json
{
  "gpu_types": ["H100"],
  "cloud": "provider-a"
}
```

Example provider-metric.json:

```json
{
  "gpu_count": 8,
  "region": "us-east-1",
  "pricing": "4.50/hour"
}
```

## Verification

```bash
# Check discovered providers on hub
kubectl get knownclusters -n fluidos

# Test resource discovery
kubectl apply -f - <<EOF
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: test-discovery
  namespace: fluidos  
spec:
  intentID: test-001
  selector:
    flavorType: K8Slice
  findCandidate: true
  reserveAndBuy: false
  establishPeering: false
EOF

kubectl get peeringcandidates -n fluidos
```

## Monitoring

```bash
# Check federation status
kubectl get brokers -n fluidos
kubectl get knownclusters -n fluidos  
kubectl get flavors -n fluidos

# Monitor allocations
kubectl get allocations -n fluidos
kubectl get discoveries -n fluidos
```

## Troubleshooting

```bash
# No providers discovered
kubectl logs -n fluidos deployment/node-network-manager | grep broker
kubectl describe brokers -n fluidos

# No GPU flavors created  
kubectl get nodes --show-labels | grep fluidos
kubectl logs -n fluidos deployment/node-local-resource-manager

# REAR connection issues
kubectl logs -n fluidos deployment/node-rear-controller
kubectl get services -n fluidos | grep gateway

# Liqo issues
liqoctl info
```

For development setup with KinD clusters, see [FLARE QuickStart Guide](FLARE_QuickStart_Guide.md)