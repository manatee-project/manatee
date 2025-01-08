<img width=100% src="docs/assets/img/logo.png">

# ManaTEE Project

ManaTEE is an open-source project for easily building and deploying data collaboration framework to the cloud using trusted execution environments (TEEs).
It allows users to easily collaborate on private datasets without leaking privacy of individual data.
ManaTEE achieves this by combining different privacy-enhancing technologies (PETs) in different stages.

# What does it offer?

ManaTEE allows organizations to quickly customize and deploy data collaboration framework in the cloud.
The organizations can provide an programming environment to the external data scientists to conduct research, while protecting the data privacy with a custom policy.

> Note: ManaTEE is under active development, and it is not production-ready. We are looking forward to your feedback and contributions. 

# Quick Start

Install Bazel with [Bazelisk](https://github.com/bazelbuild/bazelisk):
```sh
brew install bazelisk # on MacOS
choco install bazelisk # on Windows
```
On Ubuntu, download the latest Bazelisk binary via [Releases](https://github.com/bazelbuild/bazelisk/releases)

Build all images
```
bazelisk build //...
```

Run all tests
```
bazelisk test //...
```

See [documents](https://manatee-project.github.io/manatee) for more details including cloud deployment.
# License

ManaTEE is licensed under the Apache License 2.0.
See [LICENSE](LICENSE) for details.