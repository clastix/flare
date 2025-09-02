<!-- markdownlint-disable first-line-h1 -->
<p align="center">
<a href="https://fluidos.eu/"><img src="https://github.com/fluidos-project/quick-start-guide/raw/1169710781fc338c977944dafcdd0e0240ae5821/.assets/img/fluidoslogo.png" width="150"/></a>
</p>

# FLARE (Federated Liquid Resources Exchange)

[**FLARE**](https://github.com/clastix/flare) (**F**ederated **L**iquid **R**esources **E**xchange) is a GPU pooling platform for AI and HPC applications. Built on [FLUIDOS](https://fluidos.eu/), it enables dynamic GPU sharing across cloud providers through intent-based allocation and federated multi-tenant cloud architecture.

> **⚠️ Research Project**  
> _This is a research project and is not intended for production use. This project explores concepts in federated GPU resource management and serves as a proof-of-concept for distributed computational resource sharing. If you are interested in using these concepts for your commercial project or want to borrow ideas from this research, please contact [CLASTIX](https://clastix.io)._


## 📚 Documentation

### Getting Started

#### [Project Overview](docs/FLARE_Project_Overview.md)
- [Executive Summary](docs/FLARE_Project_Overview.md#executive-summary)
- [Problem Statement and Solution](docs/FLARE_Project_Overview.md#problem-statement-and-proposed-solution)
- [FLARE Architecture on FLUIDOS](docs/FLARE_Project_Overview.md#flare-architecture-on-fluidos)
- [Technical Differentiation](docs/FLARE_Project_Overview.md#technical-differentiation)
- [Implementation Roadmap](docs/FLARE_Project_Overview.md#implementation-roadmap)
- [Additional Resources](docs/FLARE_Project_Overview.md#additional-resources)

#### [QuickStart Guide](docs/FLARE_QuickStart_Guide.md)
- [Overview](docs/FLARE_QuickStart_Guide.md#overview)
- [Prerequisites](docs/FLARE_QuickStart_Guide.md#prerequisites)
- [Development Environment Setup](docs/FLARE_QuickStart_Guide.md#development-environment-setup)
- [Configuration Summary](docs/FLARE_QuickStart_Guide.md#configuration-summary)
- [Troubleshooting](docs/FLARE_QuickStart_Guide.md#troubleshooting)
- [Cleanup](docs/FLARE_QuickStart_Guide.md#cleanup)
- [Next Steps](docs/FLARE_QuickStart_Guide.md#next-steps)

### Core Concepts

#### [FLUIDOS Basic Workflow](docs/FLUIDOS_Basic_Workflow.md)
- [FLUIDOS Overview](docs/FLUIDOS_Basic_Workflow.md#fluidos-overview)
- [Prerequisites Check](docs/FLUIDOS_Basic_Workflow.md#prerequisites-check)
- [Provider Resources (Flavors)](docs/FLUIDOS_Basic_Workflow.md#part-1-provider-resources-flavors)
- [Consumer Resource Discovery](docs/FLUIDOS_Basic_Workflow.md#part-2-consumer-resource-discovery)
- [Resource Reservation and Contract](docs/FLUIDOS_Basic_Workflow.md#part-3-resource-reservation-and-contract)
- [Establishing Cluster Peering](docs/FLUIDOS_Basic_Workflow.md#part-4-establishing-cluster-peering)
- [Deploying Workloads on Virtual Nodes](docs/FLUIDOS_Basic_Workflow.md#part-5-deploying-workloads-on-virtual-nodes)
- [Multiple Virtual Nodes](docs/FLUIDOS_Basic_Workflow.md#part-6-virtual-node-from-different-provider)
- [Multiple Virtual Nodes from Same Provider](docs/FLUIDOS_Basic_Workflow.md#part-7-multiple-virtual-nodes-from-same-provider)
- [Troubleshooting Guide](docs/FLUIDOS_Basic_Workflow.md#part-8-troubleshooting-guide)
- [Quick Reference Commands](docs/FLUIDOS_Basic_Workflow.md#part-9-quick-reference-commands)

#### [FLARE GPU Pooling Guide](docs/FLARE_GPU_Pooling_Guide.md)
- [Overview](docs/FLARE_GPU_Pooling_Guide.md#overview)
- [Prerequisites](docs/FLARE_GPU_Pooling_Guide.md#prerequisites)
- [Development Timeline Context](docs/FLARE_GPU_Pooling_Guide.md#development-timeline-context)
- [Workflow Evolution](docs/FLARE_GPU_Pooling_Guide.md#workflow-evolution)
  - [Manual Process (Original FLUIDOS)](docs/FLARE_GPU_Pooling_Guide.md#workflow-1-manual-process-original-fluidos)
  - [Semi-Automated (Enhanced FLUIDOS)](docs/FLARE_GPU_Pooling_Guide.md#workflow-2-semi-automated-enhanced-fluidos)
  - [Fully Automated (FLARE)](docs/FLARE_GPU_Pooling_Guide.md#workflow-3-fully-automated-flare)
- [Workflow Comparison](docs/FLARE_GPU_Pooling_Guide.md#workflow-comparison)
- [GPU Annotation Reference](docs/FLARE_GPU_Pooling_Guide.md#gpu-annotation-reference)
- [Use Cases and Examples](docs/FLARE_GPU_Pooling_Guide.md#use-cases-and-examples)
- [Troubleshooting](docs/FLARE_GPU_Pooling_Guide.md#troubleshooting)
- [Cleanup](docs/FLARE_GPU_Pooling_Guide.md#cleanup)

### Architecture & API

#### [FLARE Architecture](docs/FLARE_Architecture.md)
- [Project Overview](docs/FLARE_Architecture.md#project-overview)
- [Overall Architecture](docs/FLARE_Architecture.md#overall-architecture)
- [Key Components](docs/FLARE_Architecture.md#key-components)
  - [FLARE Platform](docs/FLARE_Architecture.md#flare-platform)
  - [FLUIDOS Integration](docs/FLARE_Architecture.md#fluidos-integration)
- [Deployment Architecture](docs/FLARE_Architecture.md#deployment-architecture)
- [Core Workflows](docs/FLARE_Architecture.md#core-workflows)
  - [GPU Provider Setup Flow](docs/FLARE_Architecture.md#1-gpu-provider-setup-flow)
  - [Basic GPU Allocation Flow](docs/FLARE_Architecture.md#2-basic-gpu-allocation-flow)
  - [No GPU Requirements Met Flow](docs/FLARE_Architecture.md#3-no-gpu-requirements-met-flow)
  - [GPU Resource Contention Flow](docs/FLARE_Architecture.md#4-gpu-resource-contention-flow)
  - [GPU Provider Failure Flow](docs/FLARE_Architecture.md#5-gpu-provider-failure-flow)

#### [FLARE API Reference](docs/FLARE_API_Reference.md)
- [Overview](docs/FLARE_API_Reference.md#overview)
- [API Resources](docs/FLARE_API_Reference.md#api-resources)
- [Workload Intent Schema](docs/FLARE_API_Reference.md#workload-intent-schema)
  - [Base Structure](docs/FLARE_API_Reference.md#base-structure)
  - [Workload Specification](docs/FLARE_API_Reference.md#workload-specification)
  - [Constraints](docs/FLARE_API_Reference.md#constraints)
- [Authentication](docs/FLARE_API_Reference.md#authentication)
- [API Endpoints](docs/FLARE_API_Reference.md#api-endpoints)
- [Complete Examples](docs/FLARE_API_Reference.md#complete-examples)
- [Response Formats](docs/FLARE_API_Reference.md#response-formats)
- [Error Codes](docs/FLARE_API_Reference.md#error-codes)

### GPU Resource Management

#### [GPU Annotations Reference](docs/FLARE_GPU_Annotations_Reference.md)
- [Overview](docs/FLARE_GPU_Annotations_Reference.md#overview)
- [Quick Start](docs/FLARE_GPU_Annotations_Reference.md#quick-start)
- [Quick Reference Table](docs/FLARE_GPU_Annotations_Reference.md#quick-reference-table)
- [Detailed Annotation Specifications](docs/FLARE_GPU_Annotations_Reference.md#detailed-annotation-specifications)
- [Core GPU Annotations](docs/FLARE_GPU_Annotations_Reference.md#core-gpu-annotations-required)
- [Location & Cost Annotations](docs/FLARE_GPU_Annotations_Reference.md#location-annotations-required---manual)
- [Performance Annotations](docs/FLARE_GPU_Annotations_Reference.md#performance-annotations-manual---optional)
- [GPU Sharing Annotations](docs/FLARE_GPU_Annotations_Reference.md#gpu-sharing-annotations-manual---optional)
- [Network & Communication Annotations](docs/FLARE_GPU_Annotations_Reference.md#network-performance-annotations-optional)
- [Provider Annotations](docs/FLARE_GPU_Annotations_Reference.md#provider-annotations-optional)
- [Annotation Examples](docs/FLARE_GPU_Annotations_Reference.md#annotation-examples)
- [FLARE API Mapping](docs/FLARE_GPU_Annotations_Reference.md#flare-api-mapping)
- [Validation Rules](docs/FLARE_GPU_Annotations_Reference.md#validation-rules)

#### [NVIDIA GPU Labels Mapping](docs/NVIDIA_GPU_Labels_Mapping.md)
- [Overview](docs/NVIDIA_GPU_Labels_Mapping.md#overview)
- [NVIDIA GPU Operator Label Reference](docs/NVIDIA_GPU_Labels_Mapping.md#nvidia-gpu-operator-label-reference)
- [Direct Label to Annotation Mappings](docs/NVIDIA_GPU_Labels_Mapping.md#direct-label-to-annotation-mappings)
- [Computed Annotation Mappings](docs/NVIDIA_GPU_Labels_Mapping.md#computed-annotation-mappings)
- [Complete Auto-Generation Example](docs/NVIDIA_GPU_Labels_Mapping.md#complete-auto-generation-example)
- [NVIDIA-Specific Extensions](docs/NVIDIA_GPU_Labels_Mapping.md#nvidia-specific-extensions)
- [Node Selection](docs/NVIDIA_GPU_Labels_Mapping.md#node-selection)

#### [AMD GPU Labels Mapping](docs/AMD_GPU_Labels_Mapping.md)
- [Overview](docs/AMD_GPU_Labels_Mapping.md#overview)
- [AMD GPU Operator Label Reference](docs/AMD_GPU_Labels_Mapping.md#amd-gpu-operator-label-reference)
- [Direct Label to Annotation Mappings](docs/AMD_GPU_Labels_Mapping.md#direct-label-to-annotation-mappings)
- [Computed Annotation Mappings](docs/AMD_GPU_Labels_Mapping.md#computed-annotation-mappings)
- [Complete Auto-Generation Example](docs/AMD_GPU_Labels_Mapping.md#complete-auto-generation-example)
- [AMD-Specific Extensions](docs/AMD_GPU_Labels_Mapping.md#amd-specific-extensions)
- [Node Selection](docs/AMD_GPU_Labels_Mapping.md#node-selection)

### Operations & Administration

#### [Admin Guide](docs/FLARE_Admin_Guide.md)
- [Overview](docs/FLARE_Admin_Guide.md#overview)
- [Prerequisites](docs/FLARE_Admin_Guide.md#prerequisites)
- [Hub Cluster Setup](docs/FLARE_Admin_Guide.md#hub-cluster-setup-flare-consumer)
- [GPU Provider Setup](docs/FLARE_Admin_Guide.md#gpu-provider-setup)
- [Broker Requirements](docs/FLARE_Admin_Guide.md#broker-requirements)
- [Verification and Testing](docs/FLARE_Admin_Guide.md#verification)
- [Monitoring](docs/FLARE_Admin_Guide.md#monitoring)
- [Troubleshooting](docs/FLARE_Admin_Guide.md#troubleshooting)

### Use Cases & Examples

#### [Sample Use Cases](docs/FLARE_Sample_Use_Cases.md)
- [AI Inference Service](docs/FLARE_Sample_Use_Cases.md#1-ai-inference-service)
- [High-Performance AI Training](docs/FLARE_Sample_Use_Cases.md#2-high-performance-ai-training)
- [LLM Fine-Tuning](docs/FLARE_Sample_Use_Cases.md#3-llm-fine-tuning)
- [High-Performance Computing](docs/FLARE_Sample_Use_Cases.md#4-high-performance-computing)
- [Real-Time Video Analytics](docs/FLARE_Sample_Use_Cases.md#5-real-time-video-analytics)
- [Edge Inference](docs/FLARE_Sample_Use_Cases.md#6-edge-inference)
- [Batch Processing](docs/FLARE_Sample_Use_Cases.md#7-batch-processing)
- [Multi-Tenant Resources](docs/FLARE_Sample_Use_Cases.md#8-multi-tenant-resources)
- [Distributed Workloads](docs/FLARE_Sample_Use_Cases.md#9-distributed-workloads)
- [Cost Optimization Scenarios](docs/FLARE_Sample_Use_Cases.md#10-cost-optimization-scenarios)

### Project Documentation

#### [Final Project Review](docs/FLARE_Final_Project_Review.md)
- [Executive Summary](docs/FLARE_Final_Project_Review.md#executive-summary)
- [Project Timeline and Deliverables](docs/FLARE_Final_Project_Review.md#project-timeline-and-deliverables)
- [Results Achieved Against KPIs and Milestones](docs/FLARE_Final_Project_Review.md#results-achieved-against-kpis-and-milestones)
- [Technical Innovations and Architecture](docs/FLARE_Final_Project_Review.md#technical-innovations-and-architecture)
- [Project Challenges and Technical Solutions](docs/FLARE_Final_Project_Review.md#project-challenges-and-technical-solutions)
- [Project Conclusion and Assessment](docs/FLARE_Final_Project_Review.md#project-conclusion-and-assessment)


## Licensing

- Documentation License: [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International (CC BY-NC-SA 4.0)](https://creativecommons.org/licenses/by-nc-sa/4.0/)
- Source Code License: [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)

## Additional Resources

- **[FLUIDOS Project](https://github.com/fluidos-project)** - Base infrastructure platform
- **[Liqo Project](https://liqo.io)** - Multi-cluster connectivity layer
- **[Fake GPU Operator](https://github.com/run-ai/fake-gpu-operator)** - GPU simulation for development and testing
- **[Capsule Project](https://projectcapsule.dev)** - Multi-tenancy solution for Kubernetes