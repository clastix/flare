# NVIDIA GPU Operator Labels to FLARE Annotations Mapping

## Table of Contents

1. [Overview](#overview)
2. [NVIDIA GPU Operator Label Reference](#nvidia-gpu-operator-label-reference)
3. [Direct Label to Annotation Mappings](#direct-label-to-annotation-mappings)
4. [Computed Annotation Mappings](#computed-annotation-mappings)
5. [Complete Auto-Generation Example](#complete-auto-generation-example)
6. [NVIDIA-Specific Extensions](#nvidia-specific-extensions)
7. [Node Selection](#node-selection)

## Overview

This document provides the mapping between NVIDIA GPU operator labels and FLARE's vendor-agnostic generic annotations. This mapping enables automatic annotation generation for NVIDIA-based clusters while maintaining compatibility with other GPU vendors.

**Key Points:**

- **Source**: NVIDIA GPU operator uses labels (no change)
- **Target**: FLARE stores GPU metadata as annotations (not labels)
- **Purpose**: Convert NVIDIA labels to FLARE annotations for better separation of concerns

**Important**: This mapping is specific to NVIDIA GPU operator integration. For other GPU vendors (AMD, Intel), separate mapping documents should be created following the same pattern.

**Quick Reference**: See [FLARE GPU Annotations Reference - Quick Reference Table](FLARE_GPU_Annotations_Reference.md#quick-reference-table) for a complete list of all FLARE annotations with units and sample values.

## NVIDIA GPU Operator Label Reference

### Complete Label Set (25+ labels from real RTX A6000 node)

**Note**: The following labels are provided as informative samples for reference purposes and are not authoritative. Actual GPU operator labels may vary by version and configuration. Please consult the official NVIDIA GPU Operator documentation for the current and complete set of labels available in your deployment.

```yaml
# CUDA Driver Information
nvidia.com/cuda.driver-version.full: "550.54.15"
nvidia.com/cuda.driver-version.major: "550"
nvidia.com/cuda.driver-version.minor: "54"
nvidia.com/cuda.driver-version.revision: "15"
nvidia.com/cuda.driver.major: "550"
nvidia.com/cuda.driver.minor: "54"
nvidia.com/cuda.driver.rev: "15"

# CUDA Runtime Information  
nvidia.com/cuda.runtime-version.full: "12.4"
nvidia.com/cuda.runtime-version.major: "12"
nvidia.com/cuda.runtime-version.minor: "4"
nvidia.com/cuda.runtime.major: "12"
nvidia.com/cuda.runtime.minor: "4"

# GPU Hardware Properties
nvidia.com/gpu.compute.major: "8"
nvidia.com/gpu.compute.minor: "6"
nvidia.com/gpu.count: "2"
nvidia.com/gpu.family: "ampere"
nvidia.com/gpu.machine: "KVM"
nvidia.com/gpu.memory: "49140"  # Total memory in MiB
nvidia.com/gpu.present: "true"
nvidia.com/gpu.product: "NVIDIA-RTX-A6000"
nvidia.com/gpu.replicas: "1"

# GPU Capabilities
nvidia.com/gpu.sharing-strategy: "none"
nvidia.com/mig.capable: "false"
nvidia.com/mig.strategy: "mixed"
nvidia.com/mps.capable: "false"

# Component Deployment Status (informational only)
nvidia.com/gpu.deploy.container-toolkit: "true"
nvidia.com/gpu.deploy.dcgm: "true"
nvidia.com/gpu.deploy.dcgm-exporter: "true"
nvidia.com/gpu.deploy.device-plugin: "true"
nvidia.com/gpu.deploy.driver: "true"
nvidia.com/gpu.deploy.gpu-feature-discovery: "true"
nvidia.com/gpu.deploy.node-status-exporter: "true"
nvidia.com/gpu.deploy.operator-validator: "true"

# Additional metadata
nvidia.com/gfd.timestamp: "1749115477"
nvidia.com/gpu-driver-upgrade-state: "upgrade-done"
```

## Direct Label to Annotation Mappings

### Core GPU Information

| NVIDIA GPU Operator Label | FLARE Generic Annotation | Conversion Logic |
|----------------------------|--------------------------|------------------|
| `nvidia.com/gpu.count` | `gpu.fluidos.eu/count` | Direct copy |
| `nvidia.com/gpu.family` | `gpu.fluidos.eu/architecture` | Direct copy (ampere→ampere) |
| N/A | `gpu.fluidos.eu/vendor` | Static value: `"nvidia"` |

### GPU Memory Calculation

| NVIDIA GPU Operator Label | FLARE Generic Annotation | Conversion Logic |
|----------------------------|--------------------------|------------------|
| `nvidia.com/gpu.memory` | `gpu.fluidos.eu/memory` | Convert MiB to Gi |

### GPU Sharing Capabilities

| NVIDIA GPU Operator Label | FLARE Generic Annotation | Conversion Logic |
|----------------------------|--------------------------|------------------|
| `nvidia.com/mig.capable` | `gpu.fluidos.eu/multi_instance` | Direct copy |
| `nvidia.com/mig.strategy` + `nvidia.com/mps.capable` | `gpu.fluidos.eu/sharing-strategy` | Logic below |


## Computed Annotation Mappings

### GPU Model Normalization

**Input:** `nvidia.com/gpu.product`  
**Output:** `gpu.fluidos.eu/model`

**Examples:**

| nvidia.com/gpu.product | gpu.fluidos.eu/model |
|------------------------|----------------------|
| `NVIDIA-RTX-A6000` | `nvidia-rtx-a6000` |
| `NVIDIA-A100-SXM4-40Gi` | `nvidia-a100` |
| `NVIDIA-H100-PCIe-80Gi` | `nvidia-h100` |
| `NVIDIA-RTX-4090` | `nvidia-rtx-4090` |
| `NVIDIA-T4` | `nvidia-t4` |

### GPU Tier Classification

**Input:** `gpu.fluidos.eu/model` (computed)  
**Output:** `gpu.fluidos.eu/tier`

### Performance Characteristics

**Input:** `gpu.fluidos.eu/model` (computed)  
**Output:** Multiple workload score annotations

### GPU Interconnect Inference

**Input:** `gpu.fluidos.eu/model` + `gpu.fluidos.eu/count`  
**Output:** `gpu.fluidos.eu/interconnect`

## Complete Auto-Generation Example

**Input (from NVIDIA GPU operator):**

```yaml
nvidia.com/gpu.count: "2"
nvidia.com/gpu.product: "NVIDIA-RTX-A6000"
nvidia.com/gpu.memory: "49140"
nvidia.com/gpu.family: "ampere"
nvidia.com/gpu.compute.major: "8"
nvidia.com/gpu.compute.minor: "6"
nvidia.com/mig.capable: "false"
nvidia.com/mps.capable: "false"
nvidia.com/gpu.sharing-strategy: "none"
nvidia.com/gpu.machine: "KVM"
```

**Auto-generated FLARE generic annotations:**

```yaml
# Core GPU annotations (required)
gpu.fluidos.eu/vendor: "nvidia"                    # Static for NVIDIA
gpu.fluidos.eu/model: "nvidia-rtx-a6000"           # Normalized from product
gpu.fluidos.eu/count: "2"                          # Direct copy
gpu.fluidos.eu/memory: "48Gi"              # 49140 MiB → 48Gi
gpu.fluidos.eu/tier: "standard"                    # Inferred from model

# Technical specifications
gpu.fluidos.eu/architecture: "ampere"              # Direct copy from family
gpu.fluidos.eu/cores: "10752"              # Inferred from model (CUDA cores)
gpu.fluidos.eu/interconnect: "pcie"                # Inferred (RTX A6000 typically PCIe)
gpu.fluidos.eu/fp32_tflops: "38.7"                 # Inferred from model
gpu.fluidos.eu/compute_capability: "8.6"               # From compute.major + minor

# Sharing capabilities
gpu.fluidos.eu/multi_instance: "false"            # From mig.capable
gpu.fluidos.eu/sharing_strategy: "none"            # From sharing-strategy

# Workload optimization scores
gpu.fluidos.eu/training_score: "0.85"         # Inferred from model
gpu.fluidos.eu/inference_score: "0.90"        # Inferred from model
gpu.fluidos.eu/hpc_score: "0.80"              # Inferred from model
gpu.fluidos.eu/graphics_score: "0.95"         # Inferred from model
```

## NVIDIA-Specific Extensions

These annotations are specific to NVIDIA and don't have generic equivalents:

```yaml
# NVIDIA-specific technical details (optional, for debugging/monitoring)
nvidia.fluidos.eu/cuda_version: "12.4"                    # From runtime-version.full
nvidia.fluidos.eu/driver_version: "550.54.15"             # From driver-version.full
nvidia.fluidos.eu/compute_capability: "8.6"               # From compute.major + minor
nvidia.fluidos.eu/mig_capable: "false"                    # From mig.capable
nvidia.fluidos.eu/mps_capable: "false"                    # From mps.capable
```

## Node Selection

While GPU metadata is stored as annotations, a simple label is added for node selection:

```bash
# Add GPU selector label for basic node filtering
kubectl label node gpu-worker-1 node.fluidos.eu/gpu="true"
```

This allows FLARE controllers to:

1. Find GPU nodes using label selectors
2. Read detailed GPU specifications from annotations
3. Create appropriate Flavors based on annotation metadata
