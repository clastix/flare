# AMD GPU Operator Labels to FLARE Annotations Mapping

## Table of Contents

1. [Overview](#overview)
2. [AMD GPU Operator Label Reference](#amd-gpu-operator-label-reference)
3. [Direct Label to Annotation Mappings](#direct-label-to-annotation-mappings)
4. [Computed Annotation Mappings](#computed-annotation-mappings)
5. [Complete Auto-Generation Example](#complete-auto-generation-example)
6. [AMD-Specific Extensions](#amd-specific-extensions)
7. [Node Selection](#node-selection)

## Overview

This document provides the mapping between AMD GPU operator labels and FLARE's vendor-agnostic generic annotations. This mapping enables automatic annotation generation for AMD-based clusters while maintaining compatibility with other GPU vendors.

**Key Points:**
- **Source**: AMD GPU operator uses labels (no change)
- **Target**: FLARE stores GPU metadata as annotations (not labels)
- **Purpose**: Convert AMD labels to FLARE annotations for better separation of concerns

**Important**: This mapping is specific to AMD GPU operator integration. For other GPU vendors (NVIDIA, Intel), separate mapping documents should be created following the same pattern.

**Quick Reference**: See [FLARE GPU Annotations Reference - Quick Reference Table](FLARE_GPU_Annotations_Reference.md#quick-reference-table) for a complete list of all FLARE annotations with units and sample values.

## AMD GPU Operator Label Reference

### Complete Label Set (25+ labels from real MI300X node)

**Note**: The following labels are provided as informative samples for reference purposes and are not authoritative. Actual GPU operator labels may vary by version and configuration. Please consult the official AMD GPU Operator or ROCm documentation for the current and complete set of labels available in your deployment.

```yaml
# ROCm Driver Information
amd.com/rocm.driver-version.full: "6.8.5"
amd.com/rocm.driver-version.major: "6"
amd.com/rocm.driver-version.minor: "8"
amd.com/rocm.driver-version.revision: "5"
amd.com/rocm.driver.major: "6"
amd.com/rocm.driver.minor: "8"
amd.com/rocm.driver.rev: "5"

# ROCm Runtime Information  
amd.com/rocm.runtime-version.full: "6.0"
amd.com/rocm.runtime-version.major: "6"
amd.com/rocm.runtime-version.minor: "0"
amd.com/rocm.runtime.major: "6"
amd.com/rocm.runtime.minor: "0"

# GPU Hardware Properties
amd.com/gpu.compute.major: "9"
amd.com/gpu.compute.minor: "4"
amd.com/gpu.count: "8"
amd.com/gpu.family: "cdna3"
amd.com/gpu.machine: "bare-metal"
amd.com/gpu.memory: "192000"  # Total memory in MB
amd.com/gpu.present: "true"
amd.com/gpu.product: "AMD-Instinct-MI300X"
amd.com/gpu.replicas: "1"

# GPU Capabilities
amd.com/gpu.sharing-strategy: "none"
amd.com/sriov.capable: "true"
amd.com/sriov.strategy: "single"
amd.com/mps.capable: "false"

# Component Deployment Status (informational only)
amd.com/gpu.deploy.rocm-toolkit: "true"
amd.com/gpu.deploy.amdgpu-exporter: "true"
amd.com/gpu.deploy.device-plugin: "true"
amd.com/gpu.deploy.driver: "true"
amd.com/gpu.deploy.gpu-feature-discovery: "true"
amd.com/gpu.deploy.node-status-exporter: "true"
amd.com/gpu.deploy.operator-validator: "true"

# Additional metadata
amd.com/gfd.timestamp: "1749115477"
amd.com/gpu-driver-upgrade-state: "upgrade-done"
```

## Direct Label to Annotation Mappings

### Core GPU Information

| AMD GPU Operator Label | FLARE Generic Annotation | Conversion Logic |
|------------------------|--------------------------|------------------|
| `amd.com/gpu.count` | `gpu.fluidos.eu/count` | Direct copy |
| `amd.com/gpu.family` | `gpu.fluidos.eu/architecture` | Direct copy (cdna3→cdna3) |
| N/A | `gpu.fluidos.eu/vendor` | Static value: `"amd"` |

### GPU Memory Calculation

| AMD GPU Operator Label | FLARE Generic Annotation | Conversion Logic |
|------------------------|--------------------------|------------------|
| `amd.com/gpu.memory` | `gpu.fluidos.eu/memory` | Convert MB to Gi |

### GPU Sharing Capabilities

| AMD GPU Operator Label | FLARE Generic Annotation | Conversion Logic |
|------------------------|--------------------------|------------------|
| `amd.com/sriov.capable` | `gpu.fluidos.eu/shared` | Convert: true/false |
| `amd.com/sriov.strategy` + `amd.com/mps.capable` | `gpu.fluidos.eu/sharing-strategy` | Logic below |

## Computed Annotation Mappings

### GPU Model Normalization

**Input:** `amd.com/gpu.product`  
**Output:** `gpu.fluidos.eu/model`

**Examples:**

| amd.com/gpu.product | gpu.fluidos.eu/model |
|---------------------|----------------------|
| `AMD-Instinct-MI300X` | `amd-mi300x` |
| `AMD-Instinct-MI250X` | `amd-mi250x` |
| `AMD-Instinct-MI250` | `amd-mi250` |
| `AMD-Radeon-Pro-W7900` | `amd-w7900` |
| `AMD-Radeon-RX-7900-XTX` | `amd-rx7900xtx` |

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

**Input (from AMD GPU operator):**

```yaml
amd.com/gpu.count: "8"
amd.com/gpu.product: "AMD-Instinct-MI300X"
amd.com/gpu.memory: "192000"
amd.com/gpu.family: "cdna3"
amd.com/gpu.compute.major: "9"
amd.com/gpu.compute.minor: "4"
amd.com/sriov.capable: "true"
amd.com/mps.capable: "false"
amd.com/gpu.sharing-strategy: "none"
amd.com/gpu.machine: "bare-metal"
```

**Auto-generated FLARE generic annotations:**

```yaml
# Core GPU annotations (required)
gpu.fluidos.eu/vendor: "amd"                       # Static for AMD
gpu.fluidos.eu/model: "amd-mi300x"                 # Normalized from product
gpu.fluidos.eu/count: "8"                          # Direct copy
gpu.fluidos.eu/memory: "192Gi"                   # 196608 MB → 192Gi
gpu.fluidos.eu/tier: "premium"                     # Inferred from model

# Technical specifications
gpu.fluidos.eu/architecture: "cdna3"               # Direct copy from family
gpu.fluidos.eu/cores: "304"                      # Inferred from model (compute units)
gpu.fluidos.eu/interconnect: "infinity-fabric"     # Inferred (MI300X multi-GPU)
gpu.fluidos.eu/fp32_tflops: "163.4"               # Inferred from model

# Sharing capabilities
gpu.fluidos.eu/shared: "true"                      # From sriov.capable
gpu.fluidos.eu/sharing-strategy: "sriov"           # From sriov.capable=true

# Workload optimization scores
workload.fluidos.eu/training-score: "0.96"         # Inferred from model
workload.fluidos.eu/inference-score: "0.94"        # Inferred from model
workload.fluidos.eu/hpc-score: "0.98"              # Inferred from model
workload.fluidos.eu/graphics-score: "0.40"         # Inferred from model
```

## AMD-Specific Extensions

These annotations are specific to AMD and don't have generic equivalents:

```yaml
# AMD-specific technical details (optional, for debugging/monitoring)
amd.fluidos.eu/rocm-version: "6.0"                        # From runtime-version.full
amd.fluidos.eu/driver-version: "6.8.5"                    # From driver-version.full
amd.fluidos.eu/compute-capability: "9.4"                  # From compute.major + minor
amd.fluidos.eu/sriov-capable: "true"                      # From sriov.capable
amd.fluidos.eu/gfx-architecture: "gfx942"                 # AMD-specific architecture
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