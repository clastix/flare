# FLARE (Federated Liquid Resources Exchange) Architecture Document

## Table of Contents

1. [Project Overview](#project-overview)
2. [Overall Architecture](#overall-architecture)
3. [Key Components](#key-components)
4. [Deployment Architecture](#deployment-architecture)
5. [Core Workflows](#core-workflows)

## Project Overview

FLARE (Federated Liquid Resources Exchange) is a GPU pooling platform for AI and HPC applications. Built on FLUIDOS, it enables dynamic GPU sharing across cloud providers through intent-based allocation and federated architecture.

### Key Objectives

- **Dynamic GPU Pooling**: Real-time sharing of idle GPU resources across providers
- **Intent-Based Orchestration**: High-level goal specification without technical complexity
- **Federated Architecture**: Cross-provider resource federation while preserving provider autonomy
- **Cost Optimization**: Operational cost reduction through improved utilization
- **High Utilization**: Improve GPU utilization rates

## Overall Architecture

```
┌─────────────────┐
│      Users      │ (ML Training, Inference, Video Processing, HPC)
└────────┬────────┘
         │ Submit Intents
         ▼
┌────────────────────────────────────────┐
│          FLARE API Gateway             │
|    (RESTful Interface for Intents)     │
└────────────────────┬───────────────────┘
                     │
                     ▼
┌────────────────────────────────────────┐
│      FLUIDOS Consumer Cluster          │
│  (Solver, Discovery, Contract, Alloc)  │
└────────────────────┬───────────────────┘
                     │ REAR Protocol
                     ▼
┌────────────────────────────────────────┐
│        GPU Provider Clusters           │
│  ┌──────────┐  ┌──────────┐  ┌──────┐  │
│  │ Provider │  │ Provider │  │ ...  │  │
│  │    1     │  │    2     │  │      │  │
│  └──────────┘  └──────────┘  └──────┘  │
│   Each GPU node = Specialized Flavor   │
└────────────────────────────────────────┘
```

## Key Components

### FLARE Platform

FLARE consists of multiple components working together to provide GPU orchestration:

**1. FLARE API Gateway**

- Primary interface for client workloads
- RESTful APIs for intent submission
- Authentication and authorization (Capsule multi-tenant support)
- Returns intent IDs for tracking

**2. Intent Processor**

- Translates user intents into FLUIDOS Solver specifications
- Maps workload requirements to GPU filters
- Applies optimization strategies based on objectives
- Validates resource availability against quotas

**3. Resource Controller**

- Monitors FLUIDOS Solver status and progression
- Tracks Contract and Allocation lifecycle
- Handles resource cleanup on intent deletion
- Manages multi-provider resource coordination

**4. Workload Controller**

- Creates and manages Kubernetes namespaces with Capsule integration
- Deploys NamespaceOffloading CRs for remote execution
- Creates workload resources (Deployments, Services, ConfigMaps)
- Monitors workload health and status

**5. Status Manager**

- Provides real-time intent status via API
- Aggregates status from multiple controllers
- Manages webhook notifications for status changes
- Maintains intent history and audit logs

The FLARE API exposes three main resources:

- **Intents** (`/api/v1/intents`) - Submit and manage GPU workloads
- **Resources** (`/api/v1/resources`) - Query available GPU resources
- **Tokens** (`/api/v1/auth/tokens`) - Manage API authentication

Refer to the [FLARE API Reference](FLARE_API_Reference.md) for complete specifications.

### FLUIDOS Integration

FLARE extends FLUIDOS with GPU-specific capabilities for resource management and orchestration. FLARE required the following enhancements to FLUIDOS core functionality to enable full GPU support:

- **GPU Flavor Creation**: Automatically create specialized Flavors for GPU worker nodes based on node annotations.
- **GPU Filtering in Solvers**: Enable Solvers to filter GPU resources based on user intents.
- **Multiple Virtual Nodes**: Enable multiple virtual nodes targeting the same remote provider cluster. 

#### GPU Flavor Creation

- Each GPU worker node becomes a unique Flavor
- FLUIDOS includes GPU specifications from node annotations

#### GPU Filtering in Solvers

- User Intents are translated by FLARE into Solvers with GPU filters
- FLUIDOS Discovery Manager applies these filters during candidate search
- Only matching resources become PeeringCandidates
- FLUIDOS automatically selects the first available match

#### Multiple Virtual Nodes

- Each GPU worker node in a provider cluster becomes a unique Flavor
- Each Flavor can generate a separate Contract when requested
- Each Contract creates its own ResourceSlice and VirtualNode
- Set the relationship: Worker Node → Flavor → Contract → Virtual Node 

## Deployment Architecture

### FLARE Deployment Model

FLARE is deployed as a set of controllers in the FLUIDOS consumer cluster:

**Deployment Configuration:**

- **Namespace**: `flare-system` (separate from `fluidos` namespace)
- **Components**: Deployed as separate Deployments for scalability
- **High Availability**: Each component supports multiple replicas

**Integration Points:**

- FLARE controllers access FLUIDOS CRs via Kubernetes API
- No direct communication with FLUIDOS controllers
- Shares same cluster RBAC and networking

## Core Workflows

### 1. GPU Provider Setup Flow

**Prerequisites**: FLARE and FLUIDOS expect GPU worker nodes to be pre-annotated with GPU characteristics by the cluster provider administrator.

1. **GPU Worker Node Joins Cluster**

- Physical GPU node joins provider cluster
- GPU device plugins (NVIDIA, AMD) installed and running
- Node has GPU resources exposed (e.g., `nvidia.com/gpu: 4`)

2. **Node Annotation by Cluster Provider Admin**

- Admin must annotate nodes with GPU specifications
- Annotations used by FLUIDOS for specialized Flavor creation and efficient filtering
- Example annotation command:

```bash
# Example: Annotating a node with NVIDIA H100 GPUs
# See [FLARE_GPU_Annotations_Reference.md](FLARE_GPU_Annotations_Reference.md) Quick Reference Table for all available annotations
kubectl annotate node gpu-worker-1 \
  gpu.fluidos.eu/model="nvidia-h100" \
  gpu.fluidos.eu/count="4" \
  gpu.fluidos.eu/memory="80Gi" \
  gpu.fluidos.eu/tier="premium" \
  gpu.fluidos.eu/interconnect="nvlink" \
  gpu.fluidos.eu/architecture="hopper" \
  location.fluidos.eu/region="eu-west-1" \
  cost.fluidos.eu/hourly-rate="2" \
  cost.fluidos.eu/currency="EUR" \
  workload.fluidos.eu/training-score="0.98"

# Example: Annotating a node with AMD GPUs  
kubectl annotate node gpu-worker-2 \
  gpu.fluidos.eu/model="amd-mi300x" \
  gpu.fluidos.eu/count="2" \
  gpu.fluidos.eu/memory="192Gi" \
  gpu.fluidos.eu/tier="premium" \
  gpu.fluidos.eu/interconnect="infinity-fabric" \
  gpu.fluidos.eu/architecture="cdna3" \
  location.fluidos.eu/region="us-east-1" \
  cost.fluidos.eu/hourly-rate="2" \
  cost.fluidos.eu/currency="EUR" \
  workload.fluidos.eu/hpc-score="0.98"

```

> Note: Many annotations can be auto-generated from GPU operator labels. See [NVIDIA_GPU_Labels_Mapping.md](NVIDIA_GPU_Labels_Mapping.md) and [AMD_GPU_Labels_Mapping.md](AMD_GPU_Labels_Mapping.md) for examples.

3. **Automatic GPU Flavor Creation**

- FLUIDOS Node Controller monitors nodes with `node-role.fluidos.eu/resources="true"`
- Reads FLARE annotations from node metadata during monitoring cycle
- Creates specialized Flavor for each GPU worker node
- GPU specifications from annotations stored in Flavor's `gpu` field
   
**Example Generated Flavor:**

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Flavor
metadata:
  name: k8slice-provider1-gpu-node-abc123
  namespace: fluidos
spec:
  flavorType:
    typeIdentifier: K8Slice
    typeData:
      characteristics:
        architecture: "amd64"
        cpu: "64"                    # From node allocatable
        memory: "256Gi"              # From node allocatable
        pods: "110"                  # From node allocatable
        gpu:                         # From FLARE annotations
          model: "nvidia-h100"
          count: "4"
          memory: "80Gi"
          tier: "premium"
          architecture: "hopper"
          interconnect: "nvlink"
      properties:
        latency: 50                  # Network latency in ms
      policies:
        partitionability:
          cpuMin: "2"
          memoryMin: "8Gi"
          gpuMin: "1"
  owner:
    domain: "provider1.fluidos.eu"
    nodeID: "provider-1"
  price:
    amount: "0"              # not used for FLARE
    currency: "EUR"          # not used for FLARE
    period: "hourly"         # not used for FLARE
  availability: true
```

4. **Resource Advertisement**

- GPU Flavor advertised via REAR protocol
- Available to consumer clusters for discovery (timing depends on network conditions)
- Updates propagated when node annotations change

### 2. Basic GPU Allocation Flow

This flow demonstrates the complete end-to-end process from user intent submission to running GPU workload.

1. **Workload Submits Intent**

Users submit workload intents using a simple, Kubernetes-agnostic format:

```json
{
  "intent": {
    "objective": "Performance_Maximization",
    "workload": {
      "type": "service",
      "name": "deepseek-r1-nvidia",
      "image": "lmsysorg/sglang:latest",
      "commands": [
        "python3 -m sglang.launch_server --model-path $MODEL_ID --port 8000 --trust-remote-code"
      ],
      "env": [
        "MODEL_ID=deepseek-ai/DeepSeek-R1-Distill-Llama-8B",
        "CUDA_VISIBLE_DEVICES=0,1"
      ],
      "ports": [
        {
          "port": 8000,
          "expose": true
        }
      ],
      "resources": {
        "cpu": "8",
        "memory": "32Gi",
        "gpu": {
          "count": 2,
          "memory": "80Gi"
        }
      },
      "storage": {
        "volumes": [
          {
            "name": "model-cache",
            "size": "100Gi",
            "type": "persistent",
            "path": "/root/.cache"
          }
        ]
      }
    },
    "constraints": {
      "max_hourly_cost": "10 EUR",
      "location": "EU",
      "max_latency_ms": 50
    },
    "sla": {
      "availability": "99.9%"
    }
  }
}
```

See [FLARE API Reference](FLARE_API_Reference.md) for complete intent schema.

2. **FLARE API Gateway Receives Intent**

- Validates workload specification against schema
- Authenticates user via Bearer token and checks Capsule tenant quotas
- Assigns unique intent ID for tracking
- Parses resource requirements and constraints
- Returns intent ID to user for status tracking

3. **Intent Processing**

- Intent Processor extracts requirements from workload spec
- Maps those requirements to FLUIDOS filter specifications
- Prepares complete Solver specification with filters and automation flags

Intent-to-Solver Mapping examples:

**Example 1: Basic GPU Request**

```json
// API Intent
"resources": {
  "gpu": {
    "model": "nvidia-a100",
    "count": 2,
    "memory": "40Gi"
  }
},
"constraints": {
  "location": "EU",
  "max_hourly_cost": "20 EUR"
}
```

this is translated as:

```yaml
# Generated Solver
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: solver-intent-abc123
  namespace: fluidos
  labels:
    flare.io/intent-id: "abc123"
    flare.io/user: "user@example.com"
    flare.io/tenant: "tenant-a"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: model
        filter: StringFilter
        data:
          value: nvidia-a100
      - field: count
        filter: NumberRangeFilter
        data:
          min: 2
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: 40Gi
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 20
      locationFilter:
        name: StringFilter
        data:
          value: EU
  findCandidate: true
  reserveAndBuy: true    # Automatic progression
  establishPeering: true # Complete automation
```

**Example 2: Performance Optimization**

```json
// API Intent
"objective": "Performance_Maximization",
"resources": {
  "gpu": {
    "tier": "premium",
    "interconnect": "nvlink"
  }
}
```

this is translated as:

```yaml
# Generated Solver
spec:
  selector:
    filters:
      gpuFilters:
      - field: tier
        filter: StringFilter
        data:
          value: premium
      - field: interconnect
        filter: StringFilter
        data:
          value: nvlink
```

4. **GPU Discovery and Matching**
   
**Discovery Process:**

- FLUIDOS Discovery Manager reads Solver spec and initiates search
- Queries all known providers via REAR protocol gateways
- Each provider returns available GPU Flavors matching basic criteria
- Discovery creates PeeringCandidate CR for each matching Flavor
   
**GPU Filter Application**:

- **Model Filter**: Matches exact GPU model (e.g., "nvidia-a100")
- **Count Filter**: Ensures sufficient GPU units available
- **Memory Filter**: Validates per-GPU memory meets minimum
- **Tier Filter**: Matches performance tier (premium/standard)
- **Location Filter**: Ensures provider in allowed regions
- **Cost Filter**: Excludes Flavors exceeding budget

**PeeringCandidate Example:**

```yaml
apiVersion: advertisement.fluidos.eu/v1alpha1
kind: PeeringCandidate
metadata:
  name: pc-gpu-a100-provider1
  namespace: fluidos
spec:
  flavor:
    spec:
      flavorType:
        typeData:
          characteristics:
            gpu:
              model: "nvidia-a100"
              count: "4"
              memory: "40Gi"
              tier: "premium"
  solverID: "solver-intent-abc123"
  available: true
```

**Discovery Status Updates:**

- Solver status shows discovery phase: `Discovering` → `DiscoveryCompleted`
- Number of candidates found reflected in Solver status
- FLARE monitors via label selector: `flare.io/intent-id=abc123`

5. **Resource Selection and Reservation**

- FLUIDOS selects best candidate based on objective criteria
- Selection based on objective criteria (implementation depends on FLUIDOS enhancements)
- Cost considerations require pricing integration with REAR protocol
- REAR protocol initiates negotiation with selected provider

6. **Contract Establishment**

- REAR protocol completes negotiation with provider
- Contract CR created with agreed terms and Liqo credentials
- Allocation CR automatically created referencing the Contract
- FLARE monitors Contract status for successful establishment

7. **Cluster Peering Establishment**

- FLUIDOS Allocation controller triggers Liqo peering
- Virtual GPU node appears in consumer cluster
- Node capacity matches Contract specifications
- Network connectivity established via Liqo tunnels
- Virtual node ready for workload scheduling

8. **Namespace and Multi-tenancy Setup**

- FLARE creates namespace: `flare-<tenant>-<workload-name>`
- Applies tenant isolation labels (integration with multi-tenancy systems like Capsule)
- Creates NamespaceOffloading CR for automatic remote scheduling:

```yaml
apiVersion: offloading.liqo.io/v1beta1
kind: NamespaceOffloading
metadata:
  name: offloading
  namespace: flare-tenant-a-deepseek-r1
spec:
  namespaceMappingStrategy: EnforceSameName
  podOffloadingStrategy: Remote  # Ensures GPU pods run on virtual nodes
```

9. **Workload Deployment**

FLARE creates Kubernetes resources based on intent:

- Deployment/Job with GPU resource requests
- ConfigMaps for environment variables
- PersistentVolumeClaims for storage
- Services for exposed ports
- Pods automatically scheduled on virtual GPU nodes
- No manual node selection needed due to NamespaceOffloading

10. **Service Access and Status Tracking**

- FLARE exposes service endpoint (if ports specified)
- Returns access information via status API:

```json
{
  "status": {
    "service_url": "https://flare-tenant-a-deepseek-r1.example.com:8000",
    "intent_id": "abc123",
    "namespace": "flare-tenant-a-deepseek-r1"
  }
}
```
- User accesses GPU workload via provided endpoint

### 3. No GPU Requirements Met Flow

When discovery finds no matching GPUs:

**Option A: Queue and Wait**

- Subscribe to Discovery updates
- Wait for GPU availability notifications
- Automatic allocation when resources become available

**Option B: Fallback Suggestions**

- Suggest alternative GPU models (RTX 4090 instead of H100)
- Offer lower-tier options with adjusted pricing
- Present cross-region options with higher latency

### 4. GPU Resource Contention Flow

When multiple workloads compete for the same GPU:

1. **Priority Evaluation** 

  - Currently: First-come, first-served via FLUIDOS
  - Future enhancements could include:
  - Intent priorities (critical vs. best-effort)
  - Price-based selection
  - SLA-aware allocation

2. **Winner Selection**

  - Create reservation for winner
  - Block resource for exclusive use

3. **Loser Handling**

  - Offer next-best GPU alternatives
  - Add to priority queue for next availability
  - Suggest spot instances if applicable


---