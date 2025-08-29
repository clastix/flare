
# Efficient GPU Management with FLARE

## Algorithm for Optimized Allocation of AI workloads on top of FLARE

---

### Table of Contents

- Introduction .......................................................................................... 3
- Scenario and Architecture .................................................................. 3
- Scheduling Algorithm ........................................................................ 5
- Optimal Solution .................................................................................. 7
- Heuristic Solution ............................................................................... 8
- Performance Evaluation .................................................................... 11
  - Metrics .............................................................................................. 11
  - Benchmark Scheme ....................................................................... 11
  - Simulation Results .......................................................................... 12
  - Demo ................................................................................................ 18
- Conclusion .......................................................................................... 21

---

## Introduction

This document presents the algorithm developed for the allocation of AI workloads in multi-cluster environments provided by FLARE. It analyzes performance against a benchmark method using both numerical simulations to validate theoretical correctness and a demo in a FLARE emulated environment.

---

## Scenario and Architecture

The allocation algorithm is designed to operate in a FLARE based environment, as illustrated in
**Figure 1 (p.3)**, where GPU resources are distributed among different Fluidos nodes (FNs) across multiple clusters.

![Figure 1 Placeholder](p3)

The algorithm relies on the GPU pooling system provided by FLARE, which aggregates available GPU resources across clusters.
This approach enables the entire infrastructure to be managed as a unified GPU cluster. Leveraging
this global view, the algorithm computes an optimal allocation for AI applications, meeting GPU
resource requirements while minimizing rental costs and ensuring adequate quality of service, based on
hardware topology and network latency of selected nodes. The allocation process begins with
gathering the infrastructure status and GPU requirements of the AI applications.

**Figure 2: Algorithm interaction with the reference architecture (p.4)**

![Figure 2 Placeholder](p4)

The key steps to provide the algorithm with the necessary data are as follows:

1. **Application requests**: Each AI application may request a variable number of GPUs to
ensure adequate inference or training performance. Applications are assumed to consist of a set of
interdependent K8s pods, each containing a shard of the neural network model. Each pod requires
1 GPU and may specify minimum requirements for memory and/or GPU model. It is assumed
that applications do not impose specific RAM or CPU requirements, or that these are minimal and
always satisfied by the selected node, thus negligible during algorithm design.

2. **Infrastructure virtualization**: The multi-cluster infrastructure is dynamically virtualized
through an aggregation process, enabling access to GPU resources regardless of the cluster they
belong to.

3. **Infrastructure monitoring**: The platform constantly monitors GPU availability and
collects information on network topology (intra- and inter-cluster latency) and hardware topology
of the nodes (type of GPU interconnect).

4. **Allocation computation**: The scheduling algorithm processes this information to
determine the optimal pod-to-node allocation.

---

## Scheduling Algorithm

The proposed algorithm computes the optimal pod allocation on a configuration of nodes, aiming to
minimize two conflicting objectives:

1. **Deployment cost**: The monetary cost of renting the required GPUs. This depends on
factors such as GPU model, hardware architecture, and geographic region of the node. Minimizing
this cost yields the most cost-effective solution.

2. **Communication cost**: The performance loss caused by network latency between nodes
and/or non-optimal GPU interconnections. Minimizing this cost leads to configurations with
theoretically best service quality, thanks to lower inter-pod communication latency.

For deployment cost, this is straightforward to model since cloud providers publish hourly costs
for each GPU type. Communication cost depends on inter-node latency (if the application is
distributed across nodes) and the type of GPU interconnect (e.g., NVLink vs PCIe). A qualitative
model is adopted, assigning increasing costs based on theoretical QoS impact, as reported in
**Table 1 (p.6)**.

---

### Optimal Solution

The objective function comprises two components:

1. The monetary cost for renting GPUs, modulated by communication cost due to GPU
interconnection type.
2. The communication cost due to multi-node allocations.

Constraints ensure allocations respect GPU count, memory limits, compatibility, and full pod
assignment.

