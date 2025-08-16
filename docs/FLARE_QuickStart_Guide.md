# FLARE QuickStart Guide

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Development Environment Setup](#development-environment-setup)
4. [Configuration Summary](#configuration-summary)
5. [Troubleshooting](#troubleshooting)
6. [Cleanup](#cleanup)
7. [Next Steps](#next-steps)

## Overview

This guide sets up a FLARE development environment using KinD (Kubernetes in Docker) clusters with LAN Discovery enabled. The setup creates one consumer hub cluster and two GPU provider clusters (each with 3 total nodes: 1 control-plane + 2 workers) for testing FLARE's multi-provider federation.

> **Reference Video**: [Basic FLUIDOS Workflow Demo](https://www.youtube.com/watch?v=aNdrWRgOz1o) - Visual demonstration of core FLUIDOS concepts and workflows that apply also to FLARE implementation.

> **Note**: This setup uses **Fake GPU Operator** for GPU simulation (not real GPUs) and **KinD clusters** for local development (not cloud providers).

## Prerequisites

Before starting, verify all tools are installed and resources are adequate:

```bash
# Check versions
docker --version  # Should be v28.1.1+
kubectl version --client  # Should be v1.33.0+
helm version  # Should be v3.17.3+
kind version  # Should be v0.27.0+
liqoctl version --client  # Should be v1.0.0+

# Verify Docker is running
docker ps

# Check available resources
free -h  # Should show at least 12Gi total (10Gi+ available)
df -h    # Should show at least 30Gi free space

# Check Docker resource allocation
docker system info | grep -E "(CPUs|Total Memory)"
```

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
# Clone FLUIDOS GPU node fork
git clone https://github.com/clastix/fluidos-node node
# Once merged use the upstream repository instead
# git clone https://github.com/fluidos-project/node

# Navigate to setup scripts directory
cd node/tools/scripts

# Run setup script
./setup.sh
```

**Understanding the options:**

- **Option 2** (Custom KIND): Allows multiple clusters vs single demo cluster
- **Consumer clusters (1)**: Hub cluster that will host FLARE API server
- **Provider clusters (2)**: GPU providers for federation testing
- **Local repositories (y/n)**: 
  - `y` for development (builds images locally)
  - `n` for testing (uses pre-built images from ghcr.io)
- **Resource auto discovery (y)**: Automatically creates Flavors from node resources
- **LAN node discovery (y)**: Enables multicast discovery between clusters

### Step 3: Configure Consumer as Non-Provider Hub

Remove resource labels from consumer nodes (hub should not provide resources):

```bash
# Remove resource labels from consumer nodes
export KUBECONFIG=fluidos-consumer-1-config
kubectl label node fluidos-consumer-1-worker node-role.fluidos.eu/resources-
kubectl label node fluidos-consumer-1-worker2 node-role.fluidos.eu/resources-

# Delete any auto-created flavors on consumer
kubectl delete flavors --all -n fluidos

# Verify consumer has no flavors
kubectl get flavors -n fluidos
```

#### Step 3 Verification

```bash
# Verify no resource labels
export KUBECONFIG=fluidos-consumer-1-config
kubectl get nodes -l node-role.fluidos.eu/resources
# Should return: No resources found

# Verify no flavors
kubectl get flavors -n fluidos
# Should return: No resources found in fluidos namespace
```

### Step 4: Install Fake GPU Operator on Providers

```bash
# Install fake-gpu-operator on Provider 1 with H100 configuration
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
  --set topology.nodePools.default.gpuMemory="81920"  # 80GiB in MiB 

# Annotate nodes to simulate GPU resources
kubectl label node fluidos-provider-1-worker run.ai/simulated-gpu-node-pool=default
kubectl label node fluidos-provider-1-worker2 run.ai/simulated-gpu-node-pool=default

# Install fake-gpu-operator on Provider 2 with A6000 configuration  
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
  --set topology.nodePools.default.gpuMemory="49152"  # 48GiB in MiB 

# Annotate nodes to simulate GPU resources
kubectl label node fluidos-provider-2-worker run.ai/simulated-gpu-node-pool=default
kubectl label node fluidos-provider-2-worker2 run.ai/simulated-gpu-node-pool=default
```

#### Step 4 Verification

```bash
# Check operator pods are running
kubectl get pods -n gpu-operator --kubeconfig fluidos-provider-1-config
kubectl get pods -n gpu-operator --kubeconfig fluidos-provider-2-config
# All pods should be Running

# Verify GPU resources
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu --kubeconfig fluidos-provider-1-config
# Should show 2 GPUs per worker node

# Monitor GPU availability if not immediately visible
watch -n 5 'kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu --kubeconfig fluidos-provider-1-config'
```

### Step 5: Annotate Provider Nodes for FLARE

For this FLARE intent demo, we use memory requirements to naturally select the appropriate GPU type:

- `memory: "70Gi"` - This requirement automatically selects H100s (80Gi) while excluding A6000s (48Gi)
- `count: 1` - Request a single GPU (we have 2 per node)

```bash
# Basic GPU annotations for Provider 1 - H100 nodes
export KUBECONFIG=fluidos-provider-1-config

# Worker 1: H100 with 80Gi memory per GPU
kubectl annotate node fluidos-provider-1-worker \
  gpu.fluidos.eu/model="nvidia-h100" \
  gpu.fluidos.eu/count="2" \
  gpu.fluidos.eu/memory="80Gi"

# Worker 2: H100 with 80Gi memory per GPU  
kubectl annotate node fluidos-provider-1-worker2 \
  gpu.fluidos.eu/model="nvidia-h100" \
  gpu.fluidos.eu/count="2" \
  gpu.fluidos.eu/memory="80Gi"

# Basic GPU annotations for Provider 2 - A6000 nodes
export KUBECONFIG=fluidos-provider-2-config

# Worker 1: A6000 with 48Gi memory per GPU
kubectl annotate node fluidos-provider-2-worker \
  gpu.fluidos.eu/model="nvidia-a6000" \
  gpu.fluidos.eu/count="2" \
  gpu.fluidos.eu/memory="48Gi"

# Worker 2: A6000 with 48Gi memory per GPU
kubectl annotate node fluidos-provider-2-worker2 \
  gpu.fluidos.eu/model="nvidia-a6000" \
  gpu.fluidos.eu/count="2" \
  gpu.fluidos.eu/memory="48Gi"

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
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable."nvidia\\.com/gpu"
kubectl get flavors -n fluidos

export KUBECONFIG=fluidos-provider-2-config
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable."nvidia\\.com/gpu"
kubectl get flavors -n fluidos

# Verify consumer has no flavors (hub only)
export KUBECONFIG=fluidos-consumer-1-config
kubectl get flavors -n fluidos
```

### Step 7: Install FLARE Components

```bash
# Switch to consumer cluster
export KUBECONFIG=fluidos-consumer-1-config

# Install FLARE API Gateway
kubectl create namespace flare-system
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flare-api-server
  namespace: flare-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: flare-api-server
  template:
    metadata:
      labels:
        app: flare-api-server
    spec:
      containers:
      - name: api-server
        image: clastix/flare-api-server:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: flare-api-service
  namespace: flare-system
spec:
  selector:
    app: flare-api-server
  ports:
  - port: 8080
    targetPort: 80
    nodePort: 30080
  type: NodePort
EOF

# Wait for FLARE API Gateway to be ready
kubectl wait --for=condition=available deployment/flare-api-server -n flare-system --timeout=300s
```

### Step 8: Configure FLARE API Authentication

For this demo, we use simple token-based authentication. In production, integrate with your organization's identity provider.

```bash
export KUBECONFIG=fluidos-consumer-1-config

# Create demo authentication token (for development only)
kubectl create secret generic flare-demo-tokens \
  --namespace flare-system \
  --from-literal=demo-token="$(echo -n "demo-user:demo-permissions" | base64 -w 0)"

# Set up token for demo usage
export FLARE_TOKEN="demo-token"
echo "FLARE_TOKEN=demo-token" > ~/.flare-env

# Test API connectivity and authentication 
FLARE_API_URL="http://localhost:30080"
curl -X GET ${FLARE_API_URL}/api/v1/auth/verify \
  -H "Authorization: Bearer ${FLARE_TOKEN}" \
  --connect-timeout 10

# Expected response: {"status": "authenticated", "user": "demo-user"}
```

**Note**: This demo token has basic permissions. In production:

- Use your organization's identity provider (OIDC, LDAP, etc.)  
- Implement proper RBAC with tenant isolation
- Use short-lived tokens with refresh capability

### Step 9: (Optional) Install FLUIDOS Node Dashboard

```bash
# Install FLUIDOS dashboard on consumer cluster for visual monitoring
export KUBECONFIG=fluidos-consumer-1-config

# Clone the dashboard repository
git clone https://github.com/fluidos-project/fluidos-node-dashboard.git
cd fluidos-node-dashboard

# Install the dashboard
./install.sh kind fluidos-consumer-1

# Create required ConfigMap for network manager (if not exists)
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluidos-network-manager-config
  namespace: fluidos
data:
  local: ""
EOF

# Get the dashboard service details
kubectl get svc | grep dashboard

# Port-forward to access the dashboard locally
kubectl port-forward service/frontend-service 8080:80 &

# For WSL users: Access from Windows host browser
# Use WSL IP address instead of localhost
# kubectl port-forward service/frontend-service 8080:80 --address 0.0.0.0 &
# kubectl port-forward service/backend-service 31000:3001 --address 0.0.0.0 &
# Then access at http://WSL_IP:8080 (e.g., http://172.30.180.176:8080)

# Access dashboard at http://localhost:8080
```

### Step 10: Complete End-to-End FLARE GPU Workload Deployment

This section demonstrates the fully automated FLARE workflow: from intent submission to GPU workload deployment with a single API call.

#### Submit GPU Intent via FLARE API

```bash
export KUBECONFIG=fluidos-consumer-1-config

# Load authentication token (from Step 8)
source ~/.flare-env
FLARE_API_URL="http://localhost:30080"

# Submit GPU workload intent
curl -X POST ${FLARE_API_URL}/api/v1/intents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${FLARE_TOKEN}" \
  -d '{
    "intent": {
      "objective": "Performance_Maximization",
      "workload": {
        "type": "service",
        "name": "gpu-demo-apache",
        "image": "httpd:2.4-alpine",
        "commands": [
          "httpd-foreground"
        ],
        "ports": [
          {
            "port": 80,
            "expose": true
          }
        ],
        "resources": {
          "cpu": "1",
          "memory": "512Mi",
          "gpu": {
            "count": 1,
            "memory": "70Gi"
          }
        }
      },
      "constraints": {
        "max_hourly_cost": "2 EUR",
        "location": "EU",
        "max_latency_ms": 100
      },
      "sla": {
        "availability": "99.0%"
      }
    }
  }'

# Expected response:
# {
#   "intent_id": "gpu-demo-abc123",
#   "status_url": "/api/v1/intents/gpu-demo-abc123/status",
#   "message": "GPU workload deployment initiated"
# }
```

#### Understanding GPU Selection

This intent uses memory requirements to naturally select the appropriate GPU type:

**Memory-Based Filtering**: `"memory": "70Gi"` - Requires at least 70GiB of GPU memory

**How FLARE Filters GPU Providers:**

```bash
# Provider 1 (H100): ✅ SELECTED
# - Memory: 80Gi ✓ (exceeds 70Gi requirement)
# - Model: nvidia-h100 (naturally selected due to memory)
# - Result: Available for selection

# Provider 2 (A6000): ❌ EXCLUDED  
# - Memory: 48Gi ❌ (below 70Gi requirement)  
# - Model: nvidia-a6000 (filtered out due to insufficient memory)
# - Result: Automatically excluded

echo "FLARE will select Provider-1 (H100) based on memory requirements"

# Memory requirement scenarios:
echo "=== Memory Requirement Examples ==="
echo "• memory: '40Gi' → Both providers match (H100: 80Gi, A6000: 48Gi)"
echo "• memory: '50Gi' → Only Provider-1 matches (H100: 80Gi > 50Gi)"  
echo "• memory: '70Gi' → Only Provider-1 matches (H100: 80Gi > 70Gi) [CURRENT]"
echo "• memory: '90Gi' → No providers match, intent would fail"
```

#### Track Intent Progress

```bash
# Check intent status (replace with actual intent_id from response)
INTENT_ID="gpu-demo-abc123"
curl ${FLARE_API_URL}/api/v1/intents/${INTENT_ID}/status

# Monitor automated workflow progress
# FLARE automatically:
# 1. Creates GPU-aware Solver with 70Gi memory filter
# 2. Discovers GPU resources across providers  
# 3. Filters out Provider-2 (A6000 with insufficient 48Gi memory)
# 4. Reserves Provider-1 (H100 with 80Gi matching criteria)
# 5. Establishes cluster peering via Liqo
# 6. Creates namespace with multi-tenant configuration
# 7. Deploys workload on virtual GPU node

# Check FLUIDOS resources created by FLARE
kubectl get solvers -n fluidos -l flare.io/intent-id=${INTENT_ID}
kubectl get peeringcandidates -n fluidos
kubectl get contracts -n fluidos
kubectl get allocations -n fluidos

# Verify Provider-1 selection (80Gi memory requirement)
echo "=== Verifying GPU Selection Based on Memory ==="
kubectl get allocations -n fluidos -o jsonpath='{.items[0].spec.remoteClusterID}' | grep provider-1 && echo "✓ Provider-1 selected (80Gi memory)" || echo "❌ Unexpected provider selected"

# Check the virtual node characteristics
kubectl get nodes | grep liqo | head -1 | xargs kubectl describe node | grep -E "(provider-1|80Gi)" || echo "Virtual node created from high-memory GPU provider"
```

#### Verify GPU Workload Deployment

```bash
# Check FLARE-managed namespace
kubectl get namespaces | grep flare-

# Check workload deployment on virtual GPU node
FLARE_NAMESPACE=$(kubectl get namespaces -o name | grep flare- | head -1 | cut -d/ -f2)
kubectl get pods -n ${FLARE_NAMESPACE} -o wide

# Expected output:
# NAME                          READY   STATUS    NODE
# gpu-demo-apache-xxx           1/1     Running   liqo-fluidos.eu-k8slice-provider-1

# Verify pod is running on high-memory GPU virtual node  
echo "=== Confirming GPU Workload Deployment ==="
POD_NAME=$(kubectl get pods -n ${FLARE_NAMESPACE} --no-headers -o custom-columns=":metadata.name" | head -1)
kubectl get pod ${POD_NAME} -n ${FLARE_NAMESPACE} -o wide | grep provider-1 && echo "✓ Workload deployed on Provider-1 (80Gi GPU)" || echo "❌ Unexpected provider"

# Verify GPU resource request in workload
kubectl get pod ${POD_NAME} -n ${FLARE_NAMESPACE} -o jsonpath='{.spec.containers[0].resources.requests.nvidia\\.com/gpu}' | grep -q "1" && echo "✓ GPU resource requested" || echo "❌ No GPU resource found"

# Check NamespaceOffloading configuration
kubectl get namespaceoffloading -n ${FLARE_NAMESPACE}

# Test workload accessibility
kubectl get services -n ${FLARE_NAMESPACE}

# Port-forward to test the Apache service
kubectl port-forward -n ${FLARE_NAMESPACE} service/gpu-demo-apache 8080:80 &

# Test the service
curl http://localhost:8080
# Expected: Apache default page

# Stop port-forward
pkill -f "kubectl port-forward"
```

#### Verify Cross-Cluster GPU Federation

```bash
# Check virtual node representing remote GPU resources
kubectl get nodes | grep liqo

# Verify GPU capacity on virtual node
VIRTUAL_NODE=$(kubectl get nodes -o name | grep liqo | head -1 | cut -d/ -f2)
kubectl describe node ${VIRTUAL_NODE} | grep -A10 "Capacity:"

# Expected to see:
# nvidia.com/gpu: 2

# Check Liqo ForeignCluster connection
kubectl get foreignclusters

# Verify cross-cluster connectivity
liqoctl info

# Check which provider cluster is serving the GPU
kubectl get allocation -n fluidos -o yaml | grep -A5 "destination:"
```


## Configuration Summary

| Cluster | Role | Resources | GPU Type | GPUs/Node |
|---------|------|-----------|----------|-----------|
| fluidos-consumer-1 | FLARE Hub | None (labels removed) | - | - |
| fluidos-provider-1 | GPU Provider | 2 worker nodes | H100 | 2 |
| fluidos-provider-2 | GPU Provider | 2 worker nodes | A6000 | 2 |

**Key Configuration Steps Applied:**

1. Consumer nodes de-labeled (no `node-role.fluidos.eu/resources`)
2. Consumer flavors deleted (hub-only mode)
3. Fake GPU operator installed on providers only
4. Provider nodes annotated with `gpu.fluidos.eu/*` labels per FLARE spec
5. Different GPU memory capacities per provider (80Gi vs 48Gi) enable memory-based selection

## Troubleshooting

### Insufficient Resources

**Symptoms:**

- Pods stuck in Pending state
- Slow cluster startup (>15 minutes)
- OOMKilled errors
- High CPU usage (>80%)

**Solutions:**

```bash
# Check resource usage
free -h && docker stats --no-stream

# If memory is low (<2Gi available):
# 1. Increase WSL memory (see Prerequisites section)
# 2. Or use reduced setup with fewer clusters

# Clean up resources
docker system prune -a --volumes
```

### Image Pull Issues

If you encounter image pull errors from ghcr.io:

1. **Check connectivity:**

```bash
docker exec fluidos-consumer-1-control-plane curl -I https://ghcr.io
```

2. **Pre-pull critical images:**

```bash
# Pre-pull images to local Docker, just as example
docker pull ghcr.io/liqotech/fabric:v1.0.0

# Load into KinD clusters
for cluster in fluidos-consumer-1 fluidos-provider-1 fluidos-provider-2; do
  kind load docker-image ghcr.io/liqotech/fabric:v1.0.0 --name $cluster
done
```

### GPU Resources Not Showing

After installing fake-gpu-operator, GPUs may take 2-3 minutes to appear:

```bash
# Monitor GPU availability
watch -n 5 'kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu --kubeconfig fluidos-provider-1-config'
```

### Pods Stuck in ContainerCreating

Check events for specific errors:
```bash
kubectl get events -n gpu-operator --sort-by='.lastTimestamp' --kubeconfig fluidos-provider-1-config
```

### FLARE API Gateway Issues

**FLARE API Gateway Not Starting**

```bash
# Check FLARE API server pod status
kubectl get pods -n flare-system
kubectl describe pod -l app=flare-api-server -n flare-system
kubectl logs -l app=flare-api-server -n flare-system

# Check service connectivity
kubectl get svc -n flare-system
curl -v http://localhost:30080/health || echo "FLARE API not accessible"
```

**Intent Submission Failures**

```bash
# Verify API endpoint is accessible  
source ~/.flare-env  # Load authentication token
FLARE_API_URL="http://localhost:30080"
curl -v ${FLARE_API_URL}/api/v1/health

# Check common intent submission issues
curl -X POST ${FLARE_API_URL}/api/v1/intents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${FLARE_TOKEN}" \
  -d '{"test": "connection"}' -v

# Expected responses:
# 200 OK: API working
# 400 Bad Request: Invalid JSON format
# 401 Unauthorized: Authentication issue
# 404 Not Found: Wrong endpoint
# 503 Service Unavailable: FLARE components not ready
```

### GPU Workload Deployment Issues

**Workload Not Scheduled on Virtual GPU Node**

```bash
# Check if virtual nodes are available
kubectl get nodes | grep liqo
# Should show virtual nodes from provider clusters

# Check NamespaceOffloading configuration
FLARE_NAMESPACE=$(kubectl get namespaces -o name | grep flare- | head -1 | cut -d/ -f2)
kubectl get namespaceoffloading -n ${FLARE_NAMESPACE}
kubectl describe namespaceoffloading -n ${FLARE_NAMESPACE}

# Check pod events if stuck
kubectl get events -n ${FLARE_NAMESPACE} --sort-by='.lastTimestamp'

# Verify pod GPU resource requests
kubectl get pod -n ${FLARE_NAMESPACE} -o yaml | grep -A10 "resources:"
```

**No GPU Resources Available**

```bash
# Check if providers have GPU resources
echo "Provider-1 GPU availability:"
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu --kubeconfig fluidos-provider-1-config

echo "Provider-2 GPU availability:"
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu --kubeconfig fluidos-provider-2-config

# Check FLUIDOS resource discovery
kubectl get peeringcandidates -n fluidos
kubectl get flavors -n fluidos --all-namespaces

# Check GPU annotations on provider nodes
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU-MODEL:.metadata.annotations.gpu\\.fluidos\\.eu/model --kubeconfig fluidos-provider-1-config
```

**Cross-Cluster Connectivity Issues**

```bash
# Check Liqo connectivity
liqoctl info

# Check ForeignCluster status
kubectl get foreignclusters
kubectl describe foreignclusters

# Check Allocation status
kubectl get allocations -n fluidos
kubectl describe allocation -n fluidos

# Verify virtual node connectivity
VIRTUAL_NODE=$(kubectl get nodes -o name | grep liqo | head -1 | cut -d/ -f2)
kubectl describe node ${VIRTUAL_NODE}
```

**Intent Status Tracking Issues**

```bash
# List all intents if status endpoint not working
kubectl get solvers -n fluidos -l flare.io/intent-id
kubectl get contracts -n fluidos
kubectl get allocations -n fluidos

# Check FLARE-managed namespaces
kubectl get namespaces | grep flare-

# Monitor resource progression manually
kubectl get solver,peeringcandidate,contract,allocation -n fluidos -w
```

## Cleanup

### Clean Up FLARE Workloads and Components

```bash
# Switch to consumer cluster
export KUBECONFIG=fluidos-consumer-1-config

# Stop any running port-forwards
pkill -f "kubectl port-forward"

# Clean up FLARE-managed namespaces and workloads
echo "=== Cleaning up FLARE workloads ==="
kubectl get namespaces | grep flare- | awk '{print $1}' | xargs -r kubectl delete namespace

# Clean up FLARE API Gateway
echo "=== Cleaning up FLARE API Gateway ==="
kubectl delete deployment flare-api-server -n flare-system --ignore-not-found=true
kubectl delete service flare-api-service -n flare-system --ignore-not-found=true
kubectl delete namespace flare-system --ignore-not-found=true

# Clean up authentication files
rm -f ~/.flare-env

# Clean up FLUIDOS resources created by FLARE
echo "=== Cleaning up FLUIDOS resources ==="
kubectl delete solvers --all -n fluidos --ignore-not-found=true
kubectl delete peeringcandidates --all -n fluidos --ignore-not-found=true
kubectl delete allocations --all -n fluidos --ignore-not-found=true
kubectl delete reservations --all -n fluidos --ignore-not-found=true
kubectl delete contracts --all -n fluidos --ignore-not-found=true

# Clean up Liqo ForeignClusters
kubectl delete foreignclusters --all --ignore-not-found=true

# Verify cleanup
echo "=== Verifying FLARE cleanup ==="
kubectl get namespaces | grep flare- || echo "No FLARE namespaces remaining"
kubectl get pods -n flare-system 2>/dev/null || echo "FLARE system namespace removed"
kubectl get solvers,contracts,allocations -n fluidos || echo "FLUIDOS resources cleaned"
```

### Clean Up GPU Providers and Annotations

```bash
# Clean up GPU annotations from provider nodes
echo "=== Cleaning up Provider-1 GPU annotations ==="
export KUBECONFIG=fluidos-provider-1-config
for node in $(kubectl get nodes --no-headers -o custom-columns=":metadata.name" | grep worker); do
  kubectl annotate node $node \
    gpu.fluidos.eu/model- \
    gpu.fluidos.eu/count- \
    gpu.fluidos.eu/memory- \
    node.fluidos.eu/gpu- --ignore-not-found=true
done

echo "=== Cleaning up Provider-2 GPU annotations ==="
export KUBECONFIG=fluidos-provider-2-config
for node in $(kubectl get nodes --no-headers -o custom-columns=":metadata.name" | grep worker); do
  kubectl annotate node $node \
    gpu.fluidos.eu/model- \
    gpu.fluidos.eu/count- \
    gpu.fluidos.eu/memory- \
    node.fluidos.eu/gpu- --ignore-not-found=true
done
```

### Remove GPU Operators Only

```bash
helm uninstall fake-gpu-operator -n gpu-operator --kubeconfig fluidos-provider-1-config
helm uninstall fake-gpu-operator -n gpu-operator --kubeconfig fluidos-provider-2-config

# Clean up GPU operator namespaces
kubectl delete namespace gpu-operator --kubeconfig fluidos-provider-1-config --ignore-not-found=true
kubectl delete namespace gpu-operator --kubeconfig fluidos-provider-2-config --ignore-not-found=true
```

### Remove FLUIDOS Only

```bash
helm uninstall node -n fluidos --kubeconfig fluidos-consumer-1-config
helm uninstall node -n fluidos --kubeconfig fluidos-provider-1-config  
helm uninstall node -n fluidos --kubeconfig fluidos-provider-2-config

# Clean up FLUIDOS namespaces
kubectl delete namespace fluidos --kubeconfig fluidos-consumer-1-config --ignore-not-found=true
kubectl delete namespace fluidos --kubeconfig fluidos-provider-1-config --ignore-not-found=true
kubectl delete namespace fluidos --kubeconfig fluidos-provider-2-config --ignore-not-found=true
```

### Remove Liqo Only

```bash
# Uninstall Liqo from all clusters
liqoctl uninstall --kubeconfig fluidos-consumer-1-config
liqoctl uninstall --kubeconfig fluidos-provider-1-config
liqoctl uninstall --kubeconfig fluidos-provider-2-config

# Clean up Liqo namespaces
kubectl delete namespace liqo --kubeconfig fluidos-consumer-1-config --ignore-not-found=true
kubectl delete namespace liqo --kubeconfig fluidos-provider-1-config --ignore-not-found=true
kubectl delete namespace liqo --kubeconfig fluidos-provider-2-config --ignore-not-found=true
```

### Complete Cleanup

```bash
# Clean up all KinD clusters and resources
cd node/tools/scripts
./clean-dev-env.sh

# Remove any remaining kubeconfig files
rm -f fluidos-consumer-1-config fluidos-provider-1-config fluidos-provider-2-config

# Verify complete cleanup
kind get clusters | grep fluidos || echo "All FLUIDOS clusters removed"
docker ps | grep fluidos || echo "No FLUIDOS containers running"
```

### Verification Script

Save as `verify-cleanup.sh`:

```bash
#!/bin/bash
echo "=== FLARE Cleanup Verification ==="

echo -e "\n1. Checking KinD clusters..."
kind get clusters | grep fluidos || echo "✓ No FLUIDOS clusters found"

echo -e "\n2. Checking Docker containers..."
docker ps | grep fluidos || echo "✓ No FLUIDOS containers running"

echo -e "\n3. Checking kubeconfig files..."
ls -la fluidos-*-config 2>/dev/null || echo "✓ No kubeconfig files found"

echo -e "\n4. Checking local Docker images..."
docker images | grep -E "(fluidos|fake-gpu|liqo)" || echo "✓ No local images (optional cleanup)"

echo -e "\n=== Cleanup Verification Complete ==="
```

## Next Steps

- For production deployment, see [FLARE Admin Guide](FLARE_Admin_Guide.md)
- For real GPU hardware setup, configure actual GPU nodes
- For testing FLARE intents, deploy FLARE API Gateway components

This multi-provider quickstart environment provides comprehensive testing of FLARE's federation capabilities across multiple simulated GPU providers.