# Tutorials

In this tutorial, we show the following scenario:

* Dataset: [US Health Insurance Dataset from Kaggle](https://www.kaggle.com/datasets/teertha/ushealthinsurancedataset/data)
* Task: Train a model predicting the insurance charge based on the data (use XGBoost regression)
* Data Provisioning: [MST](https://github.com/ryan112358/private-pgm/blob/9fc3a829f344a2ed0b662f4f8d92d5efe18e42ab/mechanisms/mst.py) (2018 NIST snythetic data challenge winner).

<iframe width="854" height="480" src="https://www.youtube.com/embed/Ig1NzZZ6yKE?si=-IjD6v2nqsrAuuuV" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>


## Overal Workflow

This tutorial uses a very simple prototype data SDK to interact with two different versions of datasets as shown below.

![ManaTEE Architecture](../assets/img/arch.png)
/// Caption
Overall ManaTEE Architecture and Demo Workflow
///

As mentioned earlier, data provisioning currently needs to be done manually by the data owner.
The data owner is required to:

* Use proper techniques to generate stage-1 dataset from the original (stage-2) dataset.
* Upload two different data versions to the cloud
* Use proper access control mechanism to both of the data versions

In general, the data access control should be implemented based on remote attestation protocol provided by the TEE instance.
Google Confidential Space [recommends](https://cloud.google.com/docs/security/confidential-space) to use Google KMS and an OpenID Connect (OIDC) token provider.