The optimal solution is computed using **SCIP** (https://www.scipopt.org/), a linear programming
solver. High variable counts pose convergence challenges in large-scale scenarios, motivating
development of a heuristic alternative.

---

### Heuristic Solution

The heuristic uses a greedy mechanism to minimize monetary cost while accounting for communication
overhead. Two solutions are compared: single-node allocation prioritizing NVLink nodes and multi-node
allocation prioritizing cheapest available GPUs. The final allocation is chosen based on cumulative cost
(deployment + communication).

Pseudo-code:

```python
Function SORT_NODES_BY_COST():
    Sort nodes by (cost, -gpu_count)
    Return ordered list of node_ids

Function FIND_SINGLE_NODE_ALLOCATION(sorted_nodes):
    For each workload w in workloads:
        For each node in sorted_nodes:
            If node has enough GPUs and single GPU memory:
                Assign all pods of w to this node
                Update node’s available GPU and memory
                Record allocation and break
    Return allocations

Function FIND_MULTI_NODE_ALLOCATION(sorted_nodes):
    For each workload w in workloads:
        gpus_needed ← w.gpu
        For each node in sorted_nodes:
            While gpus_needed <= node GPUs AND w fits single GPU memory:
                Assign 1 GPU to w on this node
                Update node resources
                Decrease gpus_needed
        Record allocation
    Return allocations

Function EVALUATE_COMMUNICATION_COST(allocation):
    total_cost ← 0
    For each workload w:
        For each node used in w:
            total_cost += node_cost + interconnect_cost
        For each node pair (i, j) in w where i ≥ j:
            total_cost += net_topology[i][j] * inter_node_cost
    Return total_cost
```

Main function:

```python
Function HEURISTIC_COMMUNICATION_AWARE(workloads, nodes, net_topology):
    sorted_nodes ← SORT_NODES_BY_COST()
    single_alloc ← FIND_SINGLE_NODE_ALLOCATION(sorted_nodes)
    multi_alloc ← FIND_MULTI_NODE_ALLOCATION(sorted_nodes)
    cost_single ← EVALUATE_COMMUNICATION_COST(single_alloc)
    cost_multi ← EVALUATE_COMMUNICATION_COST(multi_alloc)
    If cost_single ≤ cost_multi:
        Use single_alloc as final allocation
    Else:
        Use multi_alloc as final allocation
    Return final allocation
```

---

## Performance Evaluation

### Metrics

- **Monetary cost** (GPU rental cost/hour)  
- **GPU interconnect types** (PCIe vs NVLink)  
- **Inter-node allocation types** (intra-cluster vs inter-cluster)  
- **QoS indicator** (score 0–1 measuring impact of communication cost)

### Benchmark Scheme

The **Cheapest GPU First (CGF)** benchmark ignores communication costs and selects only least
expensive nodes. Comparison with CA_OPT and CA_HEU shows the importance of balancing
communication and monetary costs.

### Simulation Results

- Deployment cost trends (**Figure 3, p.12**)  
- NVLink vs PCIe node usage (**Figures 4 & 5, p.13**)  
- Multi-node allocation analysis (**Figure 6, p.14**)  
- QoS performance comparison (**Figure 7, p.15**)  
- Deployment vs QoS trade-off (**Figure 8, p.16**)  
- Computational complexity (**Figure 9, p.17**)

---

## Demo

The demo testbed was implemented using a FLARE setup simultaed with KinD. The architecture emulates GPU availability and validates algorithm
integration with Kubernetes native scheduling.

Resulting allocation using CA_OPT is shown in **Figure 11 (p.19)**.

---

## Conclusion

The project tackled GPU optimization in heterogeneous multi-cluster infrastructures, aiming to reduce
deployment costs while maintaining AI service performance. Two algorithms were proposed:

- **Optimal**: Best allocation but computationally expensive.  
- **Heuristic**: Near-optimal with reduced complexity, suited for large-scale deployments.

Simulations confirmed improvements over cost-only benchmarks. A Kubernetes-based demo
demonstrated real-world feasibility. Both approaches balance deployment and communication
costs, outperforming simpler strategies in QoS.

---

---
## Table of Figures

- **Figure 1**: Considered multi-cluster scenario (p.3) – ![Placeholder](p3)
- **Figure 2**: Algorithm interaction with the reference architecture (p.4) – ![Placeholder](p4)
- **Figure 3**: Deployment cost trends (p.12) – ![Placeholder](p12)
- **Figure 4**: Number of nodes with NVLink interconnections (p.13) – ![Placeholder](p13)
- **Figure 5**: Number of nodes with PCIe interconnections (p.13) – ![Placeholder](p13)
- **Figure 6**: Multi-node allocation within same cluster (p.14) – ![Placeholder](p14)
- **Figure 7**: QoS performance comparison (p.15) – ![Placeholder](p15)
- **Figure 8**: Deployment vs QoS trade-off (p.16) – ![Placeholder](p16)
- **Figure 9**: Computational complexity comparison (p.17) – ![Placeholder](p17)
- **Figure 10**: Testbed architecture for the demo (p.18) – ![Placeholder](p18)
- **Figure 11**: Allocation result from demo (p.19) – ![Placeholder](p19)
