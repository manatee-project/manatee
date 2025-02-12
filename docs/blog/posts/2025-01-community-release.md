---
date: 2025-01-07
---

# First Community Release of ManaTEE

We are thrilled to announce the first public community release of ManaTEE, an open-source framework for private data analytics.
ManaTEE was introduced as a [key privacy solution](https://developers.tiktok.com/blog/privacygo-data-clean-room-open-source)  for private data collaboration at TikTok, which built [one of its products](https://developers.tiktok.com/doc/vce-getting-started) on top of this solution. The team decided to improve and consolidate the solution by open-sourcing it.
To further its momentum as an open-source private data analytics framework, TikTok has [donated the project](https://developers.tiktok.com/blog/tiktok-open-source-project-donation-manatee) to the Confidential Computing Consortium under the Linux Foundation.
After months of development, testing, and refinement, we’re excited to share this project with the broader community.

## What is New?

In the community release, we are providing the following features:

* [Test deployment in minikube](../../getting-started/minikube.md) without cloud accounts (e.g., GCP)
* Full [tutorial](../../getting-started/tutorials.md) to reproduce the demo

We also worked hard to refactor the code, to make it much more extensible. It now leverages Bazel for hermetic and reproducible builds, and has a basic CI/CD pipeline setup. The project is now ready to get contribution from the community!

## What's Next?

This is just the beginning. There are still many work to be done, such as:

* **Diverse backend support**: ManaTEE currently only supports Google Confidential Space as the TEE backend, but different use cases may need diifferent backend. For example, some may want to use an on-prem TEE cluster, or a different cloud. Some might even want to deploy the system in multiple clouds. 
* **Integrated data pipeline**: One of the big challenge for organizations to share data is to process or filter the data to protect privacy and maintain data compliance. To ensure end-to-end data privacy, the data management should be closely integrated with the framework that consumes the data.
* **Output privacy**: Although TEE provides data privacy during execution, the outputs of the execution needs extra efforts to protect data privacy.
* **Support for confidential GPUs**: Data analytics these days often rely on large AI models requiring hardware accelerators such as GPUs. Now that confidential GPUs are readily available, we are ready to support GPU workloads seemlessly in ManaTEE framework.

We are in the process of forming a Technical Steering Committee (TSC) to govern the project and drive its roadmap. Stay tuned for more updates in future posts.

## Join Us

We’d love your feedback to help shape the future of ManaTEE and private data research framework. 
Please feel free to open issues, contribute code, or suggest ideas on GitHub. Please subscribe to our [mailing list](https://groups.google.com/u/1/g/manatee-project) for updates, too!