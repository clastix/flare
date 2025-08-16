# FLUIDOS Basic Workflow

## Table of Contents

1. [FLUIDOS Overview](#fluidos-overview)
2. [Prerequisites Check](#prerequisites-check)
3. [Part 1: Provider Resources (Flavors)](#part-1-provider-resources-flavors)
4. [Part 2: Consumer Resource Discovery](#part-2-consumer-resource-discovery)
5. [Part 3: Resource Reservation and Contract](#part-3-resource-reservation-and-contract)
6. [Part 4: Establishing Cluster Peering](#part-4-establishing-cluster-peering)
7. [Part 5: Deploying Workloads on Virtual Nodes](#part-5-deploying-workloads-on-virtual-nodes)
8. [Part 6: Virtual Node from Different Provider](#part-6-virtual-node-from-different-provider)
9. [Part 7: Multiple Virtual Nodes from Same Provider](#part-7-multiple-virtual-nodes-from-same-provider)
10. [Part 8: Troubleshooting Guide](#part-8-troubleshooting-guide)
11. [Part 9: Quick Reference Commands](#part-9-quick-reference-commands)
12. [FLUIDOS Controller Summary](#fluidos-controller-summary)
13. [Resource Flow Overview - The Foundation for FLARE](#resource-flow-overview---the-foundation-for-flare)

## FLUIDOS Overview

FLUIDOS (Federated Liquid Infrastructure Distributed Operating System) is like a "marketplace" for computing resources where:

- **Providers** advertise computing resources as specialized Flavors
- **Consumers** discover and reserve resources via intent-based Solvers  
- **REAR Protocol** handles resource advertisement, discovery, and reservation
- **Liqo** establishes secure cross-cluster networking and workload execution

This basic workflow forms the foundation for FLARE's GPU federation platform.

### FLUIDOS Architecture

The key relationship in FLUIDOS is:
```
Worker Node (with label) → Flavor → Contract → Virtual Node
```

Resources flow through these Custom Resources:
```
Solver → Discovery → PeeringCandidates → Reservation → Contract → Allocation → Virtual Node
```

## Prerequisites Check

Before we start, let's verify our environment is ready:

```bash
# Check all three clusters are running
kubectl config get-contexts | grep fluidos

# Expected output:
# * kind-fluidos-consumer-1    kind-fluidos-consumer-1    ...
#   kind-fluidos-provider-1    kind-fluidos-provider-1    ...
#   kind-fluidos-provider-2    kind-fluidos-provider-2    ...
```

## Part 1: Provider Resources (Flavors)

### What is a Flavor?

A **Flavor** is FLUIDOS's way of describing available computing resources. Think of it as a "product catalog entry" that tells consumers what a provider offers:

- Provider worker nodes with `node-role.fluidos.eu/resources=true` become Flavors
- Local Resource Manager creates Flavor CRs with node capacity specs
- Each Flavor has resource characteristics (CPU, memory, pods) and policies
- For GPU resources, Flavors include GPU specifications from node annotations

### Step 1.1: Check Provider Flavors

```bash
# Switch to provider 1 cluster
export KUBECONFIG=fluidos-provider-1-config

# List all Flavors
kubectl get flavors -n fluidos
```

**Expected output:**

```
NAME                      PROVIDER ID   TYPE      AVAILABLE   AGE
fluidos.eu-k8slice-52dd   gorlzgruty    K8Slice   true        5m
fluidos.eu-k8slice-da9d   gorlzgruty    K8Slice   true        5m
```

### Step 1.2: Examine a Flavor in Detail

```bash
# Look at a specific Flavor
kubectl get flavor fluidos.eu-k8slice-52dd -n fluidos -o yaml
```

**Understanding the Flavor specification:**

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Flavor
metadata:
  name: fluidos.eu-k8slice-52dd
spec:
  flavorType:
    typeIdentifier: K8Slice  # Type of resource (container-based)
    typeData:
      characteristics:       # What resources are available
        architecture: amd64
        cpu: "15149658066n"  # ~15 CPU cores in nanocores
        memory: "7564624Ki"  # ~7.5Gi memory
        pods: "110"          # Can run up to 110 pods
        gpu:                 # GPU specifications (if available)
          model: "nvidia-h100"
          cores: "4"
          memory: "80Gi"
      properties:
        latency: 50          # Network latency characteristic
      policies:
        partitionability:    # How resources can be divided
          cpuMin: "0"
          memoryMin: "0"
          gpuMin: "1"        # Minimum GPU allocation
  owner:                     # Provider identity
    domain: fluidos.eu
    nodeID: gorlzgruty
  price:                     # Cost information
    amount: "10.00"
    currency: "EUR"
    period: "hourly"
  availability: true         # Currently available for reservation
```

**Important Fields Explained:**

- **typeIdentifier**: FLUIDOS supports 4 types: K8Slice (containers), VM, Service, Sensor
- **characteristics**: The actual hardware specifications
- **partitionability**: Defines how resources can be divided (important for GPU sharing)
- **availability**: Changes to `false` when reserved by a consumer

### Step 1.3: Understand Flavor Creation

```bash
# Check which nodes have the resource label
kubectl get nodes -l node-role.fluidos.eu/resources=true

# Check Flavor characteristics match node capacity
kubectl get flavors -n fluidos -o jsonpath='{range .items[*]}{.metadata.name}: cpu={.spec.flavorType.typeData.characteristics.cpu}, memory={.spec.flavorType.typeData.characteristics.memory}, pods={.spec.flavorType.typeData.characteristics.pods}{"\n"}{end}'

fluidos.eu-k8slice-ca82: cpu=15017691944n, memory=7566408Ki, pods=110
fluidos.eu-k8slice-d6b4: cpu=15776763969n, memory=7610592Ki, pods=110
```

Only nodes with `node-role.fluidos.eu/resources=true` label become Flavors. This allows providers to control which nodes participate in FLUIDOS.

## Part 2: Consumer Resource Discovery

### What is a Solver?

A **Solver** is the master orchestrator in FLUIDOS. It represents a consumer's need for resources and drives the entire acquisition workflow through three automated phases:

1. **Discovery** (`findCandidate: true`) - Find available resources matching intent filters
2. **Reservation** (`reserveAndBuy: true`) - Reserve selected resource via REAR protocol negotiation
3. **Peering** (`establishPeering: true`) - Establish secure cross-cluster connection via Liqo

Think of a Solver as your "resource shopping assistant" that handles everything from searching to purchasing to setting up delivery.

This three-phase approach forms the foundation for FLARE's high-level intent processing, where user GPU requirements are automatically translated into Solver specifications.

### Step 2.1: Verify Automatic Cluster Discovery

FLUIDOS uses automatic discovery through the Network Manager component. Let's verify that clusters have discovered each other:

```bash
# Switch to consumer cluster
export KUBECONFIG=fluidos-consumer-1-config

# Check for automatically discovered clusters
kubectl get knownclusters -n fluidos
```

If no `KnownClusters` appear, check the Network Manager logs:

```bash
# Check network manager discovery logs
kubectl logs -n fluidos -l app.kubernetes.io/instance=node-local-resource-manager

```

**What is a KnownCluster?**

A `KnownCluster` represents another FLUIDOS node in the federation:

```yaml
apiVersion: network.fluidos.eu/v1alpha1
kind: KnownCluster
spec:
  clusterID: "cluster-provider-1"
  endpoint: "https://provider1.k8s.local:6443"
  domain: "provider1.fluidos.eu"
```

**How Auto-Discovery Works:**

- Network Manager pods multicast advertisements on the local network
- Other FLUIDOS nodes receive these advertisements and create KnownCluster entries
- The REAR protocol uses these entries to enable resource discovery

### Step 2.2: Create a Solver

Let's create a Solver that searches for resources with specific requirements:

```bash
# Apply the solver with pod capacity filter
kubectl apply -f - <<EOF
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: pods-solver
  namespace: fluidos
spec:
  selector:
    flavorType: K8Slice
    filters:
      architectureFilter:
        name: Match
        data:
          value: "amd64"
      cpuFilter:
        name: Range
        data:
          min: "1000m"
      memoryFilter:
        name: Range
        data:
          min: "1Gi"
      podsFilter:
        name: Range
        data:
          min: "10"
  intentID: basic-pods-intent
  findCandidate: true
  reserveAndBuy: false
  establishPeering: false
EOF

# Examine what we created
kubectl get solver pods-solver -n fluidos -o yaml
```

**Understanding Solver Filters:**

FLUIDOS filters enable precise resource matching for diverse requirements. Filter types include:

- **String Filters**: Exact match or selection (architecture, GPU models)
- **Resource Quantity Filters**: Range-based matching (CPU, memory, pods, storage)
- **GPU Filters** (FLARE extension): GPU-specific matching (vendor, model, memory, tier)

The basic solver configuration:

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: pods-solver
  namespace: fluidos
  labels:
    flare.io/scenario: "basic-fluidos"    # FLARE integration label
spec:
  selector:
    flavorType: K8Slice
    filters:
      architectureFilter:
        name: Match
        data:
          value: "amd64"
      cpuFilter:
        name: Range
        data:
          min: "1000m"
      memoryFilter:
        name: Range
        data:
          min: "1Gi"
      podsFilter:
        name: Range
        data:
          min: "10"
      # gpuFilter:
      #   modelFilter:
      #     filter: MatchSelector
      #     data: ["nvidia-h100"]
      #   countFilter:
      #     filter: GreaterThanOrEqual
      #     data: "2"
  intentID: basic-pods-intent
  findCandidate: true
  reserveAndBuy: false
  establishPeering: false
```

### Step 2.3: Watch Discovery in Action

```bash
# Monitor the Solver status
kubectl get solver pods-solver -n fluidos -w

# Expected progression:
# STATUS: Pending -> Running (Finding candidates) -> Solved
```

In another terminal, check the discovery process:

```bash
# Check what Discovery was created
kubectl get discoveries -n fluidos

# Look at Discovery details
kubectl describe discovery discovery-pods-solver -n fluidos
```

**What is a Discovery?**

A `Discovery` is the search operation that queries all known clusters:

```yaml
apiVersion: advertisement.fluidos.eu/v1alpha1
kind: Discovery
spec:
  solverID: "pods-solver"    # Which Solver requested this
  selector:                   # What to search for
    flavorType: K8Slice
    filters: {...}            # Same filters from Solver
```

**What's happening behind the scenes:**

1. Solver controller creates a Discovery resource
2. Discovery Manager queries each `KnownCluster` via REAR protocol
3. Each provider's REAR Manager returns Flavors matching the filters
4. `PeeringCandidates` are created for each matching `Flavor`

### Step 2.4: Examine Discovery Results

```bash
# Check PeeringCandidates (available options)
kubectl get peeringcandidates -n fluidos

# Expected output:
# NAME                          AGE
# pc-52dd-gorlzgruty-fluidos    30s
# pc-da9d-gorlzgruty-fluidos    30s
# pc-30f9-qmgbumr7ak-fluidos    30s
# pc-9c4e-qmgbumr7ak-fluidos    30s

# Look at a specific candidate
kubectl get peeringcandidate <peeringcandidate_name> -n fluidos -o yaml | grep -A15 "flavor:"
```

**What is a PeeringCandidate?**

A `PeeringCandidate` represents a specific Flavor that matches your requirements:

```yaml
apiVersion: advertisement.fluidos.eu/v1alpha1
kind: PeeringCandidate
spec:
  interestedSolverIDs:        # Which Solvers want this
    - "pods-solver"
  flavor:                     # Complete Flavor from provider
    spec:
      flavorType:
        typeData:
          characteristics:
            cpu: "15149658066n"
            memory: "7564624Ki"
  available: true            # Still available for reservation
```

**Key Points:**

- Each `PeeringCandidate` contains the complete `Flavor` specification from the provider
- Multiple `Solvers` can be interested in the same `PeeringCandidate`
- This is where consumer-side filtering would happen for advanced requirements

## Part 3: Resource Reservation and Contract

### Step 3.1: Select and Reserve Resources

Now let's move to the reservation phase. According to the FLUIDOS workflow, setting `reserveAndBuy: true` triggers:

1. Creation of a Reservation CR
2. REAR protocol negotiation with the provider
3. Contract creation upon successful negotiation

```bash
# Enable reservation phase
kubectl patch solver pods-solver -n fluidos --type merge -p '{"spec": {"reserveAndBuy": true}}'

# Watch the reservation process
kubectl get solver pods-solver -n fluidos -w

# Expected progression:
# STATUS: Running (Reserve running) -> Solved
```

### Step 3.2: Understanding the Reservation Process

When you set `reserveAndBuy: true`, FLUIDOS creates a `Reservation`:

```bash
# Check if a Reservation was created
kubectl get reservations -n fluidos

# Check if a Contract was created
kubectl get contracts -n fluidos

# Look at contract details
 kubectl -n fluidos describe $(kubectl get contracts -n fluidos -o name | head -1) | grep -A10 "Spec:"
```

**What is a Reservation?**

A `Reservation` is your booking request for specific resources:

```yaml
apiVersion: reservation.fluidos.eu/v1alpha1
kind: Reservation
spec:
  solverID: "pods-solver"        # Requesting Solver
  buyer:                         # Your cluster identity
    domain: "fluidos.eu"
    nodeID: "nt2zqultl5"
  seller:                        # Provider cluster identity
    domain: "fluidos.eu"
    nodeID: "gorlzgruty"
  flavorRef:                     # Which Flavor to reserve
    name: "fluidos.eu-k8slice-52dd"
  partition:                     # How much to allocate
    cpu: "15000m"
    memory: "7Gi"
```

**What is a Contract?**

A `Contract` is the binding agreement for resource sharing:

```yaml
apiVersion: reservation.fluidos.eu/v1alpha1
kind: Contract
spec:
  buyer:                     # Your cluster identity
    domain: fluidos.eu
    nodeID: nt2zqultl5
  seller:                    # Provider cluster identity
    domain: fluidos.eu
    nodeID: gorlzgruty
  peeringTargetCredentials:  # Critical: Liqo connection info
    clusterID: ...
    clusterName: ...
    endpoint: ...
    token: ...              # Authentication token for peering
  configuration:            # Exact resources allocated
    cpu: "15000m"
    memory: "7Gi"
    pods: "110"
  transactionID: ...        # Tracks the reservation process
```

**Critical Elements:**

- **peeringTargetCredentials**: Contains everything Liqo needs to connect
- **configuration**: The exact resources you'll receive
- **transactionID**: Links to the Transaction that tracks the process

The Contract contains everything needed for Liqo to establish a secure connection between clusters.

### Step 3.3: Check the Transaction

```bash
# Check Transaction (tracks reservation workflow)
kubectl get transaction -n fluidos
```

**What is a Transaction?**

A `Transaction` tracks the entire reservation workflow:

```yaml
apiVersion: reservation.fluidos.eu/v1alpha1
kind: Transaction
spec:
  transactionID: "txn-abc123"
  flavorRef:
    name: "fluidos.eu-k8slice-52dd"
  buyer:
    domain: "fluidos.eu"
    nodeID: "nt2zqultl5"
status:
  phase:
    phase: "Solved"
    message: "Transaction completed"
```

## Part 4: Establishing Cluster Peering

### Step 4.1: Enable Liqo Peering

The final phase creates the actual network connection between clusters:

```bash
# Enable peering phase
kubectl patch solver pods-solver -n fluidos --type merge -p '{"spec": {"establishPeering": true}}'

# Monitor peering establishment
kubectl get solver pods-solver -n fluidos -w

# Expected progression:
# STATUS: Running (Allocation: peering) -> Solved (90 seconds typical)
```

### Step 4.2: Understanding the Allocation Process

```bash
# Monitor Allocation status progression
kubectl get allocation -n fluidos -w

# Expected status progression:
# Provisioning -> ResourceCreation -> Peering -> Active
```

**What is an Allocation?**

An `Allocation` represents resources that have been reserved and are being activated:

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Allocation
spec:
  contract:                     # Links to our Contract
    name: "contract-pods-solver"
    namespace: "fluidos"
  forwarding: false            # Not a proxy allocation
status:
  status: "Active"             # Current state
  message: "Allocation successful, virtual node ready"
```

**Status Progression Explained:**

- `Provisioning`: Initial setup
- `ResourceCreation`: Creating Kubernetes resources
- `Peering`: Establishing Liqo connection
- `Active`: Ready for workload deployment

### Step 4.3: Verify Virtual Infrastructure

```bash
# Check for virtual nodes
kubectl get nodes

# Expected: A new node with name like "liqo-fluidos.eu-k8slice-52dd"

# Examine the virtual node
kubectl describe node $(kubectl get nodes | grep liqo | awk '{print $1}') | grep -A10 "Labels:" | grep -E "liqo.io|fluidos"

# Check Liqo ForeignCluster
kubectl get foreignclusters

# Expected output:
# NAME                 ROLE       AGE
# fluidos-provider-2   Provider   2m53s
```

**Understanding Virtual Nodes:**

Virtual nodes are Kubernetes nodes that represent remote resources:

```yaml
# Key characteristics of a virtual node
Labels:
  liqo.io/type: virtual-node          # Marks as virtual
  liqo.io/remote-cluster-id: xxx      # Which provider
  
Taints:
  virtual-node.liqo.io/not-allowed: NoExecute  # Prevents accidental scheduling

Capacity:                              # Matches Contract
  cpu: 15000m
  memory: 7Gi
  pods: 110
```

**Key Virtual Node Characteristics:**

- **Name**: Contains the Flavor name for identification
- **Labels**: `liqo.io/type=virtual-node` marks it as virtual
- **Taints**: `virtual-node.liqo.io/not-allowed=true:NoExecute` prevents accidental scheduling
- **Capacity**: Matches the Contract configuration

## Part 5: Deploying Workloads on Virtual Nodes

### Step 5.1: Understanding NamespaceOffloading

Before deploying workloads, we need to configure where pods should run. Liqo uses `NamespaceOffloading` to control pod placement:

```bash
# Create a test namespace
kubectl create namespace workload-test

# Apply namespace offloading configuration
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

# If you need to verify the actual cluster ID from your virtual node:
# kubectl get nodes -l liqo.io/type=virtual-node -o jsonpath='{.items[0].metadata.labels.liqo\.io/remote-cluster-id}'

# Check offloading status
kubectl get namespaceoffloading -n workload-test -w

# Expected output:
# NAME             REMOTENAMESPACE   OFFLOADINGPHASE
# workload-test    workload-test     Ready
```

**Pod Offloading Strategies:**

- `Local`: Pods run only on local nodes
- `Remote`: Pods run only on virtual nodes (our choice)
- `LocalAndRemote`: Pods can run anywhere

### Step 5.2: Deploy Your First Remote Workload

```bash
# Deploy nginx on the virtual node
kubectl -n workload-test create deployment nginx-test --image=nginx:1.23 --replicas=2
```

### Step 5.3: Verify Remote Execution

```bash
# Confirm pods are running on virtual node
kubectl get pods -n workload-test -o custom-columns=NAME:.metadata.name,NODE:.spec.nodeName,IP:.status.podIP

# Get pod logs (works transparently across clusters!)
kubectl logs -n workload-test deployment/nginx-test
```

**Key Points:**

- Pod IPs come from the remote cluster's pod CIDR
- Logs and exec commands work transparently through Liqo
- The consumer cluster treats remote pods as if they were local

**What is a ForeignCluster?**

A `ForeignCluster` (Liqo resource) represents the actual network connection established between clusters:

```bash
# Check the ForeignCluster
kubectl get foreignclusters

# Expected output:
# NAME                 ROLE       AGE
# fluidos-provider-1   Provider   2m53s

# Examine the ForeignCluster details
kubectl describe foreigncluster fluidos-provider-1
```

```yaml
apiVersion: liqo.io/v1beta1
kind: ForeignCluster
metadata:
  name: fluidos-provider-1
spec:
  clusterIdentity:
    clusterID: "cluster-provider-1"
    clusterName: "provider-1"
  networkingEnabled: true          # Network tunnel active
  authenticationEnabled: true      # Authentication completed
status:
  role: Provider                   # This cluster provides resources
  outgoingPeering:
    phase: Established            # Connection successful
```

**Key Functions:**

- **Created using credentials** from the Contract's `peeringTargetCredentials`
- **Manages the secure tunnel** between clusters (typically WireGuard)
- **Enables transparent pod communication** across cluster boundaries
- **Provides the foundation** for virtual nodes to appear and function


## Part 6: Virtual Node from Different Provider

Now let's demonstrate FLUIDOS's true power by creating another Solver that will establish a connection to a different provider cluster, giving us parallel workloads running on both providers simultaneously.

### Step 6.1: Create Another Solver

```bash
# Apply solver-02 configuration
kubectl apply -f - <<EOF
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: pods-solver-02
  namespace: fluidos
spec:
  selector:
    flavorType: K8Slice
    filters:
      architectureFilter:
        name: Match
        data:
          value: "amd64"
      cpuFilter:
        name: Range
        data:
          min: "1000m"
      memoryFilter:
        name: Range
        data:
          min: "1Gi"
      podsFilter:
        name: Match
        data:
          value: "110"
  intentID: basic-pods-intent-02
  findCandidate: true
  reserveAndBuy: false
  establishPeering: false
EOF

# Examine the solver-02 configuration
kubectl get solver pods-solver-02 -n fluidos -o yaml | grep -A20 "spec:"
```

The `pods-solver-02` uses a Match filter for exactly 110 pods to demonstrate different filtering strategies.

### Step 6.2: Monitor Discovery Process for pods-solver-02

```bash
# Watch pods-solver-02's discovery phase
kubectl get solver pods-solver-02 -n fluidos -w

# Check PeeringCandidates (should see the same ones from both providers)
kubectl get peeringcandidates -n fluidos

# Verify both solvers have found candidates
kubectl get solver -n fluidos
```

**Expected output:**

```
NAME            INTENT ID          FIND CANDIDATE   RESERVE AND BUY   PEERING   STATUS
pods-solver     intent-sample      true             true              true      Solved
pods-solver-02  intent-sample-02   true             false             false     Solved
```

### Step 6.3: Reserve Resources from Different Provider

FLUIDOS will automatically select a different provider for pods-solver-02 to demonstrate load balancing:

```bash
# Enable reservation for pods-solver-02
kubectl patch solver pods-solver-02 -n fluidos --type merge -p '{"spec": {"reserveAndBuy": true}}'

# Monitor reservation process
kubectl get solver pods-solver-02 -n fluidos -w

# Check contracts - should now have two
kubectl get contracts -n fluidos

# Verify we have contracts with different providers
kubectl get contracts -n fluidos -o custom-columns=NAME:.metadata.name,SELLER:.spec.seller.nodeID
```

**Expected output:**

```
NAME                       SELLER
contract-pods-solver       gorlzgruty      # Provider 1
contract-pods-solver-02    qmgbumr7ak      # Provider 2
```

### Step 6.4: Establish Peering Connection for pods-solver-02

```bash
# Enable peering for pods-solver-02
kubectl patch solver pods-solver-02 -n fluidos --type merge -p '{"spec": {"establishPeering": true}}'

# Monitor peering establishment
kubectl get solver pods-solver-02 -n fluidos -w

# Check allocations - should have two active
kubectl get allocations -n fluidos

# Verify we now have multiple ForeignClusters
kubectl get foreignclusters
```

**Expected output:**

```
NAME                   ROLE       AGE
fluidos-provider-1     Provider   10m
fluidos-provider-2     Provider   2m
```

### Step 6.5: Verify Dual Virtual Node Infrastructure

```bash
# Check for multiple virtual nodes
kubectl get nodes

# Expected: Two virtual nodes, one from each provider
# Local nodes plus:
# liqo-fluidos.eu-k8slice-xxxx   # From provider 1
# liqo-fluidos.eu-k8slice-yyyy   # From provider 2

# Examine virtual node labels to identify providers
kubectl get nodes -l liqo.io/type=virtual-node -o custom-columns=NAME:.metadata.name,CLUSTER:.metadata.labels.liqo\\.io/remote-cluster-id

# Check capacity for all virtual nodes
kubectl get nodes -l liqo.io/type=virtual-node -o custom-columns=NAME:.metadata.name,CPU:.status.allocatable.cpu,MEMORY:.status.allocatable.memory
```

### Step 6.6: Deploy Multi-Provider Workloads

Now let's create workloads that will be distributed across both provider clusters:

```bash
# Create namespace for multi-provider workloads
kubectl create namespace workload-test-02

# Apply namespace offloading configuration (supports both providers)
kubectl apply -f - <<EOF
apiVersion: offloading.liqo.io/v1beta1
kind: NamespaceOffloading
metadata:
  name: offloading
  namespace: workload-test-02
spec:
  clusterSelector:
    nodeSelectorTerms:
      - matchExpressions:
      - key: liqo.io/remote-cluster-id
        operator: In
        values:
          - fluidos-provider-1
        - fluidos-provider-2
  namespaceMappingStrategy: DefaultName
  podOffloadingStrategy: Remote
EOF

# Verify offloading is ready
kubectl get namespaceoffloading -n workload-test-02
```

### Step 6.7: Deploy Load-Balanced Application

```bash
# Deploy nginx with multiple replicas to spread across providers
kubectl -n workload-test-02 create deployment nginx-test-02 --image=nginx:1.23 --replicas=2

# Wait for deployment to be ready
kubectl rollout status deployment/nginx-test-02 -n workload-test-02

# Check deployment status
kubectl get deployment nginx-test-02 -n workload-test-02
```

### Step 6.8: Verify Multi-Provider Pod Distribution

```bash
# Check pod distribution across virtual nodes
kubectl get pods -n workload-test-02 -o wide

# Verify pods are running on different providers
kubectl get pods -n workload-test-02 -o custom-columns=NAME:.metadata.name,NODE:.spec.nodeName,STATUS:.status.phase | grep Running
```

### Step 6.9: Scale across providers

```bash
kubectl scale deployment nginx-test-02 --replicas=12 -n workload-test-02
kubectl get pods -n workload-test-02 -o wide
```

### Step 6.10: Multi-Provider Resource Overview

```bash
# Get comprehensive overview of our federated setup
echo -e "\n1. Active Solvers:"
kubectl get solver -n fluidos -o custom-columns=NAME:.metadata.name,STATUS:.status.findCandidate,RESERVED:.status.reserveAndBuy,PEERED:.status.peering

echo -e "\n2. Provider Contracts:"
kubectl get contracts -n fluidos -o custom-columns=NAME:.metadata.name,SELLER:.spec.seller.nodeID

echo -e "\n3. Active Allocations:"
kubectl get allocations -n fluidos -o custom-columns=NAME:.metadata.name,STATUS:.status.status,CONTRACT:.spec.contract.name

echo -e "\n4. Virtual Infrastructure:"
kubectl get nodes -o custom-columns=NAME:.metadata.name,TYPE:.metadata.labels.liqo\\.io/type,STATUS:.status.conditions[3].status | grep -E "NAME|virtual-node"

echo -e "\n5. Provider Connections:"
kubectl get foreignclusters -o custom-columns=NAME:.metadata.name,ROLE:.status.role,STATUS:.status.outgoingPeering.phase

echo -e "\n6. Workload Distribution:"
kubectl get pods -n workload-test-02 -o custom-columns=NAME:.metadata.name,NODE:.spec.nodeName,STATUS:.status.phase | grep -E "NAME|liqo"

```

## Part 7: Multiple Virtual Nodes from Same Provider

In production environments, a provider cluster may have multiple worker nodes with different characteristics (e.g., different GPU types, architectures, or resource capacities). FLUIDOS represents each worker node as a distinct Flavor, allowing consumers to request specific resources. This section demonstrates how to create multiple virtual nodes from the same provider cluster.

### Step 7.1: Create Additional Solver for Same Provider

Let's create an additional solver (pods-solver-03) that will target a different flavor from the same provider:

```bash
# Create pods-solver-03
kubectl apply -f - <<EOF
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: pods-solver-03
  namespace: fluidos
spec:
  selector:
    flavorType: K8Slice
    filters:
      architectureFilter:
        name: Match
        data:
          value: "amd64"
      cpuFilter:
        name: Range
        data:
          min: "1000m"
      memoryFilter:
        name: Range
        data:
          min: "1Gi"
      podsFilter:
        name: Range
        data:
          min: "10"
  intentID: basic-pods-intent-03
  findCandidate: true
  reserveAndBuy: false
  establishPeering: false
EOF
```

### Step 7.2: Discovery and Reservation

```bash
# Monitor discovery
kubectl get solver pods-solver-03 -n fluidos -w

# Enable reservation
kubectl patch solver pods-solver-03 -n fluidos --type merge -p '{"spec": {"reserveAndBuy": true}}'

# Wait for contract creation
kubectl get contracts -n fluidos

# Both contracts should be from same provider but different flavors
kubectl get contracts -n fluidos -o custom-columns=NAME:.metadata.name,FLAVOR:.spec.flavor.metadata.name,SELLER:.spec.seller.nodeID
```

### Step 7.3: Enable Peering for Additional Virtual Node

```bash
# Enable peering
kubectl patch solver pods-solver-03 -n fluidos --type merge -p '{"spec": {"establishPeering": true}}'

# Monitor the solver - it will get stuck in "Allocation: peering"
kubectl get solver pods-solver-03 -n fluidos -w
```

**Expected Issue:** The pods-solver-03 will get stuck in "Allocation: peering" state. This is due to a FLUIDOS limitation where it checks if a `ForeignCluster` already exists and skips `ResourceSlice` creation.

### Step 7.4: Manual `ResourceSlice` Creation (Workaround)

Manually create an additional `ResourceSlice`:

```bash
# Get the contract details for pods-solver-03
export CONTRACT_NAME=$(kubectl get contracts -n fluidos -o json | jq -r '.items[2].metadata.name')
echo "Contract for pods-solver-03: $CONTRACT_NAME"

# Get the flavor details from the contract
kubectl get contract $CONTRACT_NAME -n fluidos -o yaml | grep -A20 "flavor:" > flavor-details.yaml

# Create ResourceSlice manually
cat <<EOF > second-resourceslice.yaml
apiVersion: authentication.liqo.io/v1beta1
kind: ResourceSlice
metadata:
  name: $CONTRACT_NAME
  namespace: liqo-tenant-fluidos-provider-1
  annotations:
    liqo.io/create-virtual-node: "true"
  labels:
    liqo.io/remote-cluster-id: fluidos-provider-1
    liqo.io/remoteID: fluidos-provider-1
    liqo.io/replication: "true"
spec:
  class: default
  consumerClusterID: fluidos-consumer-1
  providerClusterID: fluidos-provider-1
  resources:
    cpu: 3978561353n
    memory: 15918948Ki
    pods: "110"
EOF

kubectl apply -f second-resourceslice.yaml
```

### Step 7.5: Verify Additional Virtual Node Creation

```bash
# Watch for ResourceSlice acceptance
kubectl get resourceslices -A -w

# Check for the additional virtual node
kubectl get nodes -l liqo.io/type=virtual-node

# Should see multiple virtual nodes including one from same provider:
# contract-fluidos-eu-k8slice-xxxx-xxxx
# contract-fluidos-eu-k8slice-yyyy-yyyy
```

### Step 7.6: Offload Namespace to Additional Virtual Node

```bash
# Create namespace for testing
kubectl create namespace workload-test-03

# Recover cluster mappings from virtual node
CLUSTER_ID=$(kubectl get node <virtual_node_name> -o jsonpath='{.metadata.labels.liqo\.io/remote-cluster-id}')

# Create NamespaceOffloading
kubectl apply -f - <<EOF
apiVersion: offloading.liqo.io/v1beta1
kind: NamespaceOffloading
metadata:
  name: offloading
  namespace: workload-test-03
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

### Step 7.7: Deploy Workloads to Additional Virtual Node

```bash
# Deploy nginx workload to the additional virtual node
kubectl -n workload-test-03 create deployment nginx-test-03 --image=nginx:1.23 --replicas=2
```

### Step 7.8: Verify Workload Distribution

```bash
# Check pod distribution across virtual nodes
kubectl get pods -n workload-test-03 -o wide
```


## Part 8: Troubleshooting Guide

### Common Issues and Solutions

#### No PeeringCandidates Found

```bash
# Check 1: KnownClusters exist
kubectl get knownclusters -n fluidos

# Check 2: Provider Flavors are available
kubectl --kubeconfig=fluidos-provider-1-config get flavors -n fluidos

# Check 3: REAR Manager is serving
kubectl --kubeconfig=fluidos-provider-1-config logs -n fluidos -l app.kubernetes.io/component=node-rear-manager

# Solution: Create KnownClusters manually for Kind environments
```

#### Solver Stuck in Running State

```bash
# Check Solver events
kubectl describe solver solver-sample -n fluidos

# Check Discovery status
kubectl get discoveries -n fluidos -o yaml | grep -A5 "status:"

# Common cause: Network connectivity issues between clusters
```

#### Pods Pending on Virtual Node

```bash
# Check 1: NamespaceOffloading status
kubectl get namespaceoffloading -n workload-test -o yaml

# Check 2: Virtual node capacity
kubectl describe node $(kubectl get nodes | grep liqo | awk '{print $1}') | grep -A10 "Allocatable:"

# Check 3: Pod configuration
kubectl describe pod <pending-pod> -n workload-test | grep -A10 "Tolerations:"

# Solution: Ensure nodeSelector and tolerations are correctly set
```

#### Allocation Stuck in Peering

```bash
# Check ForeignCluster status
kubectl get foreignclusters -o yaml | grep -A10 "status:"

# Check Liqo controller logs
kubectl logs -n liqo -l app.kubernetes.io/name=liqo-controller-manager

# Common cause: Leftover resources from previous runs
```

## Part 9: Quick Reference Commands

### Essential Commands by Phase

**Discovery Phase:**

```bash
kubectl apply -f solver.yaml
kubectl get solver,discovery,peeringcandidates -n fluidos
```

**Reservation Phase:**

```bash
kubectl patch solver pods-solver -n fluidos --type merge -p '{"spec":{"reserveAndBuy":true}}'
kubectl get reservation,contract,transaction -n fluidos
```

**Peering Phase:**

```bash
kubectl patch solver pods-solver -n fluidos --type merge -p '{"spec":{"establishPeering":true}}'
kubectl get allocation -n fluidos
kubectl get foreignclusters
kubectl get nodes | grep liqo
```

**Workload Deployment:**

```bash
kubectl create namespace workload-test
kubectl apply -f namespace-offloading.yaml
kubectl -n workload-test create deployment nginx-test --image=nginx:1.23 --replicas=2
kubectl get pods -n workload-test -o wide
```

### Cleanup Commands

```bash
# Delete resources in dependency order

# Set kubeconfig to consumer cluster
export KUBECONFIG=fluidos-consumer-1-config

# 1. Delete workload namespaces (contains pods using virtual nodes)
kubectl delete namespace workload-test workload-test-02 workload-test-03 --ignore-not-found

# 2. Delete Solvers (orchestrates Discovery, Reservation, and Allocation)
kubectl delete solver --all -n fluidos

# 3. Delete Allocations (references Contracts, manages Liqo peering)
kubectl delete allocation --all -n fluidos

# 4. Delete Contracts (created from successful Reservations)
kubectl delete contract --all -n fluidos

# 5. Delete Reservations (creates Contracts and Transactions)
kubectl delete reservation --all -n fluidos

# 6. Delete Transactions (tracks reservation process)
kubectl delete transaction --all -n fluidos

# 7. Delete PeeringCandidates (created by Discovery from Flavors)
kubectl delete peeringcandidate --all -n fluidos

# 8. Delete Discoveries (created by Solver to find resources)
kubectl delete discovery --all -n fluidos

# 9. Optional: Clean up Liqo resources if needed
# kubectl delete foreignclusters --all
# kubectl delete resourceslices --all -A

# On provider clusters
# export KUBECONFIG=fluidos-provider-1-config
# kubectl delete contract --all -n fluidos
# kubectl delete transaction --all -n fluidos
# kubectl delete reservation --all -n fluidos
# 
# export KUBECONFIG=fluidos-provider-2-config
# kubectl delete contract --all -n fluidos
# kubectl delete transaction --all -n fluidos
# kubectl delete reservation --all -n fluidos

# Note: Resources NOT deleted (managed by platform):
# - Flavors (provider-managed, represents available capacity)
# - KnownClusters (network discovery, auto-recreated)
# - Brokers (network configuration, platform-managed)

```

## FLUIDOS Controller Summary

### FLUIDOS Resources

| Resource | Controller | Component | Description |
|----------|------------|-----------|-------------|
| **Flavor** | `NodeReconciler` & `ServiceBlueprintReconciler` | local-resource-manager | Creates Flavors from labeled nodes and ServiceBlueprints |
| **Solver** | `SolverReconciler` | rear-manager | Orchestrates entire resource acquisition workflow |
| **Discovery** | `DiscoveryReconciler` | rear-controller | Handles resource discovery across clusters |
| **PeeringCandidate** | Managed by `SolverReconciler` | rear-manager | created during discovery |
| **Reservation** | `ReservationReconciler` | rear-controller | Manages resource reservation requests |
| **Contract** | Created by `ReservationReconciler` | rear-controller | result of successful reservation |
| **Transaction** | Created by `ReservationReconciler` | rear-controller | tracks reservation process |
| **Allocation** | `AllocationReconciler` | rear-manager | Manages resource allocation and Liqo peering |
| **KnownCluster** | Managed by Network Manager | network-manager | Created via network discovery |

### Liqo Resources

| Resource | Controller | Component | Description |
|----------|------------|-----------|-------------|
| **ForeignCluster** | `ForeignClusterReconciler` | liqo-controller-manager | Manages cross-cluster peering connections |
| **NamespaceOffloading** | `NamespaceOffloadingReconciler` | liqo-controller-manager | Controls workload placement on virtual nodes |

### Component Responsibilities

| Component | Role | Resources Managed |
|-----------|------|-------------------|
| **local-resource-manager** | Provider-side resource advertisement | Flavors |
| **rear-manager** | Consumer-side orchestration | Solver, Allocation, PeeringCandidates |
| **rear-controller** | Provider-side REAR protocol | Discovery, Reservation, Contract, Transaction |
| **network-manager** | Cluster discovery and networking | KnownClusters |
| **liqo-controller-manager** | Cross-cluster connectivity | ForeignCluster, NamespaceOffloading |

## Resource Flow Overview - The Foundation for FLARE

Summarize the core concepts and see how they form the foundation for FLARE's GPU federation platform:

- **Provider Side**: `Worker Node → Flavor → REAR Advertisement`
- **Consumer Side**: `Solver → Discovery → Contract → Allocation → Virtual Node`

**Complete Resource Lifecycle**:

```
FLUIDOS Core Workflow (Current)
├── Solver (lifecycle controller)
│   ├── Discovery → PeeringCandidates (resource matching)
│   ├── Reservation → Contract (with resource specifications)
│   └── Allocation → ForeignCluster → Virtual Node
│
Workload Deployment
└── NamespaceOffloading → Remote Execution
```

**What We've Learned in This Tutorial:**

1. **Flavors** advertise what providers offer (CPU, memory, pods)
2. **Solvers** orchestrate the entire consumer workflow through three phases
3. **Discoveries** find matching resources across the federation
4. **Contracts** establish binding agreements with Liqo credentials
5. **Allocations** trigger actual infrastructure creation
6. **Virtual Nodes** make remote resources appear local
7. **NamespaceOffloading** controls where pods run
8. **Multiple Virtual Nodes** enable precise resource targeting from different providers

**FLARE Extensions**:

When FLARE adds GPU support, these same patterns will be extended:
- **GPU-Enhanced Flavors**: Include GPU specifications from node annotations
- **GPU Filtering**: Solvers will match specific GPU requirements  
- **Virtual GPU Nodes**: Remote GPU resources will appear as specialized local nodes
- **Intent Translation**: High-level user goals will become GPU-aware Solver specifications

This basic FLUIDOS workflow provides the foundation for FLARE's GPU federation platform, where the same resource sharing concepts enable transparent GPU access across cloud providers!