# FLARE Sample Use Cases

This document provides comprehensive use case scenarios demonstrating FLARE's GPU federation capabilities across different application domains. Each scenario follows the complete automated end-to-end workflow from user intent submission to workload deployment.

## Table of Contents

1. [AI Inference Service](#1-ai-inference-service)
2. [High-Performance AI Training](#2-high-performance-ai-training) 
3. [LLM Fine-Tuning](#3-llm-fine-tuning)
4. [High-Performance Computing](#4-high-performance-computing)
5. [Real-Time Video Analytics](#5-real-time-video-analytics)
6. [Edge Inference](#6-edge-inference)
7. [Batch Processing](#7-batch-processing)
8. [Multi-Tenant Resources](#8-multi-tenant-resources)
9. [Distributed Workloads](#9-distributed-workloads)


## 1. AI Inference Service

### Scenario Overview

**Business Need**: Deploy a cost-effective AI inference service for LLM model serving  
**User**: Startup founder with no Kubernetes expertise  
**Optimization**: Minimize costs while meeting performance requirements

### Prerequisites

- FLARE platform deployed with FLUIDOS GPU enhancements
- Multiple providers with GPU resources in the federation
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

User submits natural language intent to FLARE:

```bash
# User submits intent via FLARE API
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Cost_Minimization",
      "workload": {
        "type": "service",
        "name": "llama2-inference",
        "image": "ghcr.io/huggingface/text-generation-inference:2.0.2",
        "commands": [
          "--model-id=meta-llama/Llama-2-7b-chat-hf",
          "--max-batch-size=16"
        ],
        "ports": [
          {
            "port": 8080,
            "expose": true
          }
        ],
        "resources": {
          "cpu": "8",
          "memory": "32Gi",
          "gpu": {
            "model": "nvidia-t4",
            "count": 1,
            "memory": "16Gi"
          }
        }
      },
      "constraints": {
        "max_hourly_cost": "5 EUR",
        "location": "EU"
      }
    }
  }'
```
Response

```json
{
  "intent_id": "intent-llama2-2024-01-15-001",
  "status": "pending",
  "estimated_cost": "0.45-0.65 EUR/hour",
  "eta_seconds": 300
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-llama2-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "16Gi"  # LLaMA-2 7B minimum
      - field: count
        filter: NumberRangeFilter
        data:
          min: 1
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 5.0  # EUR per hour from user budget
  # Full automation enabled
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters GPU resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **GPU Flavor discovery across clusters** - Scans all federated providers
2. **Native GPU filtering** - Applies GPU memory and cost filters
3. **PeeringCandidate creation** - Only for matching resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: RTX 4080 (16Gi) - €0.45/hour - Germany ✓ (matches filters)
- Provider-2: RTX 4090 (24Gi) - €0.65/hour - Netherlands ✓ (matches filters)  
- Provider-3: A100 (40Gi) - €3.50/hour - France ✓ (matches filters)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: RTX 4080 (€0.45/hour) is selected
- No performance evaluation - just availability

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource
2. Liqo establishes secure tunnel to provider
3. Virtual node appears in consumer cluster

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the workload:

**What FLARE does**:

1. Creates namespace with offloading configuration
2. Generates optimized deployment for LLaMA-2
3. Configures service exposure and storage
4. Applies all resources

**Status Update to User**:

```json
{
  "intent_id": "intent-llama2-2024-01-15-001",
  "status": "completed",
  "endpoint": "https://llama2-api.flare.example.com",
  "actual_cost": "0.45 EUR/hour",
  "gpu": "RTX 4080 (16Gi)",
  "location": "Germany",
  "deployment_time": "4m 32s"
}
```

## 2. High-Performance AI Training

### Scenario Overview

**Business Need**: Train BERT-large model with distributed training for maximum performance  
**User**: Machine learning researcher with limited infrastructure knowledge  
**Optimization**: Performance maximization, shortest training time

### Prerequisites

- FLARE platform deployed with FLUIDOS GPU enhancements
- Multiple providers with multi-GPU resources in the federation
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

User submits high-level training intent to FLARE:

```bash
# User submits training intent via FLARE API
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Performance_Maximization",
      "workload": {
        "type": "job",
        "name": "bert-large-training",
        "image": "pytorch/pytorch:2.0.1-cuda11.7-cudnn8-devel",
        "commands": [
          "torchrun", "--nproc_per_node=4", "--nnodes=2",
          "train_bert.py", "--model=bert-large-uncased",
          "--epochs=10", "--batch_size=32"
        ],
        "resources": {
          "cpu": "64",
          "memory": "256Gi",
          "gpu": {
            "model": "nvidia-rtx-4090",
            "count": 8,
            "memory": "24Gi",
            "interconnect": "nvlink"
          }
        },
        "storage": {
          "volumes": [
            {
              "name": "training-data",
              "size": "500Gi",
              "type": "persistent",
              "path": "/data"
            }
          ]
        }
      },
      "constraints": {
        "deadline": "2024-12-15T08:00:00Z",
        "location": "EU"
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-ai-training-2024-01-15-001",
  "status": "pending",
  "estimated_completion_time": "3.5 hours",
  "eta_seconds": 600
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-ai-training-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: interconnect
        filter: StringFilter
        data:
          value: "nvlink"  # Fast interconnect
      - field: count
        filter: NumberRangeFilter
        data:
          min: 8
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "24Gi"
      - field: interconnect
        filter: StringFilter
        data:
          value: "nvlink"
  # Full automation enabled
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters multi-GPU resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **Multi-GPU Flavor discovery** - Scans all federated providers for high-performance configurations
2. **Native interconnect filtering** - Applies NVLink requirements for optimal distributed training
3. **Performance-aware filtering** - Considers GPU count, memory, and communication bandwidth
4. **PeeringCandidate creation** - Only for matching multi-GPU resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: H100 8x80Gi NVLink - €75/hour - Germany ✓ (matches filters)
- Provider-2: A100 8x40Gi NVLink - €25/hour - Netherlands ✓ (matches filters)  
- Provider-3: RTX 4090 8x24Gi NVLink - €4.50/hour - France ✓ (matches filters)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: RTX 4090 8x24Gi (€4.50/hour) is selected
- Performance optimization prioritizes multi-GPU configurations

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource for multi-GPU training configuration
2. Liqo establishes secure tunnel to provider
3. Virtual node appears in consumer cluster with training capabilities

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the distributed training workload:

**What FLARE does**:

1. Creates namespace with offloading configuration
2. Generates optimized Job for BERT distributed training
3. Configures storage and environment variables
4. Applies all resources with training-specific settings

**Status Update to User**:

```json
{
  "intent_id": "intent-ai-training-2024-01-15-001",
  "status": "running",
  "progress": "45%",
  "estimated_completion": "2.1 hours remaining",
  "gpu": "RTX 4090 8x24Gi NVLink",
  "location": "France",
  "training_metrics": {
    "loss": 0.24,
    "throughput": "1200 samples/sec"
  }
}
```

## 3. LLM Fine-Tuning

### Scenario Overview

**Business Need**: Fine-tune LLaMA-7B model efficiently using parameter-efficient methods  
**User**: AI researcher with limited GPU optimization expertise  
**Optimization**: Memory efficiency and cost-effectiveness

### Prerequisites

- FLARE platform deployed with FLUIDOS GPU enhancements
- Multiple providers with memory-optimized GPU resources in the federation
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

User submits memory-efficient fine-tuning intent to FLARE:

```bash
# User submits fine-tuning intent via FLARE API
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Cost_Minimization",
      "workload": {
        "type": "job",
        "name": "llama7b-lora-finetuning",
        "image": "huggingface/transformers-pytorch-gpu:4.22.0",
        "commands": [
          "python", "finetune_lora.py",
          "--model_name=meta-llama/Llama-2-7b-hf",
          "--dataset=/data/custom-dataset",
          "--method=lora", "--lora_rank=16",
          "--epochs=3", "--batch_size=4"
        ],
        "resources": {
          "cpu": "16",
          "memory": "64Gi",
          "gpu": {
            "model": "nvidia-rtx-4080",
            "count": 1,
            "memory": "12Gi"
          }
        },
        "storage": {
          "volumes": [
            {
              "name": "training-data",
              "size": "100Gi",
              "type": "persistent",
              "path": "/data"
            }
          ]
        }
      },
      "constraints": {
        "max_hourly_cost": "8 EUR",
        "deadline": "2024-12-15T14:00:00Z"
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-finetuning-2024-01-15-001",
  "status": "pending",
  "estimated_completion_time": "5.2 hours",
  "eta_seconds": 300
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-finetuning-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: count
        filter: NumberRangeFilter
        data:
          min: 1
          max: 1
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "12Gi"
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 8.0  # EUR per hour from user budget
  # Full automation enabled
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters GPU resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **Memory-optimized Flavor discovery** - Scans all federated providers for fine-tuning configurations
2. **Native GPU filtering** - Applies memory requirements for LoRA fine-tuning
3. **Cost optimization filtering** - Considers budget constraints for extended training
4. **PeeringCandidate creation** - Only for cost-effective memory resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: RTX 4090 (24Gi) - €0.60/hour - Germany ✓ (matches filters)
- Provider-2: RTX 4080 Mobile (12Gi) - €0.14/hour - Netherlands ✓ (matches filters)  
- Provider-3: A100 (40Gi) - €3.30/hour - France ✓ (matches filters)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: RTX 4080 Mobile (€0.14/hour) is selected
- Cost optimization prioritizes most affordable option

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource for memory-optimized configuration
2. Liqo establishes secure tunnel to provider
3. Virtual node appears in consumer cluster with fine-tuning capabilities

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the fine-tuning workload:

**What FLARE does**:

1. Creates namespace with offloading configuration
2. Generates optimized Job for LoRA fine-tuning
3. Configures storage and model artifacts
4. Applies all resources with memory-efficient settings

**Status Update to User**:

```json
{
  "intent_id": "intent-finetuning-2024-01-15-001",
  "status": "running",
  "progress": "60%",
  "estimated_completion": "2.1 hours remaining",
  "gpu": "RTX 4080 Mobile (12Gi)",
  "location": "Netherlands",
  "training_metrics": {
    "memory_usage": "9.8Gi/12Gi",
    "lora_efficiency": "95.3%"
  }
}
```

## 4. High-Performance Computing

### Scenario Overview

**Business Need**: Run large-scale scientific simulation requiring massive computational power  
**User**: Research scientist with complex computational requirements  
**Optimization**: Compute performance maximization for scientific workloads

### Prerequisites

- FLARE platform deployed with FLUIDOS GPU enhancements
- Multiple providers with HPC-optimized GPU clusters in the federation
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

User submits high-performance computing intent to FLARE:

```bash
# User submits HPC intent via FLARE API
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Performance_Maximization",
      "workload": {
        "type": "job",
        "name": "molecular-dynamics-simulation",
        "image": "gromacs/gromacs:2023.3-cuda",
        "commands": [
          "gmx_mpi", "mdrun", "-v", "-deffnm", "simulation",
          "-ntomp", "8", "-nb", "gpu", "-pme", "gpu",
          "-npme", "1", "-nsteps", "500000"
        ],
        "resources": {
          "cpu": "64",
          "memory": "256Gi",
          "gpu": {
            "model": "nvidia-a100",
            "count": 8,
            "memory": "40Gi",
            "interconnect": "nvlink"
          }
        }
      },
      "constraints": {
        "max_hourly_cost": "100 EUR",
        "deadline": "2024-12-15T20:00:00Z"
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-hpc-2024-01-15-001",
  "status": "pending",
  "estimated_completion_time": "12 hours",
  "eta_seconds": 450
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-hpc-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: count
        filter: NumberRangeFilter
        data:
          min: 8
          max: 8
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "40Gi"
      - field: interconnect
        filter: StringFilter
        data:
          value: "nvlink"
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 100.0  # EUR per hour from user budget
  # Full automation enabled
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters HPC GPU resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **HPC Flavor discovery across clusters** - Scans all federated providers for high-performance configurations
2. **Native multi-GPU filtering** - Applies 8-GPU count and NVLink requirements
3. **Performance optimization filtering** - Considers interconnect and memory requirements for HPC workloads
4. **PeeringCandidate creation** - Only for matching HPC resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: A100 8x40Gi NVLink - €85/hour - Germany ✓ (matches filters)
- Provider-2: H100 8x80Gi NVLink - €120/hour - Netherlands ✗ (exceeds budget)
- Provider-3: A100 4x40Gi NVLink - €45/hour - France ✗ (insufficient GPU count)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: A100 8x40Gi (€85/hour) is selected
- Performance optimization prioritizes multi-GPU configurations

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource for 8-GPU HPC configuration
2. Liqo establishes secure tunnel to provider
3. Virtual node appears in consumer cluster with HPC capabilities

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the HPC workload:

**What FLARE does**:

1. Creates namespace with offloading configuration
2. Generates optimized Job for molecular dynamics simulation
3. Configures storage and environment variables
4. Applies all resources with HPC-specific settings

**Status Update to User**:

```json
{
  "intent_id": "intent-hpc-2024-01-15-001",
  "status": "running",
  "progress": "25%",
  "estimated_completion": "9 hours remaining",
  "gpu": "A100 8x40Gi NVLink",
  "location": "Germany", 
  "actual_cost": "85 EUR/hour",
  "deployment_time": "7m 15s",
  "simulation_metrics": {
    "timesteps_per_day": "2.3M",
    "efficiency": "87.2%",
    "gpu_utilization": "94%"
  }
}
```

## 5. Real-Time Video Analytics

### Scenario Overview

**Business Need**: Deploy real-time video analytics for traffic monitoring with ultra-low latency  
**User**: Smart city operator with limited edge computing expertise  
**Optimization**: Latency minimization

### Prerequisites

- FLARE platform deployed with FLUIDOS GPU enhancements
- Multiple edge providers with video processing GPU resources in the federation
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

User submits low-latency video analytics intent to FLARE:

```bash
# User submits video processing intent via FLARE API
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Latency_Minimization",
      "workload": {
        "type": "service",
        "name": "traffic-video-analytics",
        "image": "nvcr.io/nvidia/deepstream:6.3-devel",
        "commands": [
          "deepstream-app", "-c", "/config/traffic_analytics.txt"
        ],
        "ports": [
          {
            "port": 8080,
            "expose": true
          }
        ],
        "resources": {
          "cpu": "16",
          "memory": "64Gi",
          "gpu": {
            "model": "nvidia-rtx-4080",
            "count": 1,
            "memory": "12Gi",
            "tier": "gaming"
          }
        }
      },
      "constraints": {
        "max_latency_ms": 20,
        "location": "EU"
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-video-2024-01-15-001",
  "status": "pending",
  "estimated_latency": "12ms",
  "eta_seconds": 180
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-video-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: count
        filter: NumberRangeFilter
        data:
          min: 1
          max: 1
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "12Gi"
      - field: tier
        filter: StringFilter
        data:
          value: "gaming"  # Optimized for video processing
      locationFilter:
        name: StringFilter
        data:
          value: "EU"
  # Full automation enabled
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters edge GPU resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **Edge Flavor discovery across clusters** - Scans all federated providers for low-latency edge configurations
2. **Native video processing filtering** - Applies GPU memory and video decode capabilities
3. **Network latency filtering** - Considers latency requirements for real-time processing
4. **PeeringCandidate creation** - Only for edge-optimized resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: RTX 4080 (16Gi) - €0.45/hour - Frankfurt Edge (8ms) ✓ (matches filters)
- Provider-2: RTX 4080 (16Gi) - €0.75/hour - Amsterdam Edge (12ms) ✓ (matches filters)
- Provider-3: A100 (40Gi) - €3.50/hour - Germany Central (45ms) ✗ (exceeds latency)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: RTX 4080 (€0.45/hour, 8ms) is selected
- Latency optimization prioritizes lowest-latency options

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource for edge GPU configuration
2. Liqo establishes secure tunnel to edge provider
3. Virtual node appears in consumer cluster with edge capabilities

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the video analytics workload:

**What FLARE does**:

1. Creates namespace with offloading configuration
2. Generates optimized Service for real-time video analytics
3. Configures port exposure and networking
4. Applies all resources with low-latency settings

**Status Update to User**:

```json
{
  "intent_id": "intent-video-2024-01-15-001",
  "status": "completed",
  "endpoint": "https://video-analytics.edge.example.com",
  "actual_latency": "8ms",
  "gpu": "RTX 4080 (16Gi)",
  "location": "Frankfurt Edge",
  "actual_cost": "0.45 EUR/hour",
  "deployment_time": "3m 45s",
  "video_metrics": {
    "fps_processed": 30,
    "streams_active": 10,
    "detection_accuracy": "94.7%",
    "gpu_utilization": "78%"
  }
}
```

## 6. Edge Inference

### Scenario Overview

**Business Need**: Deploy lightweight AI inference at edge locations for IoT applications  
**User**: IoT platform developer requiring distributed edge deployment  
**Optimization**: Power efficiency

### Prerequisites

- FLARE platform deployed with edge provider support
- Multiple edge locations with inference-optimized GPU resources
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

```bash
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Cost_Minimization",
      "workload": {
        "type": "service",
        "name": "iot-edge-inference",
        "image": "tensorflow/tensorflow:2.13.0-gpu",
        "commands": [
          "python3 inference_server.py --model mobilenet --port 8080"
        ],
        "env": [
          "TF_CPP_MIN_LOG_LEVEL=2",
          "CUDA_VISIBLE_DEVICES=0"
        ],
        "ports": [
          {
            "port": 8080,
            "protocol": "TCP",
            "expose": true
          }
        ],
        "resources": {
          "cpu": "4",
          "memory": "8Gi",
          "gpu": {
            "model": "nvidia-t4",
            "count": 1,
            "memory": "4Gi",
            "tier": "inference"
          }
        }
      },
      "constraints": {
        "max_hourly_cost": "2 EUR",
        "location": "edge",
        "max_latency_ms": 50
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-edge-inference-2024-01-15-001",
  "status": "pending",
  "estimated_cost": "0.25 EUR/hour",
  "eta_seconds": 240
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-edge-inference-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: count
        filter: NumberRangeFilter
        data:
          min: 1
          max: 1
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "4Gi"
      - field: tier
        filter: StringFilter
        data:
          value: "inference"
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 2.0  # EUR per hour
  # Full automation enabled
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters edge inference resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **Edge inference Flavor discovery** - Scans all federated providers for inference-optimized configurations
2. **Native power efficiency filtering** - Applies inference tier and memory requirements
3. **Cost optimization filtering** - Considers budget constraints for edge deployment
4. **PeeringCandidate creation** - Only for edge-optimized inference resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: T4 (16Gi) - €0.25/hour - Edge Location A (25ms) ✓ (matches filters)
- Provider-2: T4 (16Gi) - €0.15/hour - Edge Location B (35ms) ✓ (matches filters)
- Provider-3: RTX 4090 (24Gi) - €0.85/hour - Cloud Center (15ms) ✗ (exceeds budget)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: T4 (€0.25/hour) is selected
- Energy efficiency optimization prioritizes inference-optimized GPUs

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource for edge inference configuration
2. Liqo establishes secure tunnel to edge provider
3. Virtual node appears in consumer cluster with inference capabilities

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the edge inference workload:

**What FLARE does**:

1. Creates namespace with offloading configuration
2. Generates optimized Service for IoT inference
3. Configures port exposure and networking
4. Applies all resources with energy-efficient settings

**Status Update to User**:

```json
{
  "intent_id": "intent-edge-inference-2024-01-15-001",
  "status": "completed",
  "endpoint": "https://iot-inference.edge.example.com",
  "actual_cost": "0.25 EUR/hour",
  "gpu": "T4 (16Gi)",
  "location": "Edge Location A",
  "deployment_time": "4m 10s",
  "inference_metrics": {
    "requests_per_second": 45,
    "average_latency": "25ms",
    "accuracy": "92.1%",
    "power_usage": "75W"
  }
}
```

## 7. Batch Processing

### Scenario Overview

**Business Need**: Process large datasets in batch mode with cost optimization  
**User**: Data scientist requiring scalable batch processing capabilities  
**Optimization**: Cost minimization with flexible scheduling

### Prerequisites

- FLARE platform deployed with batch job support
- Multiple providers with cost-effective GPU resources
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

```bash
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Cost_Minimization",
      "workload": {
        "type": "batch",
        "name": "data-processing-batch",
        "image": "apache/spark:3.4.0-gpu",
        "resources": {
          "cpu": "32",
          "memory": "128Gi",
          "gpu": {
            "model": "nvidia-rtx-4080",
            "count": 4,
            "memory": "16Gi"
          }
        },
        "batch": {
          "parallel_tasks": 100,
          "completion_policy": "All",
          "max_retries": 3
        }
      },
      "constraints": {
        "max_total_cost": "50 EUR",
        "preemptible": true,
        "deadline": "2024-12-20T00:00:00Z"
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-batch-2024-01-15-001",
  "status": "pending",
  "estimated_completion_time": "10 hours",
  "eta_seconds": 360
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-batch-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: count
        filter: NumberRangeFilter
        data:
          min: 4
          max: 4
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "16Gi"
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 12.5  # EUR per hour (50 EUR total / 4 hours)
  # Full automation enabled with preemptible pricing
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters batch processing resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **Batch processing Flavor discovery** - Scans all federated providers for cost-effective configurations
2. **Native cost filtering** - Applies budget constraints and preemptible pricing
3. **Multi-GPU filtering** - Considers parallel task requirements
4. **PeeringCandidate creation** - Only for batch-optimized resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: RTX 4080 (16Gi) - €0.35/hour - Germany ✓ (matches cost filters)
- Provider-2: RTX 4090 (24Gi) - €0.55/hour - Netherlands ✓ (matches cost filters)  
- Provider-3: A100 (40Gi) - €2.10/hour - France ✓ (preemptible pricing)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: RTX 4080 (€0.35/hour) is selected
- Cost optimization prioritizes most affordable option

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource for batch processing configuration
2. Liqo establishes secure tunnel to provider
3. Virtual node appears in consumer cluster with batch capabilities

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the batch processing workload:

**What FLARE does**:

1. Creates namespace with offloading configuration
2. Generates optimized Batch Job for data processing
3. Configures parallel task distribution and storage
4. Applies all resources with cost-optimized settings

**Status Update to User**:

```json
{
  "intent_id": "intent-batch-2024-01-15-001",
  "status": "running",
  "progress": "30%",
  "estimated_completion": "8 hours remaining",
  "gpu": "RTX 4080 (16Gi) 4x nodes",
  "location": "Germany",
  "batch_metrics": {
    "tasks_completed": 30,
    "tasks_remaining": 70,
    "cost_per_hour": "1.40 EUR",
    "total_estimated_cost": "11.20 EUR"
  }
}
```

## 8. Multi-Tenant Resources

### Scenario Overview

**Business Need**: Provide isolated GPU resources for multiple tenants with resource quotas  
**User**: Cloud platform operator requiring tenant isolation  
**Optimization**: Resource utilization with strict isolation

### Prerequisites

- FLARE platform deployed with multi-tenancy support
- Capsule or similar tenant management system
- Multiple providers supporting tenant isolation

### Automated Workflow Steps

#### 1. User Intent Submission

```bash
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $TENANT_A_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Balanced_Optimization",
      "workload": {
        "type": "service",
        "name": "tenant-a-ml-service",
        "image": "pytorch/pytorch:2.0.1-cuda11.7-cudnn8-runtime",
        "resources": {
          "cpu": "8",
          "memory": "32Gi",
          "gpu": {
            "model": "nvidia-rtx-4090",
            "count": 2,
            "memory": "16Gi"
          }
        }
      },
      "constraints": {
        "max_hourly_cost": "15 EUR"
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-multitenant-2024-01-15-001",
  "status": "pending",
  "estimated_cost": "1.20 EUR/hour",
  "eta_seconds": 300
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates a GPU-aware Solver:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-multitenant-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: count
        filter: NumberRangeFilter
        data:
          min: 2
          max: 2
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "16Gi"
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 15.0  # EUR per hour from user budget
  # Full automation enabled with tenant isolation
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters multi-tenant GPU resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **Multi-tenant Flavor discovery** - Scans all federated providers for tenant isolation capabilities
2. **Native isolation filtering** - Applies strict tenant separation requirements
3. **Resource quota filtering** - Considers tenant-specific resource limits
4. **PeeringCandidate creation** - Only for tenant-isolated resources

**Discovery Results** (handled by FLUIDOS):

- Provider-1: RTX 4090 (24Gi) 2x - €1.20/hour - Germany ✓ (supports tenant isolation)
- Provider-2: A100 (40Gi) 2x - €6.50/hour - Netherlands ✓ (secure tenant isolation)
- Provider-3: RTX 4080 Mobile (12Gi) 2x - €0.90/hour - France ✓ (cost-effective option)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically selects the first available PeeringCandidate and creates a Contract:

**FLUIDOS first-match selection**:

- Takes first available candidate from filtered list
- In this case: RTX 4090 2x (€1.20/hour) is selected
- Multi-tenant optimization prioritizes isolation capabilities

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish peering and create virtual node:

**What happens automatically**:

1. FLUIDOS creates Allocation resource for multi-tenant configuration
2. Liqo establishes secure tunnel to provider with isolation
3. Virtual node appears in consumer cluster with tenant separation

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys the multi-tenant workload:

**What FLARE does**:

1. Creates namespace with Capsule tenant isolation configuration
2. Generates optimized Service with tenant-specific resources
3. Configures RBAC and network policies for isolation
4. Applies all resources with strict tenant separation

**Status Update to User**:

```json
{
  "intent_id": "intent-multitenant-2024-01-15-001",
  "status": "completed",
  "endpoint": "https://tenant-a.flare.example.com",
  "gpu": "RTX 4090 (24Gi) 2x nodes",
  "location": "Germany",
  "tenant_metrics": {
    "isolation_level": "strict",
    "resource_quota": "2 GPUs, 16 CPU, 64Gi RAM",
    "cost_per_hour": "1.20 EUR",
    "tenant_id": "tenant-a-ml-service"
  }
}
```

## 9. Distributed Workloads

### Scenario Overview

**Business Need**: Deploy workloads across multiple geographic regions for global applications  
**User**: Global application developer requiring multi-region deployment  
**Optimization**: Geographic distribution with performance optimization

### Prerequisites

- FLARE platform deployed across multiple regions
- Cross-region network connectivity via FLUIDOS federation
- User registered with FLARE API access

### Automated Workflow Steps

#### 1. User Intent Submission

```bash
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {
      "objective": "Balanced_Optimization",
      "workload": {
        "type": "service",
        "name": "global-ml-inference",
        "image": "ml-coordinator:v1.0",
        "resources": {
          "cpu": "8",
          "memory": "32Gi",
          "gpu": {
            "model": "nvidia-rtx-4080",
            "count": 1,
            "memory": "16Gi"
          }
        },
        "scaling": {
          "min_replicas": 3,
          "max_replicas": 3
        }
      },
      "constraints": {
        "regions": ["EU", "US-East", "Asia-Pacific"],
        "max_latency_between_regions": 200,
        "max_total_cost": "25 EUR/hour"
      }
    }
  }'
```

Response

```json
{
  "intent_id": "intent-distributed-2024-01-15-001",
  "status": "pending",
  "estimated_cost": "4.00 EUR/hour",
  "eta_seconds": 450
}
```

#### 2. FLARE Intent Processing (Automated)

FLARE automatically translates the user intent into technical specifications and creates multiple GPU-aware Solvers for distributed deployment:

```yaml
metadata:
  labels:
    flare.io/intent-id: "intent-distributed-2024-01-15-001"
spec:
  selector:
    flavorType: K8Slice
    filters:
      gpuFilters:
      - field: count
        filter: NumberRangeFilter
        data:
          min: 1
          max: 1
      - field: memory
        filter: ResourceRangeSelector
        data:
          min: "16Gi"
      - field: hourly_rate
        filter: NumberRangeFilter
        data:
          max: 8.33  # EUR per hour (25 EUR total / 3 regions)
      locationFilter:
        name: StringFilter
        data:
          value: "EU"  # First region-specific solver
  # Full automation enabled for multi-region deployment
  findCandidate: true
  reserveAndBuy: true
  establishPeering: true
```

#### 3. GPU Discovery and Filtering (FLUIDOS Enhanced)

FLUIDOS automatically discovers and filters distributed GPU resources based on Solver requirements:

**What FLUIDOS does automatically**:

1. **Multi-region Flavor discovery** - Scans all federated providers across specified regions
2. **Native geographic filtering** - Applies region-specific constraints
3. **Latency optimization filtering** - Considers inter-region network requirements
4. **PeeringCandidate creation** - Only for geographically distributed resources

**Discovery Results** (handled by FLUIDOS):

- EU Region: RTX 4080 (16Gi) - €0.75/hour - Germany ✓ (matches regional constraint)
- US-East Region: RTX 4090 (24Gi) - €0.85/hour - Virginia ✓ (matches regional constraint)
- Asia-Pacific Region: A100 (40Gi) - €2.40/hour - Singapore ✓ (matches regional constraint)

#### 4. Reservation and Contract Creation (FLUIDOS)

FLUIDOS automatically creates multiple contracts across regions:

**FLUIDOS multi-region selection**:

- Creates separate contracts for each region
- EU: RTX 4080 (€0.75/hour), US-East: RTX 4090 (€0.85/hour), Asia-Pacific: A100 (€2.40/hour)
- Geographic distribution optimization ensures global coverage

#### 5. Remote Peering and Allocation (Liqo)

FLUIDOS triggers Liqo to establish multi-region peering:

**What happens automatically**:

1. FLUIDOS creates Allocation resources for each region
2. Liqo establishes secure tunnels to all three provider regions
3. Virtual nodes appear in consumer cluster representing each region

#### 6. Workload Deployment (FLARE)

FLARE automatically generates and deploys distributed workloads across regions:

**What FLARE does**:

1. Creates namespace with multi-region offloading configuration
2. Generates coordinated Services for global deployment
3. Configures inter-region communication and load balancing
4. Applies all resources with geographic distribution settings

**Status Update to User**:

```json
{
  "intent_id": "intent-distributed-2024-01-15-001",
  "status": "completed",
  "total_cost": "4.00 EUR/hour",
  "deployments": [
    {
      "region": "EU",
      "gpu": "RTX 4080 (16Gi)",
      "location": "Germany", 
      "endpoint": "https://eu.global-ml.example.com",
      "latency_to_coordinator": "45ms"
    },
    {
      "region": "US-East", 
      "gpu": "RTX 4090 (24Gi)",
      "location": "Virginia",
      "endpoint": "https://us.global-ml.example.com", 
      "latency_to_coordinator": "120ms"
    },
    {
      "region": "Asia-Pacific",
      "gpu": "A100 (40Gi)", 
      "location": "Singapore",
      "endpoint": "https://ap.global-ml.example.com",
      "latency_to_coordinator": "180ms"
    }
  ]
}
```

## Summary

These use cases demonstrate FLARE's versatility across different application domains:

- **Cost Optimization**: AI inference, fine-tuning, batch processing
- **Performance Maximization**: AI training, HPC simulations  
- **Latency Minimization**: Video analytics, edge inference
- **Specialized Requirements**: Multi-tenancy, geographic distribution

Each scenario follows the same automated pattern:

1. **User Intent Submission** - Simple API call
2. **FLARE Processing** - Automatic translation to technical specifications
3. **FLUIDOS Discovery** - GPU resource discovery and filtering
4. **Resource Allocation** - Contract creation and peering
5. **Workload Deployment** - Automated deployment and monitoring

This demonstrates FLARE's ability to abstract complex GPU federation workflows into simple, intent-based interactions suitable for users without deep Kubernetes expertise.