---
date: 2024-01-31
---

# First Community Release of ManaTEE

We are thrilled to announce the first public community release of ManaTEE, an open-source framework for private data analytics. After months of development, testing, and refinement, we’re excited to share this project with the broader community.

## What is New?

In the community release, we are providing the following features:

* Deployment in test environment (e.g., Minikube) without cloud accounts (e.g., GCP)
* Full [tutorial](../../getting-started/tutorials.md) to reproduce the demo

We also worked hard to refactor the code, to make it much more extensible. It now leverages Bazel for hermetic and reproducible builds, and has a basic CI/CD pipeline setup. The project is now ready to get contribution from the community!

## What's Next?

This is just the beginning! Here’s what we’re planning for future releases (subject to change):

* Bring-your-own-TEE deployment
* Automated data pipeline
* Output privacy
* Support for confidential GPUs

We are also in the process of forming a Technical Steering Committee (TSC) to govern the project and drive its roadmap. Stay tuned for more updates in future posts!

## Join Us

We’d love your feedback to help shape the future of ManaTEE. 
Please feel free to open issues, contribute code, or suggest ideas on GitHub. Please subscribe to our [mailing list](https://groups.google.com/u/1/g/manatee-project) for updates, too!