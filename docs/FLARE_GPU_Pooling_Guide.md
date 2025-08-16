# FLARE GPU Pooling Guide

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Development Timeline Context](#development-timeline-context)
4. [Workflow Evolution](#workflow-evolution)
   - [Workflow 1: Manual Process (Original FLUIDOS)](#workflow-1-manual-process-original-fluidos)
   - [Workflow 2: Semi-Automated (Enhanced FLUIDOS)](#workflow-2-semi-automated-enhanced-fluidos)
   - [Workflow 3: Fully Automated (FLARE)](#workflow-3-fully-automated-flare)
5. [Workflow Comparison](#workflow-comparison)
6. [GPU Annotation Reference](#gpu-annotation-reference)
7. [Use Cases and Examples](#use-cases-and-examples)
8. [Troubleshooting](#troubleshooting)
9. [Cleanup](#cleanup)

## Overview

This guide demonstrates the evolution of GPU resource pooling from complex manual processes to full automation. Will track three distinct workflows that showcase how FLARE transforms GPU orchestration into simple intent-based operations.

### The GPU Pooling Challenge

GPU resources are expensive and often underutilized (30%+ idle time in typical deployments). GPU pooling enables:

- Dynamic sharing of idle GPU resources across providers
- Cost optimization through efficient utilization
- Seamless access to specialized GPU models when needed
- Transparent cross-cluster GPU workload execution

### FLARE's Solution

FLARE transforms the complex, multi-step process of GPU workload allocation into a single API call.

## Prerequisites

### GPU Node Requirements

Providers must have GPU-enabled worker nodes with appropriate annotations. This guide uses simulated GPUs via `fake-gpu-operator` for demonstration.

### Expected Node Configuration

```yaml
# Provider 1 - NVIDIA A100 GPUs (simulated)
apiVersion: v1
kind: Node
metadata:
  labels:
    node-role.fluidos.eu/resources: "true"  # Mark as FLUIDOS resource
    nvidia.com/gpu.product: NVIDIA-A100     # GPU model (from operator)
    nvidia.com/gpu.memory: "40960"          # GPU memory in MiB
    nvidia.com/gpu.count: "2"               # Number of GPUs
```

```yaml
# Provider 2 - NVIDIA H100 GPUs (simulated)
apiVersion: v1
kind: Node
metadata:
  labels:
    node-role.fluidos.eu/resources: "true"  # Mark as FLUIDOS resource
    nvidia.com/gpu.product: NVIDIA-H100     # GPU model (from operator)
    nvidia.com/gpu.memory: "81920"          # GPU memory in MiB
    nvidia.com/gpu.count: "2"               # Number of GPUs
```

### GPU Annotation Setup

FLARE expects GPU nodes to have `gpu.fluidos.eu/*` annotations. Providers must translate hardware-specific labels to FLARE annotations:

```bash
# Provider 1: Annotate nodes with A100 GPU properties
export KUBECONFIG="fluidos-provider-1-config"

# Get worker nodes and annotate with GPU properties
WORKER_NODES=$(kubectl get nodes --no-headers -o custom-columns=":metadata.name" | grep worker)

for NODE in $WORKER_NODES; do
  kubectl annotate node $NODE \
    gpu.fluidos.eu/model="nvidia-a100" \
    gpu.fluidos.eu/count="2" \
    gpu.fluidos.eu/memory="40Gi"
done

# Verify annotations
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU-MODEL:.metadata.annotations.gpu\\.fluidos\\.eu/model,GPU-COUNT:.metadata.annotations.gpu\\.fluidos\\.eu/count,GPU-MEMORY:.metadata.annotations.gpu\\.fluidos\\.eu/memory
```

```bash
# Provider 2: Annotate nodes with H100 GPU properties
export KUBECONFIG="fluidos-provider-2-config"

WORKER_NODES=$(kubectl get nodes --no-headers -o custom-columns=":metadata.name" | grep worker)

for NODE in $WORKER_NODES; do
  kubectl annotate node $NODE \
    gpu.fluidos.eu/model="nvidia-h100" \
    gpu.fluidos.eu/count="2" \
    gpu.fluidos.eu/memory="80Gi"
done
```

## Development Timeline Context

- **Workflow 1 (Manual)**: Original FLUIDOS without GPU support
- **Workflow 2 (Semi-Automated)**: After FLUIDOS GPU enhancements  
- **Workflow 3 (FLARE)**: Complete FLARE implementation

## Workflow Evolution

### Workflow 1: Manual Process (Original FLUIDOS)

This workflow demonstrates the original FLUIDOS capabilities before GPU enhancements, requiring manual intervention at each step.

#### Step 1: Manual GPU Flavor Creation

Since original FLUIDOS didn't automatically detect GPUs, this workflow shows manual Flavor patching:

```bash
# Provider 1: Add GPU properties to existing Flavors
export KUBECONFIG="fluidos-provider-1-config"

# Get all Flavors and patch with GPU characteristics
FLAVORS=$(kubectl get flavors -n fluidos --no-headers -o custom-columns=":metadata.name")

for FLAVOR in $FLAVORS; do
  kubectl patch flavor $FLAVOR -n fluidos --type merge -p '{
    "spec": {
      "flavorType": {
        "typeData": {
          "characteristics": {
            "gpu": {
              "model": "nvidia-a100",
              "count": 2,
              "memory": "40Gi"
            }
          }
        }
      }
    }
  }'
done

# Verify GPU properties were added
kubectl get flavor -n fluidos -o json | jq '.items[].spec.flavorType.typeData.characteristics.gpu'
```

```bash
# Provider 2: Add H100 GPU properties
export KUBECONFIG="fluidos-provider-2-config"

FLAVORS=$(kubectl get flavors -n fluidos --no-headers -o custom-columns=":metadata.name")

for FLAVOR in $FLAVORS; do
  kubectl patch flavor $FLAVOR -n fluidos --type merge -p '{
    "spec": {
      "flavorType": {
        "typeData": {
          "characteristics": {
            "gpu": {
              "model": "nvidia-h100",
              "count": 2,
              "memory": "80Gi"
            }
          }
        }
      }
    }
  }'
done
```

#### Step 2: Create Basic GPU Solver

Create a Solver to discover GPU resources (without native filtering):

```bash
# Consumer: Create GPU Solver
export KUBECONFIG="fluidos-consumer-1-config"

kubectl apply -f - <<EOF
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: gpu-solver-a100
  namespace: fluidos
spec:
  selector:
    flavorType: K8Slice
    filters: {}  # No GPU filtering available
  intentID: "gpu-intent-a100"
  findCandidate: true      # Start discovery
  reserveAndBuy: false     # Manual reservation
  establishPeering: false  # Manual peering
EOF

# Monitor discovery
kubectl get solver gpu-solver-a100 -n fluidos -w
```

#### Step 3: Manual GPU Filtering

Since FLUIDOS doesn't filter by GPU, manually inspect and filter PeeringCandidates:

```bash
# List all discovered candidates
kubectl get peeringcandidates -n fluidos

# Manually inspect each candidate for GPU properties
for pc in $(kubectl get peeringcandidates -n fluidos -o name); do
  PC_NAME=$(basename $pc)
  echo "=== Analyzing $PC_NAME ==="
  
  # Check for GPU characteristics
  kubectl get $pc -n fluidos -o jsonpath='{.spec.flavor.spec.flavorType.typeData.characteristics.gpu}' | jq
  echo ""
done

# Remove non-GPU candidates from solver interest
CANDIDATES=$(kubectl get peeringcandidates -n fluidos --no-headers -o custom-columns=":metadata.name")

for CANDIDATE in $CANDIDATES; do
  GPU_MODEL=$(kubectl get peeringcandidate $CANDIDATE -n fluidos -o jsonpath='{.spec.flavor.spec.flavorType.typeData.characteristics.gpu.model}')
  
  if [[ "$GPU_MODEL" != *"a100"* ]]; then
    echo "Removing $CANDIDATE - GPU model: $GPU_MODEL"
    kubectl patch peeringcandidate $CANDIDATE -n fluidos --type='merge' -p='{"spec":{"interestedSolverIDs":[]}}'
  fi
done
```

#### Step 4: Manual Reservation

After filtering, manually trigger reservation:

```bash
# Enable reservation for GPU resources
kubectl patch solver gpu-solver-a100 -n fluidos --type merge -p '{"spec": {"reserveAndBuy": true}}'

# Monitor reservation and contract creation
kubectl get solver gpu-solver-a100 -n fluidos -w
kubectl get contracts -n fluidos
```

#### Step 5: Manual Peering

Establish cluster peering manually:

```bash
# Enable peering to create virtual node
kubectl patch solver gpu-solver-a100 -n fluidos --type merge -p '{"spec": {"establishPeering": true}}'

# Monitor virtual node creation
kubectl get nodes | grep liqo
```

#### Step 6: Manual Namespace Setup

Create and configure namespace for GPU workload:

```bash
# Create namespace
kubectl create namespace workload-test

# Configure namespace offloading
kubectl apply -f - <<EOF
apiVersion: offloading.liqo.io/v1beta1
kind: NamespaceOffloading
metadata:
  name: offloading
  namespace: workload-test
spec:
  clusterSelector:
    nodeSelectorTerms:
    - matchExpressions:
      - key: liqo.io/remote-cluster-id
        operator: In
        values:
        - fluidos-provider-1
  namespaceMappingStrategy: DefaultName
  podOffloadingStrategy: Remote
EOF
```

#### Step 7: Manual Workload Deployment

Deploy GPU workload manually:

```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-inference
  namespace: workload-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gpu-inference
  template:
    metadata:
      labels:
        app: gpu-inference
    spec:
      containers:
      - name: inference
        image: nvcr.io/nvidia/pytorch:23.10-py3
        command: ["python", "-c"]
        args:
        - |
          import torch
          import time
          
          print("Checking GPU availability...")
          if torch.cuda.is_available():
              device = torch.cuda.get_device_name(0)
              count = torch.cuda.device_count()
              memory = torch.cuda.get_device_properties(0).total_memory / 1024**3
              print(f"GPU Available: {device}")
              print(f"GPU Count: {count}")
              print(f"GPU Memory: {memory:.2f} GB")
              
              # Simple inference simulation
              print("\nRunning inference simulation...")
              x = torch.randn(1000, 1000).cuda()
              for i in range(10):
                  y = torch.matmul(x, x)
                  torch.cuda.synchronize()
                  print(f"Iteration {i+1} completed")
                  time.sleep(1)
              print("Inference simulation completed!")
          else:
              print("No GPU available!")
          
          # Keep pod running for inspection
          while True:
              time.sleep(30)
              print("Pod is running...")
        resources:
          requests:
            nvidia.com/gpu: "1"
          limits:
            nvidia.com/gpu: "1"
EOF

# Verify GPU workload is running
kubectl get pods -n workload-test -o wide
kubectl logs -n workload-test deployment/gpu-inference
```

### Workflow 2: Semi-Automated (Enhanced FLUIDOS)

With FLUIDOS GPU enhancements, several steps become automatic.

#### Step 1: Automatic GPU Flavor Creation

Enhanced FLUIDOS automatically creates GPU Flavors from node annotations:

```bash
# Provider: GPU Flavors created automatically from annotations
export KUBECONFIG="fluidos-provider-1-config"

# Verify automatic GPU Flavor creation
kubectl get flavors -n fluidos -o custom-columns=NAME:.metadata.name,GPU-MODEL:.spec.flavorType.typeData.characteristics.gpu.model,GPU-MEMORY:.spec.flavorType.typeData.characteristics.gpu.memory
```

#### Step 2: Enhanced GPU Solver with Filtering

Create Solver with native GPU filtering:

```bash
# Consumer: Create enhanced GPU Solver
export KUBECONFIG="fluidos-consumer-1-config"

kubectl apply -f - <<EOF
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: gpu-solver-enhanced
  namespace: fluidos
spec:
  selector:
    flavorType: K8Slice
    filters:
      cpuFilter:
        name: Range
        data:
          min: "2"
      memoryFilter:
        name: Range
        data:
          min: "4Gi"
      # GPU filters (FLUIDOS enhancement)
      gpuFilters:
      - field: model
        filter: StringFilter
        data:
          value: nvidia-a100
      - field: count
        filter: NumberRangeFilter
        data:
          min: 1
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: 40Gi
  intentID: "gpu-intent-enhanced"
  findCandidate: true      # Automatic discovery with GPU filtering
  reserveAndBuy: true      # Automatic reservation
  establishPeering: true   # Automatic peering
EOF
```

#### Step 3: Automatic Discovery and Federation

FLUIDOS automatically:

- Discovers GPU resources matching filters
- Reserves the best match
- Establishes peering
- Creates virtual GPU node

```bash
# Monitor automatic progression
kubectl get solver gpu-solver-enhanced -n fluidos -w

# Verify virtual GPU node
kubectl get nodes | grep liqo
```

#### Step 4: Manual Namespace and Workload (Still Required)

```bash
# Create namespace
kubectl create namespace workload-gpu

# Configure offloading (manual)
kubectl apply -f - <<EOF
apiVersion: offloading.liqo.io/v1beta1
kind: NamespaceOffloading
metadata:
  name: offloading
  namespace: workload-gpu
spec:
  podOffloadingStrategy: Remote
EOF

# Deploy workload (manual)
kubectl -n workload-gpu create deployment gpu-inference --image=nvcr.io/nvidia/pytorch:23.10-py3
```

### Workflow 3: Fully Automated (FLARE)

FLARE transforms the entire process into a single API call.

#### The FLARE Difference

Instead of multiple manual steps, users submit their intent:

```bash
# Complete GPU workload deployment with one API call
curl -X POST https://flare-api.fluidos.eu/api/v1/intents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "gpu-inference-job",
    "objective": "Performance_Maximization",
    "requirements": {
      "gpu": {
        "model": "nvidia-a100",
        "count": 1,
        "memory": "40Gi"
      },
      "cpu": {
        "cores": 2
      },
      "memory": {
        "size": "8Gi"
      }
    },
    "workload": {
      "type": "deployment",
      "replicas": 1,
      "containers": [{
        "name": "inference",
        "image": "nvcr.io/nvidia/pytorch:23.10-py3",
        "command": ["python", "inference.py"],
        "resources": {
          "requests": {
            "nvidia.com/gpu": "1"
          }
        }
      }]
    }
  }'

# Response
{
  "intent_id": "intent-gpu-abc123",
  "status": "pending",
  "message": "GPU workload deployment initiated"
}
```

#### Behind the Scenes

FLARE automatically handles:

1. **GPU Flavor Discovery**: Queries all federated providers for matching GPUs
2. **Solver Creation**: Generates FLUIDOS Solver with appropriate GPU filters
3. **Resource Selection**: Evaluates candidates based on objective (Performance/Cost/Latency)
4. **Contract Negotiation**: Reserves GPU resources through REAR protocol
5. **Cluster Peering**: Establishes Liqo connection to provider
6. **Namespace Setup**: Creates Capsule-managed namespace with offloading
7. **Workload Deployment**: Deploys containers on virtual GPU node

#### Status Monitoring

```bash
# Check intent status
curl https://flare-api.fluidos.eu/api/v1/intents/intent-gpu-abc123

# Response shows progress
{
  "intent_id": "intent-gpu-abc123",
  "status": "running",
  "phase": "WORKLOAD_DEPLOYED",
  "allocation": {
    "provider": "fluidos-provider-1",
    "gpu": {
      "model": "nvidia-a100",
      "count": 1,
      "memory": "40Gi"
    },
    "virtual_node": "liqo-fluidos.eu-k8slice-gpu-xyz"
  },
  "workload": {
    "namespace": "flare-intent-gpu-abc123",
    "deployment": "gpu-inference",
    "status": "running",
    "pods": [{
      "name": "gpu-inference-5d4b9c-x7z",
      "status": "running",
      "node": "liqo-fluidos.eu-k8slice-gpu-xyz"
    }]
  }
}
```

## Workflow Comparison

| Task | Manual (Current) | Semi-Automated | FLARE |
|------|-----------------|----------------|-------|
| Node Annotation | Manual kubectl | Manual kubectl | Provider admin |
| GPU Flavor Creation | Manual patch | **Automatic** | **Automatic** |
| Solver Creation | Manual apply | Manual apply | **Automatic** |
| GPU Filtering | Manual inspect | **Automatic** | **Automatic** |
| Reservation | Manual patch | **Automatic** | **Automatic** |
| Peering | Manual patch | **Automatic** | **Automatic** |
| Namespace Setup | Manual create | Manual create | **Automatic** |
| Offloading Config | Manual apply | Manual apply | **Automatic** |
| Workload Deploy | Manual apply | Manual apply | **Automatic** |
| **Total Steps** | **8+ manual** | **4 manual** | **1 API call** |
| **Time to Deploy** | 15-20 minutes | 5-10 minutes | < 1 minute |
| **Expertise Required** | High (Kubernetes/FLUIDOS) | Medium | **Low (API only)** |

## GPU Annotation Reference

For comprehensive GPU annotation specifications and examples, see the [FLARE GPU Annotations Reference](FLARE_GPU_Annotations_Reference.md) document. This reference provides:

- Complete annotation namespace and format specifications
- GPU vendor mapping details (NVIDIA, AMD)
- Validation rules and constraints
- Real-world annotation examples
- API field mapping information

## Use Cases and Examples

### Inference Workload

Typical inference workload requiring single GPU:

```json
{
  "name": "llm-inference",
  "requirements": {
    "gpu": {
      "model": "nvidia-a100",
      "count": 1,
      "memory": "40Gi"
    }
  },
  "objective": "Latency_Minimization"
}
```

### Training Workload

Multi-GPU training job:

```json
{
  "name": "model-training",
  "requirements": {
    "gpu": {
      "model": "nvidia-h100",
      "count": 4,
      "memory": "80Gi"
    }
  },
  "objective": "Performance_Maximization"
}
```

### Cost-Optimized Batch Processing

Batch job with flexible GPU requirements:

```json
{
  "name": "batch-processing",
  "requirements": {
    "gpu": {
      "count": 1,
      "memory": "24Gi"
    }
  },
  "objective": "Cost_Minimization"
}
```

## Troubleshooting

### Common Issues

#### No GPU Flavors Created

- **Cause**: Nodes missing `gpu.fluidos.eu/*` annotations
- **Solution**: Verify node annotations are applied correctly
- **Check**: `kubectl get nodes -o yaml | grep gpu.fluidos.eu`

#### GPU Workload Not Scheduled

- **Cause**: No matching GPU resources available
- **Solution**: Relax GPU requirements or wait for resources
- **Check**: `kubectl get peeringcandidates -n fluidos`

#### Virtual Node Missing GPU Capacity

- **Cause**: Contract doesn't include GPU specifications
- **Solution**: Verify Flavor has GPU characteristics
- **Check**: `kubectl get contract -n fluidos -o yaml`

### Debugging Commands

```bash
# Check GPU annotations on nodes
kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.metadata.annotations.gpu\\.fluidos\\.eu/model

# Verify GPU Flavors
kubectl get flavors -n fluidos -o json | jq '.items[].spec.flavorType.typeData.characteristics.gpu'

# Check Solver GPU filters
kubectl get solver -n fluidos -o yaml | grep -A20 gpuFilters

# Inspect PeeringCandidates for GPU
kubectl get peeringcandidates -n fluidos -o json | jq '.items[] | select(.spec.flavor.spec.flavorType.typeData.characteristics.gpu != null)'

# Check virtual node GPU capacity
kubectl get nodes -l liqo.io/type=virtual-node -o custom-columns=NAME:.metadata.name,GPU:.status.capacity.nvidia\\.com/gpu
```

## Cleanup

Remove test resources after completion:

```bash
# Delete GPU workloads
kubectl delete namespace workload-test workload-gpu

# Delete GPU solvers
kubectl delete solver gpu-solver-a100 gpu-solver-enhanced -n fluidos

# Clean up federation resources
kubectl delete allocations --all -n fluidos
kubectl delete contracts --all -n fluidos
kubectl delete peeringcandidates --all -n fluidos
```

