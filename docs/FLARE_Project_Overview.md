# FLARE Project Overview

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement and Proposed Solution](#problem-statement-and-proposed-solution)
3. [FLARE Architecture on FLUIDOS](#flare-architecture-on-fluidos)
   - [System Architecture Overview](#system-architecture-overview)
   - [Core Technical Components](#core-technical-components)
   - [Core Workflows](#core-workflows)
4. [API Specifications](#api-specifications)
   - [API Resources](#api-resources)
   - [Intent Schema](#intent-schema)
5. [GPU Resource Management](#gpu-resource-management)
   - [GPU Discovery and Annotation System](#gpu-discovery-and-annotation-system)
   - [Vendor Integration](#vendor-integration)
   - [GPU Pooling Evolution](#gpu-pooling-evolution)
6. [Implementation Roadmap](#implementation-roadmap)
7. [User Guide](#user-guide)
8. [Admin Guide](#admin-guide)
9. [Use Cases and Applications](#use-cases-and-applications)
10. [Cost Optimization Scenarios](#cost-optimization-scenarios)
11. [Results](#results)
    - [Architecture and Integration Success](#architecture-and-integration-success)
    - [Efficient GPUs Placement](#efficient-gpus-placement)
    - [Technical Excellence Achieved](#technical-excellence-achieved)
12. [Additional Resources](#additional-resources)

## Executive Summary

**FLARE** (**F**ederated **L**iquid **R**esources **E**xchange) is a GPU pooling platform for AI and HPC applications. Built on [FLUIDOS](https://fluidos.eu/), it enables dynamic GPU sharing across cloud providers through intent-based allocation and federated multi-tenant cloud architecture. FLARE uses FLUIDOS as its federation engine, leveraging Kubernetes custom resources, the [REAR protocol](https://github.com/fluidos-project/REAR), the [Liqo](https://liqo.io/) cross-cluster networking, and [Capsule](https://projectcapsule.dev/) to create a multi-tenant GPU pooling across several cloud providers. The platform increases GPU utilization while reducing their idle time through dynamic resource sharing.

## Problem Statement and Proposed Solution

### Problem Statement
Enterprise GPU infrastructure faces significant inefficiencies. GPU underutilization is a common problem, with clusters idle over 30% of the time due to workload variance and rigid static allocation models. Resource fragmentation compounds this issue as organizations operate isolated GPU pools across multiple cloud providers without coordination. Manual GPU orchestration creates bottlenecks where workload placement requires specialized knowledge of each provider's offerings and manual intervention from infrastructure teams. This results in cost inefficiency where organizations pay for fixed GPU reservations regardless of actual usage patterns.

### FLARE Solution
FLARE addresses these challenges through distributed operating system capabilities designed for GPU federation. The platform leverages cross-cluster resource discovery implemented by FLUIDOS to enable real-time GPU advertisement across providers, creating visibility into available resources regardless of physical location. Intent-based orchestration allows users to specify high-level requirements such as "minimize cost for batch inference," which are automatically translated into provider-specific infrastructure configurations. Federated networking through Liqo establishes transparent connectivity between clusters, eliminating network complexity typically associated with multi-cloud deployments. Resource negotiation handles contract-based GPU reservation with automated peering establishment, removing manual coordination between providers and consumers.

## FLARE Architecture on FLUIDOS

### System Architecture Overview

FLARE operates as an intelligent orchestration layer on top of FLUIDOS, transforming complex multi-cluster GPU management into simple intent-based operations. The architecture consists of three main layers:

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

For complete architecture details see: [FLARE Architecture Document](FLARE_Architecture.md)

### Core Technical Components

**FLARE Platform Layer** provides the intelligence and automation:

1. **API Gateway** ([Details](FLARE_Architecture.md#key-components))

   - RESTful interface for intent submission
   - Authentication via Capsule multi-tenancy
   - Converts user intents to FLUIDOS Solver CRs
   - Returns tracking IDs for asynchronous operations

2. **Intent Processor** ([Details](FLARE_Architecture.md#key-components))

   - Objective-based optimization selection
   - Translates "minimize cost" into specific GPU filters
   - Applies optimization strategies (cost, performance, latency)
   - Validates against resource quotas and policies

3. **Resource Controller** ([Details](FLARE_Architecture.md#key-components))

   - Monitors FLUIDOS Solver progression through Discovery → Contract → Allocation
   - Manages multi-provider resource coordination
   - Handles cleanup on intent deletion
   - Maintains resource state consistency

4. **Workload Controller** ([Details](FLARE_Architecture.md#key-components))

   - Creates Kubernetes namespaces with Capsule tenant isolation
   - Deploys NamespaceOffloading CRs for remote execution
   - Manages workload resources (Deployments, Services, ConfigMaps)
   - Monitors health and handles failures

5. **Status Manager** ([Details](FLARE_Architecture.md#key-components))

   - Provides real-time intent status via API
   - Aggregates status from multiple controllers
   - Manages webhook notifications for status changes
   - Maintains intent history and audit logs

**FLUIDOS Integration** required these enhancements:

- GPU Flavor creation from node annotations
- GPU-specific filtering in Solvers

### Core Workflows

FLARE implements several workflows that demonstrate its capabilities:

1. **GPU Provider Setup Flow** ([Detailed Flow](FLARE_Architecture.md#1-gpu-provider-setup-flow))

   - Providers annotate GPU nodes with standardized metadata
   - FLUIDOS automatically creates Flavors for each GPU configuration
   - Resources become discoverable via REAR protocol
   - No manual catalog management required

2. **Basic GPU Allocation Flow** ([Detailed Flow](FLARE_Architecture.md#2-basic-gpu-allocation-flow))

   - User submits intent: "I need 4 A100 GPUs for training"
   - FLARE creates Solver with GPU requirements
   - FLUIDOS discovers matching providers
   - Automated contract negotiation and peering
   - Workload deployed to virtual GPU node
   - Complete in <5 seconds for standard requests

3. **No GPU Requirements Met Flow** ([Detailed Flow](FLARE_Architecture.md#3-no-gpu-requirements-met-flow))

   - Graceful handling when no providers match
   - Alternative suggestions provided
   - Queue for future availability
   - Notification when resources become available

4. **GPU Resource Contention Flow** ([Detailed Flow](FLARE_Architecture.md#4-gpu-resource-contention-flow))

   - Priority-based allocation during high demand
   - Preemption policies for critical workloads
   - Fair-share algorithms for multi-tenant scenarios
   - Cost-based bidding for scarce resources

5. **GPU Provider Failure Flow** ([Detailed Flow](FLARE_Architecture.md#5-gpu-provider-failure-flow))

   - Automatic failover to alternative providers
   - State preservation during migration
   - Minimal disruption to running workloads
   - Self-healing through continuous reconciliation

See also [FLUIDOS Basic Workflow](FLUIDOS_Basic_Workflow.md) for understanding the underlying federation mechanisms

## API Specifications

FLARE exposes a simple yet powerful RESTful API that abstracts away infrastructure complexity:

### API Resources

Three primary resources enable complete GPU lifecycle management:

1. **Intents API** (`/api/v1/intents`) ([Full Specification](FLARE_API_Reference.md#api-endpoints))

   - Submit GPU workload requests with business objectives
   - Track intent status and resource allocation
   - Modify or cancel running intents
   - Query historical intent data

2. **Resources API** (`/api/v1/resources`) ([Full Specification](FLARE_API_Reference.md#api-endpoints))

   - Discover available GPU types across federation
   - Query real-time availability and pricing
   - Reserve resources for future use
   - Monitor resource utilization metrics

3. **Tokens API** (`/api/v1/auth/tokens`) ([Full Specification](FLARE_API_Reference.md#authentication))

   - Generate API access tokens
   - Manage token lifecycle and permissions
   - Integration with enterprise SSO
   - Role-based access control via Capsule

### Intent Schema

The core innovation is the Intent schema that captures user requirements without infrastructure details:

```json
{
  "intent": {
    "objective": "Latency_Minimization",
    "workload": {
      "type": "service",
      "name": "llm-inference",
      "image": "lmsysorg/sglang:latest",
      "commands": [
        "python3 -m sglang.launch_server --model-path $MODEL_ID --port 8000"
      ],
      "env": [
        "MODEL_ID=meta-llama/Llama-2-7b-chat-hf",
        "CUDA_VISIBLE_DEVICES=0"
      ],
      "ports": [
        {
          "port": 8000,
          "protocol": "TCP",
          "expose": true,
          "domain": "llm-api.mycompany.com"
        }
      ],
      "resources": {
        "cpu": "4",
        "memory": "16Gi",
        "gpu": {
          "model": "nvidia-a100",
          "count": 1,
          "memory": "40Gi",
          "tier": "premium"
        }
      },
      "storage": {
        "volumes": [
          {
            "name": "model-cache",
            "size": "50Gi",
            "type": "persistent",
            "path": "/root/.cache"
          }
        ]
      },
      "secrets": [
        {
          "name": "huggingface-token",
          "env": "HF_TOKEN"
        }
      ]
    },
    "constraints": {
      "max_hourly_cost": "5 EUR",
      "location": "EU",
      "max_latency_ms": 50
    },
    "sla": {
      "availability": "99.9%",
      "max_interruption_time": "2m"
    }
  }
}
```

**For complete API documentation see**: [FLARE API Reference](FLARE_API_Reference.md)

- [Workload Intent Schema](FLARE_API_Reference.md#workload-intent-schema) - Request format details
- [Complete Examples](FLARE_API_Reference.md#complete-examples) - Real-world usage patterns
- [Error Codes](FLARE_API_Reference.md#error-codes) - Troubleshooting guide

## Technical Differentiation

- **Cross-Provider Federation** differentiates FLARE from vendor-specific solutions by using FLUIDOS to create provider-agnostic GPU pools. While competitors focus on single-cloud optimization, FLARE treats AWS, Google Cloud, Azure, and independent GPU providers as a unified resource pool.

- **Intent-Based Management** shifts the paradigm from infrastructure-focused to outcome-focused operations. Users specify business objectives such as cost minimization or latency optimization, while FLUIDOS handles the technical complexity of provider selection, resource allocation, and workload placement.

- **Real-Time Resource Discovery** through the REAR protocol eliminates dependencies on centralized registries or marketplaces. GPU providers advertise availability directly to the federation, creating a resilient and responsive resource discovery mechanism.

- **Transparent Networking** via Liqo ensures remote GPU workloads behave identically to local resources. Existing monitoring, logging, and management tools work unchanged, reducing operational complexity for distributed deployments.

- **Capsule Multi-Tenancy** integration provides security through namespace isolation for federated workloads, ensuring organizational boundaries are maintained across providers.

- **Annotation-Based GPU Metadata** using standardized annotations creates consistent GPU characteristics representation across heterogeneous hardware vendors (NVIDIA, AMD, Intel, etc.), enabling reliable resource matching and allocation decisions.

## GPU Resource Management

FLARE improves GPU resource management through intelligent abstraction and automation:

### GPU Discovery and Annotation System

FLARE uses a comprehensive annotation system that provides rich GPU metadata for intelligent resource matching:

**Core GPU Annotations** ([Full Reference](FLARE_GPU_Annotations_Reference.md)):

- `gpu.fluidos.eu/vendor`: Manufacturer (nvidia, amd, intel)
- `gpu.fluidos.eu/model`: Specific model (a100, h100, mi300x)
- `gpu.fluidos.eu/memory`: VRAM per GPU (40Gi, 80Gi)
- `gpu.fluidos.eu/count`: Number of GPUs per node
- `gpu.fluidos.eu/tier`: Performance classification (premium, standard, economy)

**Advanced Annotations** for optimization:

- `gpu.fluidos.eu/interconnect`: NVLink, InfinityFabric, PCIe
- `gpu.fluidos.eu/topology`: DGX, HGX, custom
- `gpu.fluidos.eu/multi_gpu_efficiency`: Scaling factor for distributed training
- `network.fluidos.eu/bandwidth`: Inter-node network performance
- `cost.fluidos.eu/hourly_rate`: Dynamic pricing information

The [GPU Annotations Reference](FLARE_GPU_Annotations_Reference.md) provides comprehensive annotation specifications.

### Vendor Integration

FLARE provides standardized annotation mappings from existing GPU operator labels:

**NVIDIA Integration** ([Mapping Guide](NVIDIA_GPU_Labels_Mapping.md)):

- Documents mapping between NVIDIA GPU Operator labels and FLARE annotations
- Provides specifications for extracting model, memory, compute capability
- Defines performance tier classification based on architecture
- Enables consistent annotation standards across NVIDIA hardware

**AMD Integration** ([Mapping Guide](AMD_GPU_Labels_Mapping.md)):

- Provides mapping specifications for AMD GPU Operator labels
- Documents MI-series accelerator annotation patterns
- Handles ROCm compatibility requirements
- Supports mixed vendor deployments through standardized annotations

### GPU Pooling Evolution

FLARE transforms GPU pooling from manual to fully automated:

**Manual Process** (Original FLUIDOS)

- Manual node labeling
- Flavor creation via YAML
- Manual Solver configuration
- Complex workload deployment

**Semi-Automated** (Enhanced FLUIDOS)

- Automatic Flavor generation from annotations
- Simplified Solver creation
- Automated peering establishment

**Fully Automated** (FLARE)

```bash
curl -X POST https://flare-api/intents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "intent": {
      "objective": "Cost_Minimization", 
      "workload": {
        "type": "job",
        "name": "gpu-training",
        "image": "pytorch/pytorch:2.1.0-cuda12.1-cudnn8-devel",
        "resources": {
          "gpu": {
            "count": 4, 
            "model": "nvidia-a100"
          }
        }
      }
    }
  }'
```

For detailed GPU management see:

- [GPU Pooling Guide](FLARE_GPU_Pooling_Guide.md) - Complete workflow comparison
- [GPU Annotations Reference](FLARE_GPU_Annotations_Reference.md) - Annotation specifications
- [Efficient GPU Management](FLARE_placeholder.md) - Optimization algorithms

## Implementation Roadmap

### Phase 1: FLUIDOS Integration

This phase focuses on proving GPU federation viability through FLUIDOS. GPU Flavor creation from standardized node labels enables automatic resource advertisement across providers. Consumer-side GPU filtering provides immediate functionality, allowing organizations to begin using federated GPU resources.

The [FLUIDOS Basic Workflow](FLUIDOS_Basic_Workflow.md) provides essential background on resource federation concepts that underpin FLARE's GPU pooling capabilities.

### Phase 2: FLARE Platform

This phase centers on user experience improvements. The RESTful API enables intent submission rather than infrastructure specifications. Automated solver creation and lifecycle management removes manual FLUIDOS operations requirements. Workload deployment automation eliminates Kubernetes expertise needs for end users, while real-time status tracking provides visibility into distributed resource utilization.

The [GPU Pooling Guide](FLARE_GPU_Pooling_Guide.md) shows the workflow evolution from manual to automated GPU orchestration.

### Phase 3: Advanced Features

The FLARE architecture supports extensibility for future enhancements. The platform's modular design enables potential integration of advanced scheduling algorithms and multi-objective optimization to balance cost, latency, and performance based on workload characteristics.

## User Guide

The [QuickStart Guide](FLARE_QuickStart_Guide.md) provides hands-on experience with local development clusters, demonstrating how to submit intents, monitor status, and deploy workloads using FLARE's API. This guide is designed to help users quickly understand the platform's capabilities and get started with GPU pooling.

## Admin Guide

The [Admin Guide](FLARE_Admin_Guide.md) provides detailed instructions for setting up and managing FLARE clusters in real-world environments.

## Use Cases and Applications

FLARE is designed to support a wide range of AI and HPC workloads across multiple domains. The platform's intent-based API allows users to specify high-level requirements without worrying about underlying infrastructure details. This enables seamless integration with existing applications and workflows:

- [AI Inference Service](FLARE_Sample_Use_Cases.md#1-ai-inference-service)
- [High-Performance AI Training](FLARE_Sample_Use_Cases.md#2-high-performance-ai-training)
- [LLM Fine-Tuning](FLARE_Sample_Use_Cases.md#3-llm-fine-tuning)
- [Scientific Computing](FLARE_Sample_Use_Cases.md#4-high-performance-computing)
- [Real-Time Video Analytics](FLARE_Sample_Use_Cases.md#5-real-time-video-analytics)

Complete list of [Sample Use Cases](FLARE_Sample_Use_Cases.md).

## Results

FLARE research validated the hypothesis that FLUIDOS could serve as an effective enabler for distributed GPU architectures. The project demonstrated significant achievements in architecture design, system integration, and advanced algorithm development through proof-of-concept implementation and simulation.

### Architecture and Integration Success

**FLUIDOS Federation Validation**:

- Successfully integrated FLUIDOS federation capabilities with GPU-specific requirements
- Demonstrated FLUIDOS effectiveness as a foundation for distributed cloud architectures
- Validated the REAR protocol's ability to handle GPU resource advertisement and discovery

**System Integration Excellence**:

- Seamlessly integrated Liqo cross-cluster networking for transparent GPU access
- Successfully implemented Capsule multi-tenancy for secure namespace isolation
- Achieved intent-based API abstraction that eliminates infrastructure complexity

**Multi-Provider Architecture**:

- Validated federation across multiple GPUs providers

### Efficient GPUs Placement

The FLARE implementation let to elaborate a sophisticated mathematical models for optimal GPU placement, a significant advancement beyond the original project scope:

- Developed multi-objective optimization algorithms balancing cost, performance, and latency
- Created constraint satisfaction models handling complex hardware and compliance requirements
- Designed scheduling algorithms for optimal resource allocation
- Advanced algorithms definition and performance analysis
- Created foundation for future AI/ML infrastructure optimization research
- Generated intellectual property with significant commercial potential

Complete algorithm specifications and performance analysis available in [Efficient GPU Management](FLARE_placeholder.md).

### Technical Excellence Achieved

**Operational Transformation**:

- Reduced GPU provisioning time from hours to seconds
- Eliminated manual configuration through complete automation
- Simplified GPU workload deployment from multiple manual steps to single API call

**Performance Optimization**:

- Achieved resource discovery across global providers federation
- Created seamless multi-provider experience for end users
- Demonstrated transparent multi-cluster GPU workload execution

**System Reliability**:

- Implemented self-healing through continuous reconciliation
- Achieved security through multi-tenant namespace isolation

Complete results analysis are available in [Final Project Review](FLARE_Final_Project_Review.md).

## Additional Resources

- **[FLUIDOS Project](https://github.com/fluidos-project)** - Base infrastructure platform
- **[Liqo Project](https://liqo.io)** - Multi-cluster connectivity layer
- **[Fake GPU Operator](https://github.com/run-ai/fake-gpu-operator)** - GPU simulation for development and testing
- **[Capsule Project](https://projectcapsule.dev)** - Multi-tenancy solution for Kubernetes
