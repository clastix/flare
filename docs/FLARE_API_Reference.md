# FLARE API Reference Documentation

## Overview

This document provides the API specification for FLARE (Federated Liquid Resources Exchange) workload submission. The FLARE API Gateway allows users to submit GPU workloads using a simple, intuitive JSON format without requiring Kubernetes knowledge.

## API Resources

The FLARE API Gateway exposes three main resources:

- **Intents** (`/intents`) - Submit and manage GPU workloads
- **Resources** (`/resources`) - Query available GPU resources
- **Tokens** (`/auth/tokens`) - Manage API authentication

## Table of Contents

1. [Workload Intent Schema](#workload-intent-schema)
2. [Authentication](#authentication)
3. [API Endpoints](#api-endpoints)
4. [Complete Examples](#complete-examples)
5. [Response Formats](#response-formats)
6. [Error Handling](#error-handling)

## Workload Intent Schema

### Base Structure

```json
{
  "intent": {
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
"objective": "Performance_Maximization" | "Cost_Minimization" | "Latency_Minimization" | "Energy_Efficiency" | "Balanced_Optimization"
```

**Available Objectives:**

- **`"Performance_Maximization"`** - Prioritize highest performance GPUs regardless of cost
- **`"Cost_Minimization"`** - Prioritize lowest cost resources, accept longer startup times and potentially lower performance  
- **`"Latency_Minimization"`** - Optimize for lowest network latency and fastest startup times
- **`"Energy_Efficiency"`** - Prioritize renewable energy sources and minimize carbon footprint, may accept higher costs for greener infrastructure
- **`"Balanced_Optimization"`** - Balance cost, performance, and latency factors

### Workload Specification

#### Core Fields

```json
"workload": {
  "type": "service" | "batch",  // required
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
  "deployment_strategy": "string", // optional - Resource deployment strategy ("colocated", "distributed", "flexibile")
  "communication_pattern": "string", // optional - Inter-process communication pattern ("all-reduce", "pipeline", "independent")
  "scaling": {                   // optional - Auto-scaling (services only)
    // See Scaling section
  },
  "batch": {                     // optional - Batch-specific settings
    // See Batch section
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
    "memory_min": "string",     // optional - Minimum GPU memory requirement (e.g., "16Gi", "80Gi")
    "memory_max": "string",     // optional - Maximum GPU memory requirement (e.g., "80Gi", "128Gi")
    "cores_min": 1024,          // optional - Minimum CUDA cores requirement
    "cores_max": 2048,          // optional - Maximum CUDA cores requirement
    "clock_speed_min": "string", // optional - Minimum GPU frequency requirement (e.g., "1.5G", "2.0G")
    "compute_capability": "string", // optional - CUDA compute capability (e.g., "8.0", "9.0")
    "architecture": "string",   // optional - GPU architecture (e.g., "hopper", "ampere")
    "tier": "string",           // optional - GPU tier preference
    "shared": false,            // optional - Allow GPU sharing with other workloads (default: false)
    "interconnect": "string",   // optional - GPU interconnect preference
    "interruptible": false,     // optional - Allow spot/preemptible instances (default: false)
    "multi_instance": false,    // optional - Support MIG (Multi-Instance GPU) (default: false)
    "dedicated": false,         // optional - Require dedicated GPU (default: false)
    "fp32_tflops": 19.5,        // optional - Performance requirement (e.g., 19.5, 83.0)
    "topology": "string",       // optional - Multi-GPU topology ("all-to-all", "nvswitch", "ring", "mesh")
    "multi_gpu_efficiency": 0.95 // optional - Multi-GPU efficiency score (e.g., 0.95, 0.85)
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

#### Batch Configuration

```json
"batch": {
  "parallel_tasks": 1,              // optional - Number of parallel instances (default: 1)
  "max_retries": 3,                 // optional - Retry attempts on failure (default: 3)
  "timeout": "2h",                  // optional - Max execution time (e.g., "2h", "30m")
  "completion_policy": "All"        // optional - Success criteria: "All" or "Any" (default: "All")
}
```

### Constraints

The `constraints` section allows you to specify deployment requirements, limits, and preferences for your workload.

#### Basic Constraints

```json
"constraints": {
  "max_hourly_cost": "10 EUR",         // optional - Max cost per hour 
  "max_total_cost": "100 EUR",         // optional - Max total cost for batch workloads
  "location": "EU",                    // optional - Geographic preference
  "availability_zone": "eu-west-1a",   // optional - Specific availability zone
  "max_latency_ms": 100,               // optional - Max network latency (default: 100)
  "deadline": "2024-12-31T23:59:59Z",  // optional - ISO 8601 deadline timestamp
  "preemptible": false,                // optional - Allow spot instances (default: false)
  "providers": ["AWS", "GCP"]          // optional - Preferred cloud providers
}
```

#### Provider Availability Requirements

```json
"availability": {
  "window_start": "09:00",             // required - Daily availability start (HH:MM)
  "window_end": "17:00",               // required - Daily availability end (HH:MM)
  "timezone": "Europe/Berlin",         // required - Timezone identifier
  "days_of_week": ["Mon", "Tue", "Wed", "Thu", "Fri"], // required - Available days
  "blackout_dates": ["2024-12-25", "2024-01-01"],      // optional - Unavailable dates (ISO 8601)
  "maintenance_windows": [             // required - Scheduled maintenance windows
    {
      "start": "2024-01-15T02:00:00Z", // required - Maintenance start (ISO 8601)
      "end": "2024-01-15T04:00:00Z",   // required - Maintenance end (ISO 8601)
      "frequency": "weekly"            // optional - Frequency (default: "weekly")
    }
  ]
}
```

#### Energy Efficiency Constraints

```json
"energy": {
  "max_carbon_footprint": "50g CO2/h", // optional - Max CO2 emissions per hour
  "renewable_energy_only": false,      // optional - Require renewable sources (default: false)
  "energy_efficiency_rating": "A",     // optional - Min efficiency rating (A-F)
  "power_usage_effectiveness": 2.0,    // optional - Max PUE for data centers (default: 2.0)
  "green_certified_only": false        // optional - Require green certifications (default: false)
}
```

#### Security Requirements

```json
"security": {
  "network_isolation": "public",       // optional - Network isolation level (default: "public")
  "firewall_rules": [                  // optional - Custom firewall rules
    {
      "port": 22,                      // required - Port number
      "protocol": "TCP",               // required - Protocol (TCP/UDP)
      "source": "10.0.0.0/8",         // required - Source CIDR block
      "action": "allow"                // required - Action (allow/deny)
    }
  ],
  "vpn_access": false,                 // optional - Require VPN access (default: false)
  "bastion_host": false,               // optional - Require bastion host (default: false)
  "intrusion_detection": false,        // optional - Enable IDS/IPS (default: false)
  "vulnerability_scanning": false      // optional - Enable vulnerability scans (default: false)
}
```

#### Performance Guarantees

```json
"performance": {
  "min_network_bandwidth": "10Gbps",   // optional - Min network bandwidth
  "max_jitter_ms": 50,                 // optional - Max network jitter (default: 50)
  "min_uptime_percent": 99.0,          // optional - Min uptime guarantee (default: 99.0)
  "max_cold_start_time": "30s",        // optional - Max startup time
  "gpu_utilization_target": 0.80,      // optional - Target GPU utilization (default: 0.80)
  "memory_utilization_target": 0.80    // optional - Target memory utilization (default: 0.80)
}
```

#### Compliance Requirements

```json
"compliance": {
  "data_residency": ["EU"],            // optional - Required data locations
  "certifications": ["ISO27001", "SOC2"], // optional - Required certifications
  "encryption_at_rest": false,         // optional - Require data encryption (default: false)
  "encryption_in_transit": false,      // optional - Require transit encryption (default: false)
  "audit_logging": false,              // optional - Require audit logs (default: false)
  "gdpr_compliant": false,             // optional - GDPR compliance required (default: false)
  "hipaa_compliant": false             // optional - HIPAA compliance required (default: false)
}
```

#### Resource Negotiation (Advanced)

```json
"negotiation": {
  "max_negotiation_rounds": 3,         // optional - Max negotiation rounds (default: 3)
  "price_flexibility": 0.15,           // optional - Price flexibility % (default: 0.15)
  "resource_flexibility": 0.10,        // optional - Resource flexibility % (default: 0.10)
  "timeout_seconds": 300,              // optional - Negotiation timeout (default: 300)
  "fallback_strategy": "queue",        // optional - Fallback strategy (default: "queue")
  "auto_accept_threshold": 0.05        // optional - Auto-accept threshold (default: 0.05)
}
```

#### Constraint Options Reference

**Location Options:**
- `EU` - European Union
- `US` - United States  
- `Asia` - Asia Pacific
- `Canada`, `Australia`, `Brazil`, `India`, `Japan`, `Singapore`
- `any` - Any location

**Provider Options:**
- `AWS` - Amazon Web Services
- `GCP` - Google Cloud Platform

**Network Isolation Levels:**
- `public` - Public internet access
- `private` - Private network only

**Fallback Strategies:**
- `queue` - Queue until resources available
- `lower_tier` - Accept lower-tier GPU
- `shared` - Accept shared resources
- `spot` - Accept spot/preemptible instances

**Energy Efficiency Ratings:**
- `A` (highest) through `F` (lowest)

**Compliance Certifications:**
- `ISO27001` - Information security standard
- `SOC2` - Service Organization Control 2

### SLA (Service Level Agreement)

```json
"sla": {
  "availability": "string",           // optional - Uptime requirement (e.g., "99.9%") (default: "99.0%")
  "max_interruption_time": "string", // optional - Max acceptable downtime (e.g., "5m")
  "backup_strategy": "string"         // optional - Data backup approach (default: none)
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

## Error Handling

### Error Response Format

All FLARE API errors follow a consistent response format with HTTP status codes, error codes, messages, and actionable suggestions.

```json
{
  "status": 400,
  "error_code": "GPU_MODEL_UNAVAILABLE",
  "message": "The requested GPU model 'nvidia-h200' is not available",
  "suggestions": [
    {
      "alternative": "nvidia-h100",
      "cost_difference": "+15%",
      "performance_impact": "-5%"
    },
    {
      "alternative": "nvidia-a100",
      "cost_difference": "-30%", 
      "performance_impact": "-20%"
    }
  ]
}
```

### Error Codes Reference

#### Request Format Errors (400 Bad Request)

- **`INVALID_FORMAT`** - Request JSON is malformed or invalid
- **`MISSING_REQUIRED_FIELD`** - Required field missing from request body

#### Resource Availability Errors (409 Conflict)

- **`INSUFFICIENT_RESOURCES`** - Not enough resources available to fulfill request
- **`GPU_MODEL_UNAVAILABLE`** - Requested GPU model not available 
- **`GPU_MEMORY_INSUFFICIENT`** - Available GPUs don't meet memory requirements
- **`GPU_COUNT_INSUFFICIENT`** - Not enough GPUs available for requested count
- **`REGION_UNAVAILABLE`** - Requested location/region not available

#### Cost and Quota Errors (402 Payment Required / 429 Too Many Requests)

- **`COST_LIMIT_EXCEEDED`** - Request exceeds specified cost limits
- **`QUOTA_EXCEEDED`** - User has exceeded resource quotas

#### Authentication Errors (401 Unauthorized)

- **`INVALID_TOKEN`** - API token is invalid or malformed
- **`TOKEN_EXPIRED`** - API token has expired
- **`TOKEN_REVOKED`** - API token has been revoked
- **`AUTHENTICATION_FAILED`** - Authentication credentials are incorrect

#### Authorization Errors (403 Forbidden)

- **`INSUFFICIENT_PERMISSIONS`** - User lacks permissions for this operation

### Error Examples

#### GPU Model Unavailable

```json
{
  "status": 409,
  "error_code": "GPU_MODEL_UNAVAILABLE", 
  "message": "The requested GPU model 'nvidia-h200' is not currently available",
  "suggestions": [
    {
      "alternative": "nvidia-h100",
      "cost_difference": "+10%",
      "performance_impact": "-8%"
    }
  ]
}
```

#### Cost Limit Exceeded

```json
{
  "status": 402,
  "error_code": "COST_LIMIT_EXCEEDED",
  "message": "Estimated cost $45.50/hour exceeds limit of $20.00/hour",
  "suggestions": [
    {
      "alternative": "nvidia-rtx-4090 (2 GPUs)",
      "cost_difference": "-60%",
      "performance_impact": "-25%"
    },
    {
      "alternative": "Enable preemptible instances", 
      "cost_difference": "-70%",
      "performance_impact": "May be interrupted"
    }
  ]
}
```

#### Invalid Authentication

```json
{
  "status": 401,
  "error_code": "TOKEN_EXPIRED",
  "message": "API token expired on 2024-01-15T10:30:00Z",
  "suggestions": []
}
```

#### Resource Quota Exceeded

```json
{
  "status": 429,
  "error_code": "QUOTA_EXCEEDED", 
  "message": "GPU quota exceeded: using 8/8 allocated GPUs",
  "suggestions": [
    {
      "alternative": "Wait for existing workloads to complete",
      "cost_difference": "Free",
      "performance_impact": "Delayed execution"
    },
    {
      "alternative": "Request quota increase",
      "cost_difference": "Contact support",
      "performance_impact": "None"
    }
  ]
}
```

## API Endpoints

### Submit Workload Intent

**POST** `/intents`

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

**GET** `/intents/{intent_id}`

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

**GET** `/intents`

List all intents for the authenticated user.

**Headers:**

- `Authorization: Bearer <token>` (required)

### Cancel Intent

**DELETE** `/intents/{intent_id}`

Cancel a running or pending intent.

**Headers:**

- `Authorization: Bearer <token>` (required)

### Get Available Resources

**GET** `/resources`

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

**POST** `/auth/tokens`

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

**GET** `/auth/tokens`

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

**DELETE** `/auth/tokens/{token_id}`

Revoke an API token.

**Headers:**

- `Authorization: Bearer <token>` (required)

#### Verify Token

**GET** `/auth/verify`

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
curl -X POST https://flare-api.example.com/auth/tokens \
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
curl -X GET https://flare-api.example.com/resources \
  -H "Authorization: Bearer flr_1234567890abcdef"
```

### 1. LLM Inference Service (High Performance)

**Request:**

```bash
curl -X POST https://flare-api.example.com/intents \
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
          "memory_min": "80Gi",
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
      "type": "batch",
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
          "memory_min": "40Gi",
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
      "type": "batch",
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
          "memory_min": "40Gi",
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
  "intent_id": "intent-abc123",
  "status": "success", 
  "message": "Intent submitted successfully",
  "estimated_cost": "8.50 EUR/hour",
  "estimated_start_time": "2024-01-15T10:35:00Z"
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
  "workload_url": "https://deepseek-api.mycompany.com",
  "current_cost": "19.12 EUR", 
  "runtime": "2h 15m",
  "gpu_utilization": "85%",
  "message": "Workload running successfully"
}
```


