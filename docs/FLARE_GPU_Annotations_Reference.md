# FLARE GPU Annotations Reference

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Quick Reference Table](#quick-reference-table)
4. [Detailed Annotation Specifications](#detailed-annotation-specifications)
5. [Core GPU Annotations (Required)](#core-gpu-annotations-required)
6. [Location Annotations (Required - Manual)](#location-annotations-required---manual)
7. [Cost Annotations (Required - Manual)](#cost-annotations-required---manual)
8. [Performance Annotations (Manual - Optional)](#performance-annotations-manual---optional)
9. [GPU Sharing Annotations (Manual - Optional)](#gpu-sharing-annotations-manual---optional)
10. [Network Performance Annotations (Optional)](#network-performance-annotations-optional)
11. [Communication-Aware Annotations (Optional)](#communication-aware-annotations-optional)
12. [Provider Annotations (Optional)](#provider-annotations-optional)
13. [Annotation Examples](#annotation-examples)
14. [Querying Nodes with Annotations](#querying-nodes-with-annotations)
15. [Migration from Labels to Annotations](#migration-from-labels-to-annotations)
16. [FLARE API Mapping](#flare-api-mapping)
17. [Supported GPU Vendors](#supported-gpu-vendors)
18. [Validation Rules](#validation-rules)

## Overview

This document is the authoritative reference for FLARE GPU annotations. It defines the complete annotation specification for GPU resource management in the FLARE/FLUIDOS ecosystem.

## Quick Start

### For Administrators

```bash
# FLARE expects these annotations to be present on GPU worker nodes:
kubectl annotate node gpu-worker-1 \
  gpu.fluidos.eu/vendor="nvidia" \
  gpu.fluidos.eu/model="nvidia-a100" \
  gpu.fluidos.eu/count="4" \
  gpu.fluidos.eu/memory="40Gi" \
  gpu.fluidos.eu/tier="premium" \
  location.fluidos.eu/region="eu-west-1" \
  cost.fluidos.eu/hourly-rate="1.2" \
  cost.fluidos.eu/currency="EUR"
```


### Annotation Namespaces

| Namespace | Purpose | Examples | Required |
|-----------|---------|----------|----------|
| `gpu.fluidos.eu/*` | GPU hardware specs | vendor, model, count, memory, tier, cores, interconnect, topology, multi_gpu_efficiency | Yes |
| `location.fluidos.eu/*` | Geographic location | region, zone | Yes |
| `cost.fluidos.eu/*` | Pricing information | hourly-rate, currency | Yes |
| `workload.fluidos.eu/*` | Performance scores | training-score, inference-score, hpc-score | Optional |
| `network.fluidos.eu/*` | Network performance | bandwidth-gbps, latency-ms, tier | Optional |  
| `provider.fluidos.eu/*` | Provider information | name, preemptible | Optional |

## Quick Reference Table

### Core GPU Annotations (Required)

| FLARE Annotation | Description | Unit | Sample Values | NVIDIA Label Source | AMD Label Source |
|-----------------|-------------|------|---------------|-------------------|------------------|
| `gpu.fluidos.eu/vendor` | GPU manufacturer | string | `"nvidia"`, `"amd"`, `"intel"` | Static: "nvidia" | Static: "amd" |
| `gpu.fluidos.eu/model` | Normalized GPU model | string | `"nvidia-h100"`, `"amd-mi300x"` | `nvidia.com/gpu.product` | `amd.com/gpu.product` |
| `gpu.fluidos.eu/count` | Number of GPUs | integer | `"1"`, `"2"`, `"4"`, `"8"` | `nvidia.com/gpu.count` | `amd.com/gpu.count` |
| `gpu.fluidos.eu/memory` | Memory per GPU | Gi | `"24Gi"`, `"40Gi"`, `"80Gi"`, `"192Gi"` | `nvidia.com/gpu.memory` (MB→Gi) | `amd.com/gpu.memory` (MB→Gi) |
| `gpu.fluidos.eu/tier` | Performance tier | enum | `"premium"`, `"standard"`, `"gaming"`, `"inference"`, `"budget"` | Inferred from model | Inferred from model |

### Location & Cost Annotations (Required)

| FLARE Annotation | Description | Unit | Sample Values |
|-----------------|-------------|------|---------------|
| `location.fluidos.eu/region` | Geographic region | string | `"eu-west-1"`, `"us-east-1"`, `"germany"` |
| `location.fluidos.eu/zone` | Availability zone | string | `"zone-a"`, `"zone-b"` (optional) |
| `cost.fluidos.eu/hourly-rate` | Hourly cost | float | `"0.1"`, `"1.2"`, `"5.0"` |
| `cost.fluidos.eu/currency` | Billing currency | ISO 4217 | `"EUR"` (implementation limitation) |

### Technical Specifications (Optional)

| FLARE Annotation | Description | Unit | Sample Values |
|-----------------|-------------|------|---------------|
| `gpu.fluidos.eu/architecture` | GPU architecture | string | `"hopper"`, `"ampere"`, `"ada-lovelace"`, `"cdna3"` |
| `gpu.fluidos.eu/interconnect` | GPU interconnect | enum | `"nvlink"`, `"infinity-fabric"`, `"pcie"`, `"xe-link"` |
| `gpu.fluidos.eu/cores` | GPU cores count | integer | `"10752"`, `"16896"`, `"304"` |
| `gpu.fluidos.eu/compute_capability` | Compute capability | version | `"8.0"`, `"8.6"`, `"9.0"`, `"9.4"` |
| `gpu.fluidos.eu/clock_speed` | Clock speed | G (Hz) | `"1.41G"`, `"2.23G"` |
| `gpu.fluidos.eu/fp32_tflops` | FP32 performance | TFLOPS | `"19.5"`, `"38.7"`, `"67.0"`, `"163.4"` |
| `gpu.fluidos.eu/interconnect_bandwidth` | Interconnect bandwidth | Gbps | `"600"`, `"900"`, `"64"` |
| `gpu.fluidos.eu/topology` | Multi-GPU topology | enum | `"all-to-all"`, `"nvswitch"`, `"ring"`, `"mesh"` |
| `gpu.fluidos.eu/multi_gpu_efficiency` | Multi-GPU efficiency | 0.0-1.0 | `"0.95"`, `"0.85"`, `"0.70"` |

### GPU Sharing Capabilities (Optional)

| FLARE Annotation | Description | Unit | Sample Values |
|-----------------|-------------|------|---------------|
| `gpu.fluidos.eu/shared` | Virtualization support | boolean | `"true"`, `"false"` |
| `gpu.fluidos.eu/sharing_strategy` | Sharing method | enum | `"mig"`, `"sriov"`, `"time-slicing"`, `"mps"`, `"none"` |
| `gpu.fluidos.eu/dedicated` | Dedicated allocation | boolean | `"true"`, `"false"` |
| `gpu.fluidos.eu/interruptible` | Spot instance support | boolean | `"true"`, `"false"` |

### Workload Scores (Optional)

| FLARE Annotation | Description | Unit | Sample Values |
|-----------------|-------------|------|---------------|
| `workload.fluidos.eu/training-score` | ML training suitability | 0.0-1.0 | `"0.95"`, `"0.85"`, `"0.70"` |
| `workload.fluidos.eu/inference-score` | ML inference suitability | 0.0-1.0 | `"0.90"`, `"0.80"`, `"0.60"` |
| `workload.fluidos.eu/hpc-score` | HPC suitability | 0.0-1.0 | `"0.98"`, `"0.95"`, `"0.80"` |
| `workload.fluidos.eu/graphics-score` | Graphics suitability | 0.0-1.0 | `"0.95"`, `"0.40"`, `"0.20"` |

### Network Performance (Optional)

| FLARE Annotation | Description | Unit | Sample Values |
|-----------------|-------------|------|---------------|
| `network.fluidos.eu/bandwidth-gbps` | Network bandwidth | Gbps | `"10"`, `"100"`, `"400"` |
| `network.fluidos.eu/latency-ms` | Network latency | ms | `"1"`, `"5"`, `"50"` |
| `network.fluidos.eu/tier` | Network tier | enum | `"premium"`, `"standard"`, `"basic"` |

### Provider Information (Optional)

| FLARE Annotation | Description | Unit | Sample Values |
|-----------------|-------------|------|---------------|
| `provider.fluidos.eu/name` | Provider identifier | string | `"aws"`, `"gcp"`, `"provider-1"` |
| `provider.fluidos.eu/preemptible` | Spot instance availability | boolean | `"true"`, `"false"` |

### Special Labels (Not Annotations)

| Label | Purpose | Unit | Value |
|-------|---------|------|-------|
| `node.fluidos.eu/gpu` | GPU node selector | boolean | **Fixed: `"true"`** |

## Detailed Annotation Specifications

The following sections provide detailed specifications for each annotation.

## Core GPU Annotations (Required)

### gpu.fluidos.eu/vendor

**Description**: GPU manufacturer  
**Values**: `nvidia`, `amd`, `intel`, `qualcomm`, `apple`, `custom`  
**Source**: Applied by provider (mapping documents provide reference)  
**Example**: `gpu.fluidos.eu/vendor="nvidia"`

### gpu.fluidos.eu/model

**Description**: Normalized GPU model identifier  
**Format**: Vendor-prefixed, lowercase, hyphenated  
**Source**: Applied by provider (mapping documents provide reference)  
**API Mapping**: Direct mapping to `resources.gpu.model`  
**Examples**:

- NVIDIA: `nvidia-h100`, `nvidia-a100`, `nvidia-rtx-4090`
- AMD: `amd-mi300x`, `amd-rx-7900xtx`  
- Intel: `intel-max-1550`, `intel-arc-a770`

### gpu.fluidos.eu/count

**Description**: Number of GPUs on the node  
**Format**: Integer string  
**Source**: Applied by provider  
**Examples**: `"1"`, `"2"`, `"4"`, `"8"`

### gpu.fluidos.eu/memory

**Description**: Memory per individual GPU  
**Format**: Kubernetes quantity (Gi)  
**Source**: Applied by provider  
**API Mapping**: Direct mapping to `resources.gpu.memory`  
**Examples**: `"12Gi"`, `"24Gi"`, `"80Gi"`

### gpu.fluidos.eu/tier

**Description**: Performance tier classification  
**Values**: `premium`, `standard`, `gaming`, `inference`, `budget`  
**Source**: Applied by provider  
**Mapping**:

- `premium`: Datacenter GPUs (H100, A100, MI300X, Max-1550)
- `standard`: Professional GPUs (RTX A6000, Pro W7900)
- `gaming`: Gaming/prosumer GPUs (RTX 4090, RX 7900XTX)
- `inference`: Inference-optimized (T4, Flex 170)
- `budget`: Entry-level/older GPUs

## Location Annotations (Required - Manual)

### location.fluidos.eu/region

**Description**: Geographic region or cloud region  
**Format**: Lowercase, hyphenated  
**Source**: Manual configuration by administrator  
**Examples**: `"eu-west-1"`, `"us-east-1"`, `"germany"`, `"singapore"`

### location.fluidos.eu/zone

**Description**: Availability zone within region  
**Format**: Lowercase  
**Source**: Manual configuration (optional)  
**Examples**: `"zone-a"`, `"zone-b"`

## Cost Annotations (Required - Manual)

### cost.fluidos.eu/hourly-rate

**Description**: Hourly cost rate in billing currency  
**Format**: `<rate>` (floating point)  
**Source**: Manual configuration by administrator  
**Examples**: `"1.2"`, `"5.0"`, `"0.3"`

### cost.fluidos.eu/currency

**Description**: Billing currency  
**Format**: ISO 4217 currency code  
**Source**: Manual configuration  
**Current Implementation**: Currently limited to `"EUR"` only

## Performance Annotations (Manual - Optional)

### gpu.fluidos.eu/architecture

**Description**: GPU architecture/generation  
**Source**: Manual configuration (derive from GPU operator family/architecture)  
**Examples**:

- NVIDIA: `"hopper"`, `"ada-lovelace"`, `"ampere"`
- AMD: `"rdna3"`, `"cdna3"`
- Intel: `"xe-hpg"`, `"ponte-vecchio"`

### gpu.fluidos.eu/interconnect

**Description**: GPU-to-GPU interconnect technology  
**Source**: Manual configuration (infer from model and count)  
**Values**: `nvlink`, `infinity-fabric`, `xe-link`, `pcie`

### gpu.fluidos.eu/fp32_tflops

**Description**: Single-precision floating point performance  
**Source**: Manual configuration (look up specs for model)  
**Format**: Decimal string  
**Examples**: `"83.0"`, `"19.5"`, `"35.6"`

### gpu.fluidos.eu/cores

**Description**: Number of GPU cores (CUDA cores, Stream processors, etc.)  
**Source**: Applied by provider  
**API Mapping**: Maps to `resources.gpu.cores`  
**Examples**: `"16384"` (RTX 4090), `"6912"` (A100), `"16896"` (H100)

### gpu.fluidos.eu/compute_capability

**Description**: CUDA compute capability or equivalent  
**Source**: Applied by provider  
**API Mapping**: Maps to `resources.gpu.compute_capability`  
**Examples**: `"9.0"` (H100), `"8.6"` (RTX 30 series), `"8.0"` (A100)

### gpu.fluidos.eu/clock_speed

**Description**: GPU clock speed  
**Source**: Applied by provider  
**API Mapping**: Maps to `resources.gpu.clock_speed`  
**Format**: Numeric with G suffix (representing Hz)  
**Examples**: `"1.41G"`, `"2.23G"`, `"1.98G"`

### Workload Optimization Scores

#### workload.fluidos.eu/training-score

**Description**: ML training suitability (0.0-1.0)  
**Source**: Manual configuration (assign based on GPU model capabilities)

#### workload.fluidos.eu/inference-score 

**Description**: ML inference suitability (0.0-1.0)  
**Source**: Manual configuration (assign based on GPU model capabilities)

#### workload.fluidos.eu/hpc-score

**Description**: HPC workload suitability (0.0-1.0)  
**Source**: Manual configuration (assign based on GPU model capabilities)

#### workload.fluidos.eu/graphics-score

**Description**: Graphics/rendering suitability (0.0-1.0)  
**Source**: Manual configuration (assign based on GPU model capabilities)

## GPU Sharing Annotations (Manual - Optional)

### gpu.fluidos.eu/shared

**Description**: Supports GPU virtualization/sharing  
**Values**: `"true"`, `"false"`  
**Source**: Manual configuration (check GPU operator capabilities)

### gpu.fluidos.eu/sharing_strategy

**Description**: GPU sharing method  
**Values**: `"mig"`, `"sriov"`, `"time-slicing"`, `"mps"`, `"none"`  
**Source**: Manual configuration (determine from GPU operator MIG/MPS capabilities)

**Implementation Status**:

- **Currently Supported**: `"none"` (dedicated allocation) and basic `"mig"` detection
- **Extension Possibilities**: Advanced sharing strategies (`"sriov"`, `"time-slicing"`, `"mps"`) would require additional GPU operator integration and cross-cluster coordination mechanisms
- **Federated Sharing**: Cross-provider GPU sharing coordination represents a potential platform enhancement

### gpu.fluidos.eu/dedicated

**Description**: Supports dedicated (non-shared) GPU allocation  
**Values**: `"true"`, `"false"`  
**Source**: Applied by provider  
**API Mapping**: Maps to `resources.gpu.dedicated`

### gpu.fluidos.eu/interruptible

**Description**: Supports spot/preemptible instances  
**Values**: `"true"`, `"false"`  
**Source**: Applied by provider  
**API Mapping**: Maps to `resources.gpu.interruptible`

## Network Performance Annotations (Optional)

### network.fluidos.eu/bandwidth-gbps

**Description**: Available network bandwidth  
**Examples**: `"10"`, `"100"`, `"400"`

### network.fluidos.eu/latency-ms

**Description**: Typical network latency  
**Examples**: `"1"`, `"5"`, `"50"`

### network.fluidos.eu/tier

**Description**: Network performance tier  
**Format**: Lowercase  
**Values**: `"premium"`, `"standard"`, `"basic"`

## Communication-Aware Annotations (Optional)

### gpu.fluidos.eu/interconnect_bandwidth

**Description**: GPU interconnect bandwidth  
**Source**: Applied by provider  
**Format**: Integer string  
**Examples**: `"600"` (NVLink), `"64"` (PCIe), `"900"` (NVLink 4.0)

### gpu.fluidos.eu/topology

**Description**: Multi-GPU topology configuration  
**Source**: Applied by provider  
**Values**: `"all-to-all"`, `"nvswitch"`, `"ring"`, `"mesh"`  
**API Mapping**: Used for topology-based optimization

### gpu.fluidos.eu/multi_gpu_efficiency

**Description**: Multi-GPU communication efficiency score (0.0-1.0)  
**Source**: Applied by provider  
**Format**: Decimal string  
**Examples**: `"0.95"`, `"0.85"`, `"0.70"`

## Provider Annotations (Optional)

### provider.fluidos.eu/name

**Description**: Provider identifier  
**Examples**: `"aws"`, `"gcp"`, `"provider-1"`

### provider.fluidos.eu/preemptible

**Description**: Spot/preemptible instance support  
**Values**: `"true"`, `"false"`

## Annotation Examples

### Minimal Configuration (Administrator)

```bash
# Only required manual annotations
kubectl annotate node gpu-worker-1 \
  location.fluidos.eu/region="eu-west-1" \
  cost.fluidos.eu/hourly-rate="0.1" \
  cost.fluidos.eu/currency="EUR"
```

### Complete Auto-Generated Result

```yaml
# Must be manually annotated by administrators using FLARE annotation specifications
# These are stored as annotations on the node
metadata:
  annotations:
    gpu.fluidos.eu/vendor: "nvidia"
    gpu.fluidos.eu/model: "nvidia-a100"
    gpu.fluidos.eu/count: "4"
    gpu.fluidos.eu/memory: "40Gi"
    gpu.fluidos.eu/tier: "premium"
    gpu.fluidos.eu/architecture: "ampere"
    gpu.fluidos.eu/interconnect: "nvlink"
    gpu.fluidos.eu/fp32_tflops: "19.5"
    gpu.fluidos.eu/shared: "true"
    gpu.fluidos.eu/sharing_strategy: "mig"
    workload.fluidos.eu/training-score: "0.95"
    workload.fluidos.eu/inference-score: "0.90"
    workload.fluidos.eu/hpc-score: "0.95"
    location.fluidos.eu/region: "eu-west-1"
    cost.fluidos.eu/hourly-rate: "0.1"
    cost.fluidos.eu/currency: "EUR"
```

## Querying Nodes with Annotations

Since annotations cannot be used directly in label selectors, FLARE implements the following approach:

### 1. Minimal Label for Selection

Nodes with GPUs should have a simple label for basic selection:

```bash
kubectl label node gpu-worker-1 node.fluidos.eu/gpu="true"
```

### 2. Query Annotations Programmatically

```bash
# Find all GPU nodes
kubectl get nodes -l node.fluidos.eu/gpu=true

# Get specific annotation values
kubectl get node gpu-worker-1 -o jsonpath='{.metadata.annotations.gpu\.fluidos\.eu/model}'

# List all GPU annotations for a node
kubectl get node gpu-worker-1 -o json | jq '.metadata.annotations | with_entries(select(.key | startswith("gpu.fluidos.eu/")))'
```

### 3. FLARE Controller Logic

The FLARE controllers query annotations directly when creating Flavors:

```go
// Pseudo-code for annotation reading
annotations := node.GetAnnotations()
gpuVendor := annotations["gpu.fluidos.eu/vendor"]
gpuModel := annotations["gpu.fluidos.eu/model"]
// ... process GPU metadata from annotations
```

## Migration from Labels to Annotations

For existing deployments using labels, use this migration script:

```bash
#!/bin/bash
# Migrate FLARE labels to annotations for a node
NODE=$1

# Get all gpu.fluidos.eu labels and convert to annotations
for label in $(kubectl get node $NODE -o json | jq -r '.metadata.labels | to_entries[] | select(.key | startswith("gpu.fluidos.eu/")) | .key'); do
  value=$(kubectl get node $NODE -o jsonpath="{.metadata.labels.$label}")
  annotation_key=$(echo $label | sed 's/\./\\./g')
  kubectl annotate node $NODE "$annotation_key=$value"
  kubectl label node $NODE "$label-"  # Remove the label
done

# Add the GPU selector label
kubectl label node $NODE node.fluidos.eu/gpu="true"
```

## FLARE API Mapping

| FLARE API Field | Maps to Annotation | Notes |
|-----------------|-------------------|-------|
| `resources.gpu.model` | `gpu.fluidos.eu/model` | Direct mapping |
| `resources.gpu.count` | `gpu.fluidos.eu/count` | Direct mapping |
| `resources.gpu.memory` | `gpu.fluidos.eu/memory` | Direct mapping |
| `resources.gpu.cores` | `gpu.fluidos.eu/cores` | Direct mapping |
| `resources.gpu.clock_speed` | `gpu.fluidos.eu/clock_speed` | API: G format (Hz) |
| `resources.gpu.compute_capability` | `gpu.fluidos.eu/compute_capability` | Direct mapping |
| `resources.gpu.architecture` | `gpu.fluidos.eu/architecture` | Direct mapping |
| `resources.gpu.tier` | `gpu.fluidos.eu/tier` | Direct mapping |
| `resources.gpu.shared` | `gpu.fluidos.eu/shared` | API: false → Annotation: "false" |
| `resources.gpu.interconnect` | `gpu.fluidos.eu/interconnect` | Direct mapping |
| `resources.gpu.interruptible` | `gpu.fluidos.eu/interruptible` | Direct mapping |
| `resources.gpu.multi_instance` | `gpu.fluidos.eu/sharing_strategy` | API: true → Annotation: "mig" |
| `resources.gpu.dedicated` | `gpu.fluidos.eu/dedicated` | Direct mapping |
| `constraints.location` | `location.fluidos.eu/region` | Direct mapping |
| `constraints.max_hourly_rate` | `cost.fluidos.eu/hourly-rate` | Cost filtering |
| `constraints.max_latency_ms` | `network.fluidos.eu/latency-ms` | Performance constraint |
| `constraints.preemptible` | `provider.fluidos.eu/preemptible` | Direct mapping |

## Supported GPU Vendors

| Vendor | Reference Document | Example Models |
|--------|-------------------|----------------|
| NVIDIA | `NVIDIA_GPU_Annotations_Mapping.md` | H100, A100, RTX series |
| AMD | `AMD_GPU_Annotations_Mapping.md` | MI300X, MI250, RX series |
| Intel | Future mapping document | Max series, Arc series |
| Custom | Use manual annotation | Any GPU model |

## Validation Rules

1. **Vendor**: Must be one of: nvidia, amd, intel, custom
2. **Model**: Must be vendor-prefixed (e.g., "nvidia-h100", "amd-mi300x")
3. **Memory**: Must use Gi suffix and be valid Kubernetes quantities
4. **Count**: Must be positive integer
5. **Scores**: Must be decimal values between 0.0 and 1.0
6. **Boolean**: Must be "true" or "false" (lowercase)
7. **Currency**: Currently limited to "EUR" (implementation constraint, multi-currency support planned)
8. **Region**: Must be lowercase with hyphens only

