# FLARE API Reference Documentation

## Overview

This document provides the API specification for FLARE (Federated Liquid Resources Exchange) workload submission. The FLARE API Gateway allows users to submit GPU workloads using a simple, intuitive JSON format without requiring Kubernetes knowledge.

## API Resources

The FLARE API Gateway exposes three main resources:

- **Intents** (`/api/v1/intents`) - Submit and manage GPU workloads
- **Resources** (`/api/v1/resources`) - Query available GPU resources
- **Tokens** (`/api/v1/auth/tokens`) - Manage API authentication

## Table of Contents

1. [Workload Intent Schema](#workload-intent-schema)
2. [Authentication](#authentication)
3. [API Endpoints](#api-endpoints)
4. [Complete Examples](#complete-examples)
5. [Response Formats](#response-formats)
6. [Error Codes](#error-codes)

## Workload Intent Schema

### Base Structure

```json
{
  "intent": {                         // required
    "objective": "<optimization_objective>",  // required
    "workload": {                     // required
      // Workload specification
    },
    "constraints": {                  // optional
      // Deployment constraints
    },
    "sla": {                         // optional
      // Service level requirements
    }
  }
}
```

### Optimization Objectives

The `objective` field specifies the primary optimization goal for FLARE resource allocation:

```json
"objective": "Cost_Minimization" | "Performance_Maximization" | "Latency_Minimization" | "Balanced_Optimization"
```

**Available Objectives:**

- **`"Cost_Minimization"`** - Prioritize lowest cost resources, accept longer startup times and potentially lower performance
- **`"Performance_Maximization"`** - Prioritize highest performance GPUs regardless of cost
- **`"Latency_Minimization"`** - Optimize for lowest network latency and fastest startup times
- **`"Balanced_Optimization"`** - Balance cost, performance, and latency factors

### Workload Specification

#### Core Fields

```json
"workload": {
  "type": "service" | "job" | "batch",  // required
  "name": "string",              // required - Unique workload identifier
  "image": "string",             // required - Docker image
  "commands": ["string"],        // optional - Container startup commands
  "env": ["KEY=value"],          // optional - Environment variables
  "ports": [                     // optional - Network ports (for services)
    {
      "port": 8000,              // required
      "protocol": "TCP",         // optional - defaults to TCP
      "expose": true,            // optional - Make accessible externally
      "domain": "string"         // optional - Custom domain
    }
  ],
  "resources": {                 // required - Resource requirements
    // See Resource Specifications section
  },
  "storage": {                   // optional - Storage requirements
    // See Storage section
  },
  "secrets": [                   // optional - Credentials/secrets
    {
      "name": "string",          // required
      "env": "ENV_VAR_NAME"      // required
    }
  ],
  "scaling": {                   // optional - Auto-scaling (services only)
    // See Scaling section
  },
  "job": {                       // optional - Job-specific settings
    // See Job Settings section
  }
}
```

#### Resource Specifications

```json
"resources": {
  "cpu": "string",              // optional - CPU cores (e.g., "8", "4.5")
  "memory": "string",           // optional - RAM (e.g., "32Gi", "16Gi")
  "gpu": {                      // optional - GPU requirements
    "model": "string",          // optional - GPU model preference (e.g., "nvidia-h100", "nvidia-a100", "amd-mi300x", "any")
    "count": 1,                 // optional - Number of GPUs (default: 1)
    "memory": "string",         // optional - GPU memory requirement (e.g., "16Gi", "80Gi")
    "cores": 1024,              // optional - CUDA cores requirement
    "clock_speed": "string",    // optional - GPU frequency requirement (e.g., "1.5G", "2.0G")
    "compute_capability": "string", // optional - CUDA compute capability (e.g., "8.0", "9.0")
    "architecture": "string",   // optional - GPU architecture (e.g., "hopper", "ampere")
    "tier": "string",           // optional - GPU tier preference
    "shared": false,            // optional - Allow GPU sharing with other workloads (default: false)
    "interconnect": "string",   // optional - GPU interconnect preference
    "interruptible": false,     // optional - Allow spot/preemptible instances (default: false)
    "multi_instance": false,    // optional - Support MIG (Multi-Instance GPU) (default: false)
    "dedicated": false,         // optional - Require dedicated GPU (default: false)
    "fp32_tflops": "string",    // optional - Performance requirement (e.g., "19.5", "83.0")
    "topology": "string",       // optional - Multi-GPU topology ("all-to-all", "nvswitch", "ring", "mesh")
    "multi_gpu_efficiency": "string" // optional - Multi-GPU efficiency score (e.g., "0.95", "0.85")
  }
}
```

**Note**: The GPU requirements above represent high-level preferences that FLARE translates into FLUIDOS resource filters. For advanced filtering (ranges, multiple criteria), FLARE converts these preferences into appropriate FLUIDOS filter expressions automatically.

##### GPU Models

- `nvidia-h100` - NVIDIA H100 (latest high-end, Hopper architecture)
- `nvidia-a100` - NVIDIA A100 (data center, Ampere architecture)
- `nvidia-rtx-4090` - NVIDIA RTX 4090 (gaming/prosumer, Ada Lovelace)
- `nvidia-rtx-4080` - NVIDIA RTX 4080 (gaming, Ada Lovelace)
- `nvidia-rtx-3090` - NVIDIA RTX 3090 (gaming/prosumer, Ampere)
- `nvidia-v100` - NVIDIA V100 (older data center, Volta)
- `nvidia-t4` - NVIDIA T4 (inference optimized, Turing)
- `amd-mi300x` - AMD MI300X (data center, CDNA3)
- `amd-rx-7900xtx` - AMD RX 7900XTX (gaming, RDNA3)
- `any` - Any available GPU

##### GPU Architectures

- `hopper` - NVIDIA Hopper (nvidia-h100) - Latest architecture
- `ada-lovelace` - NVIDIA Ada Lovelace (RTX 40 series)
- `ampere` - NVIDIA Ampere (nvidia-a100, RTX 30 series)
- `turing` - NVIDIA Turing (nvidia-t4, RTX 20 series)
- `volta` - NVIDIA Volta (nvidia-v100)
- `cdna3` - AMD CDNA3 (amd-mi300x)
- `rdna3` - AMD RDNA3 (amd-rx-7900xtx)
- `any` - Any architecture

##### GPU Tiers

- `premium` - High-end data center GPUs (nvidia-h100, nvidia-a100) - highest performance
- `standard` - Mid-range GPUs (nvidia-rtx-4080, nvidia-rtx-3080) - balanced price/performance
- `gaming` - Gaming GPUs (nvidia-rtx-4090, nvidia-rtx-3090) - cost-effective for many workloads
- `inference` - Inference-optimized (nvidia-t4) - optimized for ML inference
- `budget` - Lower-end options - minimal cost
- `any` - Any available tier

##### GPU Interconnects

- `nvlink` - NVIDIA NVLink (fastest, for multi-GPU setups) - 600-900 GB/s
- `nvswitch` - NVIDIA NVSwitch (highest bandwidth) - up to 2.4 TB/s
- `infiniband` - InfiniBand (high-performance networking) - 200-400 Gb/s
- `pcie` - PCIe (standard connection) - 32-64 GB/s
- `any` - Any available interconnect

##### Multi-GPU Topologies

- `all-to-all` - Full mesh connectivity between GPUs
- `nvswitch` - NVIDIA NVSwitch topology for high bandwidth
- `ring` - Ring topology for collective operations
- `mesh` - Mesh topology for distributed workloads

##### CUDA Compute Capabilities

- `9.0` - Hopper (nvidia-h100)
- `8.6` - Ampere (RTX 30 series)
- `8.0` - Ampere (nvidia-a100)
- `7.5` - Turing (RTX 20 series, nvidia-t4)
- `7.0` - Volta (nvidia-v100)
- `any` - Any capability

#### Storage Specifications

```json
"storage": {
  "volumes": [                  // optional - Storage volumes
    {
      "name": "string",         // required - Volume identifier
      "size": "string",         // required - Size (e.g., "100Gi", "1Ti")
      "type": "persistent" | "temporary",  // required
      "path": "string",         // required - Mount path in container
      "source": {               // optional - External source
        "type": "s3" | "gcs" | "azure",  // required if source specified
        "uri": "string",        // required if source specified
        "credentials": "string" // required if source specified
      }
    }
  ]
}
```

#### Scaling Configuration (Services)

```json
"scaling": {
  "min_replicas": 1,            // optional - Minimum instances (default: 1)
  "max_replicas": 10,           // optional - Maximum instances (default: 10)
  "auto_scale": true,           // optional - Enable auto-scaling (default: false)
  "target_cpu_percent": 70,     // optional - CPU threshold for scaling (default: 70)
  "target_gpu_percent": 80      // optional - GPU threshold for scaling (default: 80)
}
```

#### Job Settings

```json
"job": {
  "parallel_tasks": 1,          // optional - Number of parallel instances (default: 1)
  "max_retries": 3,             // optional - Retry attempts on failure (default: 3)
  "timeout": "string",          // optional - Max execution time (e.g., "2h", "30m")
  "completion_policy": "All" | "Any"  // optional - Success criteria (default: "All")
}
```

### Constraints

#### Currency Support
**Current Status**: EUR only  
**Format**: `"<amount> EUR"` (e.g., "10 EUR")  
**Rate Format**: `"<amount> EUR/hour"` (e.g., "0.45 EUR/hour")

```json
"constraints": {
  "max_hourly_cost": "string",     // optional - Max cost per hour (e.g., "10 EUR")
  "max_total_cost": "string",      // optional - Max total cost for jobs
  "location": "string",            // optional - Geographic preference
  "availability_zone": "string",   // optional - Specific zone
  "max_latency_ms": 100,           // optional - Max network latency (default: 100)
  "deadline": "string",            // optional - ISO 8601 timestamp
  "preemptible": true,             // optional - Allow spot instances (default: false)
  "providers": ["string"],         // optional - Preferred providers
  "availability": {                // optional - Provider availability requirements
    "window_start": "string",            // optional - Availability window start (e.g., "09:00")
    "window_end": "string",              // optional - Availability window end (e.g., "17:00")
    "timezone": "string",                // optional - Timezone (e.g., "Europe/Berlin")
    "days_of_week": ["string"],          // optional - Days available (e.g., ["Mon", "Tue"])
    "blackout_dates": ["string"],        // optional - Unavailable dates (ISO 8601)
    "maintenance_windows": [             // optional - Scheduled maintenance
      {
        "start": "string",               // required - Start time (ISO 8601)
        "end": "string",                 // required - End time (ISO 8601)
        "frequency": "weekly"            // optional - Frequency (weekly, monthly) (default: "weekly")
      }
    ]
  },
  "negotiation": {                 // optional - Real-time negotiation parameters (FUTURE IMPROVEMENT)
    "max_negotiation_rounds": 3,         // optional - Max rounds of resource negotiation (default: 3)
    "price_flexibility": 0.15,          // optional - Price flexibility percentage (default: 0.15)
    "resource_flexibility": 0.10,       // optional - Resource substitution flexibility (default: 0.10)
    "timeout_seconds": 300,             // optional - Negotiation timeout (default: 300)
    "fallback_strategy": "queue",       // optional - Fallback if negotiation fails (default: "queue")
    "auto_accept_threshold": 0.05       // optional - Auto-accept if within threshold (default: 0.05)
    // NOTE: Advanced negotiation features require FLUIDOS enhancements for auction-based resource allocation
  },
  "energy": {                      // optional - Energy efficiency constraints
    "max_carbon_footprint": "string",   // optional - Max CO2 per hour (e.g., "50g CO2/h")
    "renewable_energy_only": false,     // optional - Require renewable energy sources (default: false)
    "energy_efficiency_rating": "A",    // optional - Minimum efficiency rating (A-F)
    "power_usage_effectiveness": 1.4,   // optional - Max PUE for data centers (default: 2.0)
    "green_certified_only": false       // optional - Require green certifications (default: false)
  },
  "compliance": {                  // optional - Regulatory and compliance requirements
    "data_residency": ["string"],       // optional - Required data locations (e.g., ["EU"])
    "certifications": ["string"],       // optional - Required certifications (ISO27001, SOC2)
    "encryption_at_rest": true,         // optional - Require data encryption (default: false)
    "encryption_in_transit": true,      // optional - Require transit encryption (default: false)
    "audit_logging": true,              // optional - Require audit logs (default: false)
    "gdpr_compliant": true,             // optional - GDPR compliance required (default: false)
    "hipaa_compliant": false            // optional - HIPAA compliance required (default: false)
  },
  "performance": {                 // optional - Performance guarantees
    "min_network_bandwidth": "string",  // optional - Min bandwidth (e.g., "10Gbps")
    "max_jitter_ms": 10,                // optional - Max network jitter (default: 50)
    "min_uptime_percent": 99.9,         // optional - Minimum uptime guarantee (default: 99.0)
    "max_cold_start_time": "string",    // optional - Max startup time (e.g., "30s")
    "min_uptime_percent": 99.9         // optional - Minimum uptime guarantee (default: 99.0)
  },
  "security": {                    // optional - Security requirements
    "network_isolation": "private",     // optional - Network isolation level (default: "public")
    "firewall_rules": [                 // optional - Custom firewall rules
      {
        "port": 22,                     // required - Port number
        "protocol": "TCP",              // required - Protocol
        "source": "10.0.0.0/8",        // required - Source CIDR
        "action": "allow"               // required - Action (allow/deny)
      }
    ],
    "vpn_access": false,                // optional - Require VPN access (default: false)
    "bastion_host": false,              // optional - Require bastion host (default: false)
    "intrusion_detection": true,       // optional - Enable IDS/IPS (default: false)
    "vulnerability_scanning": true     // optional - Enable vulnerability scans (default: false)
  }
}
```

#### Location Options

- `EU` - European Union
- `US` - United States
- `Asia` - Asia Pacific
- `Canada` - Canada
- `Australia` - Australia
- `Brazil` - Brazil
- `India` - India
- `Japan` - Japan
- `Singapore` - Singapore
- `any` - Any location

#### Fallback Strategies

- `queue` - Queue request until resources available
- `lower_tier` - Accept lower-tier GPU if available
- `shared` - Accept shared GPU resources
- `spot` - Accept spot/preemptible instances
- `fail` - Fail immediately if requirements not met

#### Network Isolation Levels

- `public` - Public internet access
- `private` - Private network only

### SLA (Service Level Agreement)

```json
"sla": {
  "availability": "string",           // optional - Uptime requirement (e.g., "99.9%") (default: "99.0%")
  "max_interruption_time": "string", // optional - Max acceptable downtime (e.g., "5m")
  "backup_strategy": "string"         // optional - Data backup approach (default: "none")
}
```


## Authentication

FLARE API Gateway uses token-based authentication. All API requests must include an authorization header with a valid API token.

### Authentication Header

```http
Authorization: Bearer <your-api-token>
```

### Authentication Flow

1. **Obtain API Token**: Contact your FLARE administrator or use the FLARE API Gateway token management endpoints
2. **Include Token**: Add the `Authorization: Bearer <token>` header to all requests
3. **Token Validation**: Tokens are validated on each request

### Error Responses

**401 Unauthorized** - Missing or invalid token:

```json
{
  "error": "unauthorized",
  "message": "Invalid or missing API token",
  "code": "INVALID_TOKEN"
}
```

**403 Forbidden** - Valid token but insufficient permissions:

```json
{
  "error": "forbidden", 
  "message": "Insufficient permissions for this operation",
  "code": "INSUFFICIENT_PERMISSIONS"
}
```

## API Endpoints

### Submit Workload Intent

**POST** `/api/v1/intents`

Submit a new workload intent for execution.

**Headers:**

- `Authorization: Bearer <token>` (required)
- `Content-Type: application/json` (required)

**Request Body:** Complete workload intent JSON

**Response:**

```json
{
  "intent_id": "string",
  "status": "pending",
  "message": "Intent received and processing",
  "estimated_cost": "5.50 EUR/hour",
  "estimated_start_time": "2024-01-15T10:30:00Z"
}
```

### Get Intent Status

**GET** `/api/v1/intents/{intent_id}`

Retrieve the current status of a submitted intent.

**Headers:**

- `Authorization: Bearer <token>` (required)

**Response:**

```json
{
  "intent_id": "string",
  "status": "running" | "pending" | "failed" | "completed",
  "workload_url": "https://my-workload.flare.example.com",
  "current_cost": "12.45 EUR",
  "runtime": "2h 15m",
  "gpu_utilization": "85%",
  "message": "Workload running successfully"
}
```

### List User Intents

**GET** `/api/v1/intents`

List all intents for the authenticated user.

**Headers:**

- `Authorization: Bearer <token>` (required)

### Cancel Intent

**DELETE** `/api/v1/intents/{intent_id}`

Cancel a running or pending intent.

**Headers:**

- `Authorization: Bearer <token>` (required)

### Get Available Resources

**GET** `/api/v1/resources`

Query available GPU resources across the federation.

**Headers:**

- `Authorization: Bearer <token>` (required)

**Response:**

```json
{
  "available_gpus": [
    {
      "model": "nvidia-h100",
      "count": 8,
      "memory": "80Gi",
      "location": "eu-west-1",
      "cost_per_hour": "4.50 EUR",
      "provider": "provider-1"
    }
  ]
}
```

### Token Management

#### Create API Token

**POST** `/api/v1/auth/tokens`

Create a new API token for FLARE API Gateway authentication.

**Headers:**

- `Content-Type: application/json` (required)
- `Authorization: Bearer <admin-token>` (required for programmatic creation)

**Request Body:**

```json
{
  "name": "string",              // required - Token name/description
  "expires_in": "string",        // optional - Expiration time (e.g., "30d", "1y") (default: "90d")
  "permissions": ["string"],     // optional - Token permissions (default: ["intents:read", "intents:write"])
  "user_id": "string"           // optional - User ID (admin only)
}
```

**Response:**

```json
{
  "token_id": "tok_abc123",
  "token": "flr_1234567890abcdef",
  "name": "My API Token",
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-04-15T10:30:00Z",
  "permissions": ["intents:read", "intents:write"]
}
```

#### List API Tokens

**GET** `/api/v1/auth/tokens`

List all API tokens for the authenticated user.

**Headers:**

- `Authorization: Bearer <token>` (required)

**Response:**

```json
{
  "tokens": [
    {
      "token_id": "tok_abc123",
      "name": "My API Token",
      "created_at": "2024-01-15T10:30:00Z",
      "expires_at": "2024-04-15T10:30:00Z",
      "last_used": "2024-01-20T15:45:00Z",
      "permissions": ["intents:read", "intents:write"],
      "status": "active"
    }
  ]
}
```

#### Revoke API Token

**DELETE** `/api/v1/auth/tokens/{token_id}`

Revoke an API token.

**Headers:**

- `Authorization: Bearer <token>` (required)

#### Verify Token

**GET** `/api/v1/auth/verify`

Verify that a token is valid and check user information.

**Headers:**

- `Authorization: Bearer <token>` (required)

**Response:**

```json
{
  "status": "authenticated",
  "user": "demo-user",
  "permissions": ["intents:read", "intents:write", "resources:read"],
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Response:**

```json
{
  "message": "Token revoked successfully",
  "token_id": "tok_abc123",
  "revoked_at": "2024-01-20T16:00:00Z"
}
```

#### Token Permissions

Available permissions for API tokens:

- `intents:read` - Read access to workload intents
- `intents:write` - Create and modify workload intents  
- `intents:delete` - Delete workload intents
- `resources:read` - Read access to resource information
- `tokens:manage` - Manage API tokens (admin only)
- `admin:*` - Full administrative access

## Complete Examples

### Authentication Example

First, create an API token:

```bash
curl -X POST https://flare-api.example.com/api/v1/auth/tokens \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{
    "name": "My Development Token",
    "expires_in": "30d",
    "permissions": ["intents:read", "intents:write", "resources:read"]
  }'
```

Response:

```json
{
  "token_id": "tok_abc123",
  "token": "flr_1234567890abcdef",
  "name": "My Development Token",
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-02-14T10:30:00Z",
  "permissions": ["intents:read", "intents:write", "resources:read"]
}
```

Use the token for subsequent requests:

```bash
curl -X GET https://flare-api.example.com/api/v1/resources \
  -H "Authorization: Bearer flr_1234567890abcdef"
```

### 1. LLM Inference Service (High Performance)

**Request:**

```bash
curl -X POST https://flare-api.example.com/api/v1/intents \
  -H "Authorization: Bearer flr_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d @intent.json
```

**Intent JSON:**

```json
{
  "intent": {
    "objective": "Performance_Maximization",
    "workload": {
      "type": "service",
      "name": "deepseek-r1-inference",
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
          "protocol": "TCP",
          "expose": true,
          "domain": "deepseek-api.mycompany.com"
        }
      ],
      "resources": {
        "cpu": "8",
        "memory": "32Gi",
        "gpu": {
          "model": "nvidia-h100",
          "count": 2,
          "memory": "80Gi",
          "tier": "premium"
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
      },
      "secrets": [
        {
          "name": "huggingface-token",
          "env": "HF_TOKEN"
        }
      ],
      "scaling": {
        "min_replicas": 2,
        "max_replicas": 10,
        "auto_scale": true,
        "target_gpu_percent": 80
      }
    },
    "constraints": {
      "max_hourly_cost": "50 EUR",
      "location": "EU",
      "max_latency_ms": 100
    },
    "sla": {
      "availability": "99.9%",
      "max_interruption_time": "2m"
    }
  }
}
```

### 2. Video Processing Batch Job (Cost Optimized)

```json
{
  "intent": {
    "objective": "Cost_Minimization",
    "workload": {
      "type": "batch",
      "name": "video-processing-batch",
      "image": "ffmpeg-gpu:latest",
      "commands": [
        "python3 process_videos.py --input /data/input --output /data/output --quality high"
      ],
      "env": [
        "THREADS=8",
        "QUALITY=high",
        "FORMAT=mp4"
      ],
      "resources": {
        "cpu": "4",
        "memory": "16Gi",
        "gpu": {
          "model": "nvidia-rtx-4090",
          "count": 1,
          "tier": "gaming",
          "shared": true
        }
      },
      "storage": {
        "volumes": [
          {
            "name": "video-input",
            "size": "500Gi",
            "type": "temporary",
            "path": "/data/input",
            "source": {
              "type": "s3",
              "uri": "s3://my-videos/raw/",
              "credentials": "aws-readonly"
            }
          },
          {
            "name": "video-output",
            "size": "1Ti",
            "type": "persistent",
            "path": "/data/output"
          }
        ]
      },
      "secrets": [
        {
          "name": "aws-credentials",
          "env": "AWS_ACCESS_KEY_ID"
        }
      ],
      "job": {
        "parallel_tasks": 20,
        "max_retries": 2,
        "timeout": "6h"
      }
    },
    "constraints": {
      "max_total_cost": "200 EUR",
      "deadline": "2024-12-15T00:00:00Z",
      "preemptible": true,
      "location": "any"
    }
  }
}
```

### 3. ML Training Job (Distributed)

```json
{
  "intent": {
    "objective": "Performance_Maximization",
    "workload": {
      "type": "job",
      "name": "llama-distributed-training",
      "image": "pytorch/pytorch:2.1.0-cuda12.1-cudnn8-devel",
      "commands": [
        "torchrun --nproc_per_node=4 --nnodes=2 train.py --model llama-7b --epochs 10"
      ],
      "env": [
        "WORLD_SIZE=8",
        "NCCL_DEBUG=INFO",
        "MASTER_PORT=29500",
        "NCCL_IB_DISABLE=1"
      ],
      "resources": {
        "cpu": "16",
        "memory": "128Gi",
        "gpu": {
          "model": "nvidia-a100",
          "count": 4,
          "memory": "40Gi",
          "interconnect": "nvlink",
          "tier": "premium"
        }
      },
      "storage": {
        "volumes": [
          {
            "name": "training-data",
            "size": "2Ti",
            "type": "persistent",
            "path": "/data/train"
          },
          {
            "name": "checkpoints",
            "size": "500Gi",
            "type": "persistent",
            "path": "/checkpoints"
          },
          {
            "name": "logs",
            "size": "50Gi",
            "type": "temporary",
            "path": "/logs"
          }
        ]
      },
      "job": {
        "parallel_tasks": 2,
        "max_retries": 1,
        "timeout": "24h",
        "completion_policy": "All"
      }
    },
    "constraints": {
      "max_total_cost": "1000 EUR",
      "location": "US",
      "providers": ["aws", "gcp"]
    },
    "sla": {
      "backup_strategy": "checkpoint"
    }
  }
}
```

### 4. Data Processing Pipeline (Spark)

```json
{
  "intent": {
    "objective": "Cost_Minimization",
    "workload": {
      "type": "batch",
      "name": "spark-etl-pipeline",
      "image": "spark-gpu:3.4.0",
      "commands": [
        "spark-submit --class com.example.DataProcessor --master local[*] /app/processor.jar"
      ],
      "env": [
        "SPARK_DRIVER_MEMORY=8g",
        "SPARK_EXECUTOR_MEMORY=16g",
        "INPUT_PATH=s3://data-lake/raw/",
        "OUTPUT_PATH=s3://data-lake/processed/",
        "PARTITION_SIZE=128MB"
      ],
      "resources": {
        "cpu": "8",
        "memory": "64Gi",
        "gpu": {
          "model": "any",
          "count": 1,
          "shared": true,
          "tier": "standard"
        }
      },
      "storage": {
        "volumes": [
          {
            "name": "spark-temp",
            "size": "200Gi",
            "type": "temporary",
            "path": "/tmp/spark"
          }
        ]
      },
      "secrets": [
        {
          "name": "aws-access-key",
          "env": "AWS_ACCESS_KEY_ID"
        },
        {
          "name": "aws-secret-key",
          "env": "AWS_SECRET_ACCESS_KEY"
        }
      ],
      "job": {
        "parallel_tasks": 10,
        "max_retries": 3,
        "timeout": "4h"
      }
    },
    "constraints": {
      "max_total_cost": "100 EUR",
      "preemptible": true,
      "deadline": "2024-12-20T08:00:00Z",
      "location": "any"
    }
  }
}
```

### 5. Interactive Jupyter Notebook Service

```json
{
  "intent": {
    "objective": "Latency_Minimization",
    "workload": {
      "type": "service",
      "name": "gpu-jupyter-lab",
      "image": "jupyter/tensorflow-notebook:latest",
      "commands": [
        "start-notebook.sh --NotebookApp.token='' --NotebookApp.password=''"
      ],
      "env": [
        "JUPYTER_ENABLE_LAB=yes",
        "CUDA_VISIBLE_DEVICES=0"
      ],
      "ports": [
        {
          "port": 8888,
          "protocol": "TCP",
          "expose": true,
          "domain": "jupyter.researcher.edu"
        }
      ],
      "resources": {
        "cpu": "4",
        "memory": "16Gi",
        "gpu": {
          "model": "nvidia-rtx-4090",
          "count": 1,
          "tier": "gaming"
        }
      },
      "storage": {
        "volumes": [
          {
            "name": "notebooks",
            "size": "100Gi",
            "type": "persistent",
            "path": "/home/jovyan/work"
          },
          {
            "name": "datasets",
            "size": "500Gi",
            "type": "persistent",
            "path": "/data"
          }
        ]
      },
      "scaling": {
        "min_replicas": 1,
        "max_replicas": 1,
        "auto_scale": false
      }
    },
    "constraints": {
      "max_hourly_cost": "5 EUR",
      "location": "EU",
      "max_latency_ms": 50
    },
    "sla": {
      "availability": "95%"
    }
  }
}
```

### 6. Communication-Aware Distributed Training (Balanced Optimization)

This example demonstrates how FLARE optimizes for both cost and communication performance in distributed training scenarios.

```json
{
  "intent": {
    "objective": "Balanced_Optimization",
    "workload": {
      "type": "job",
      "name": "gpt-7b-distributed-training",
      "image": "nvcr.io/nvidia/pytorch:24.01-py3",
      "commands": [
        "torchrun --nproc_per_node=4 --nnodes=1 train_gpt.py"
      ],
      "env": [
        "NCCL_SOCKET_IFNAME=eth0",
        "NCCL_IB_DISABLE=1",
        "PYTORCH_CUDA_ALLOC_CONF=max_split_size_mb:512"
      ],
      "resources": {
        "cpu": "32",
        "memory": "256Gi",
        "gpu": {
          "count": 4,
          "memory": "40Gi",
          "model": "nvidia-a100",
          "interconnect": "nvlink"
        }
      },
      "storage": {
        "volumes": [
          {
            "name": "training-dataset",
            "size": "1Ti",
            "type": "persistent",
            "path": "/data/dataset"
          },
          {
            "name": "model-checkpoints",
            "size": "200Gi",
            "type": "persistent",
            "path": "/checkpoints"
          }
        ]
      },
      "job": {
        "max_retries": 2,
        "timeout": "12h",
        "completion_policy": "All"
      }
    },
    "constraints": {
      "max_hourly_cost": "25 EUR",
      "max_network_latency": "50ms",
      "location": "EU"
    },
    "sla": {
      "availability": "99.5%",
      "max_interruption_time": "10m"
    }
  }
}
```

**Key Communication-Aware Features:**

- `interconnect: "nvlink"` - Requires high-speed GPU interconnect  
- `max_network_latency: "50ms"` - Ensures acceptable latency for federated operations
- `objective: "Balanced_Optimization"` - FLARE balances cost with communication performance

**FLARE's Optimization Process:**

1. Filters providers with NVLink-enabled nodes having 4+ GPUs
2. Calculates total cost including interconnect overhead  
3. Prioritizes low-latency configurations for multi-GPU workloads
4. Falls back to alternative topologies if preferred interconnect unavailable
5. Selects configuration with best cost/performance ratio

## Response Formats

### Success Response

```json
{
  "status": "success",
  "intent_id": "intent-abc123",
  "message": "Intent submitted successfully",
  "data": {
    "estimated_cost": "8.50 EUR/hour",
    "estimated_start_time": "2024-01-15T10:35:00Z",
    "selected_provider": "provider-2",
    "gpu_allocation": {
      "model": "nvidia-h100",
      "count": 2,
      "location": "eu-west-1"
    }
  }
}
```

### Error Response

```json
{
  "status": "error",
  "error_code": "GPU_MODEL_UNAVAILABLE",
  "message": "No nvidia-h100 GPUs available in EU region",
  "suggestions": [
    {
      "alternative": "nvidia-a100 GPUs available",
      "cost_difference": "+15%",
      "performance_impact": "-10%"
    }
  ]
}
```

### Status Response

```json
{
  "intent_id": "intent-abc123",
  "status": "running",
  "workload": {
    "name": "deepseek-r1-inference",
    "type": "service",
    "url": "https://deepseek-api.mycompany.com",
    "start_time": "2024-01-15T10:35:00Z",
    "runtime": "2h 15m"
  },
  "resources": {
    "provider": "provider-2",
    "location": "eu-west-1",
    "gpu": {
      "model": "nvidia-h100",
      "count": 2,
      "utilization": "85%"
    }
  },
  "costs": {
    "hourly_rate": "8.50 EUR/hour",
    "total_cost": "19.12 EUR",
    "currency": "EUR"
  },
  "performance": {
    "latency_p50": "45ms",
    "latency_p99": "120ms",
    "throughput": "150 req/s"
  }
}
```

## Error Codes

| Code | Description | Action |
|------|-------------|---------|
| `INVALID_FORMAT` | JSON format error | Fix JSON syntax |
| `MISSING_REQUIRED_FIELD` | Required field missing | Add missing field |
| `INSUFFICIENT_RESOURCES` | No matching resources | Adjust requirements or wait |
| `GPU_MODEL_UNAVAILABLE` | Requested GPU model not found | Try alternative models or wait for availability |
| `GPU_MEMORY_INSUFFICIENT` | Available GPUs have less memory than required | Reduce memory requirements or try different regions |
| `GPU_COUNT_INSUFFICIENT` | Not enough GPUs available for multi-GPU request | Reduce GPU count or try distributed deployment |
| `COST_LIMIT_EXCEEDED` | Estimated cost too high | Increase budget or reduce resources |
| `QUOTA_EXCEEDED` | User quota exceeded | Contact support or wait |
| `INVALID_TOKEN` | Invalid or missing API token | Check Authorization header |
| `INSUFFICIENT_PERMISSIONS` | Token lacks required permissions | Request elevated permissions |
| `TOKEN_EXPIRED` | API token has expired | Generate new token |
| `TOKEN_REVOKED` | API token has been revoked | Generate new token |
| `AUTHENTICATION_FAILED` | Authentication failed | Check token format |
| `REGION_UNAVAILABLE` | Requested region not available | Choose different region |

