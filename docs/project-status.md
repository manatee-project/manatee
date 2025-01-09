# Project Roadmap

A few necessary components such as data SDK are not included in the open source version.
However, you can still try to reproduce our demo by following [tutorials](getting-started/tutorials.md).

## Feature Status

Many parts of ManaTEE are still under active development.

|                         | Current (Alpha)          | Future                    |
|-------------------------|--------------------------|---------------------------|
| **Users**               | One-Way Data Sharing     | Multi-Way Data Sharing   |
| **Backend**             | Single Backend (Goole Cloud Platform)     | Multiple Backend          |
| **Data Provisioning**   | Manual                   | Automated                 |
| **Policy and Attestation** | Manual                | Automated                 |
| **Compute**             | CPU                      | CPU/GPU                   |

* **Data Provisioning, Policy, and Attestation**: Currently, the data owner is responsible for manually setting all the infrastructure including data and the access control. However, future versions will make this easier by including a generic interface for uploading data and configuring the data access policies based on attestation. 

* **Backend**: We only support a single TEE backend called [Confidential Space](https://cloud.google.com/confidential-computing/confidential-space/docs/confidential-space-overview) provided by Google Cloud Platform (GCP). In the future, it will be extended to support more TEE backends including other cloud providers or native confidential VMs/containers.

* **Compute**: ManaTEE currently does not support confidential GPU or any accelerator-based computation.

## Roadmap

We are currently forming Technical Steering Committee (TSC) for governing the project and driving the roadmap. 
If you're interested in joining the project, please reach out to the team via our [mailing list](manatee-project@googlegroups.com).
