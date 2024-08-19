# <img title="SSM Injector" src="./docs/images/Kubernetes.png" style="width: 24px;">  Kubernetes:  <img title="SSM Injector" src="./docs/images/AWS-SSM.png" style="margin-left: 5px; position: relative; width: 24px;"> AWS SSM Parameter Injector

A Kubernetes admissions controller which searches specific resources for SSM Parameter keys and, when found, retrieves the values for those parameters and injects them into the relevant requests.

## TL;DR

`helm install my-release oci://public.ecr.aws/aedificans/ssm-param-injector`

## Getting Started

### Prerequisites
- `go` version `v1.22.0+`,
- `docker` version `17.03+`.
- `helm` version `v3.13.0+`.
- `kubectl` version `v1.11.3+`.
- Access to a Kubernetes `v1.20.0+` cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/ssm-param-injector-webhook:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Deploy the `MutatingWebhookConfiguration` to the cluster:**

```sh
make helm-install IMAGE_NAME=<some-registry>/ssm-param-injector-webhook IMAGE_TAG=tag 
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**UnDeploy the `MutatingWebhookConfiguration` from the cluster:**

```sh
make helm-uninstall
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/k8s-ssm-param-webhook:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/k8s-ssm-param-injector/<tag or branch>/dist/install.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

