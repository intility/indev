
# Minctl

Minctl is a command-line interface (CLI) tool designed for managing Kubernetes resources efficiently. With Minctl, you can perform a variety of operations such as creating and deleting clusters, deploying and managing applications, and more. It is built with simplicity and ease of use in mind, making Kubernetes operations more accessible.

## Features

- Create and delete Kubernetes clusters with ease
- Deploy applications to a Kubernetes cluster
- Delete application resources from a cluster
- Forward ports from Kubernetes deployments
- Generate Kubernetes manifests for deployments, services, Dockerfiles, and network policies

## Getting Started

Before you begin, ensure that you have the following installed:
- Docker
- `kubectl` command-line tool
- Go (1.22.1 or higher)

### Installation

To install Minctl, follow these steps:

1. Clone the repository:

    ```sh
    git clone git@gitlab.intility.com:developer-infrastructure/platform-2.0/minctl.git
    cd minctl
    ```

2. Build from source:

    ```sh
    go build -o minctl main.go
    ```

3. (Optional) Move the binary to a location in your PATH:

    ```sh
    mv minctl /usr/local/bin/minctl
    ```

### Usage

Here are some of the commonly used Minctl commands:

- Create a cluster:
    ```sh
    minctl cluster create
    ```

- List clusters:
    ```sh
    minctl cluster list
    ```

- Delete a cluster:
    ```sh
    minctl cluster delete --name <cluster-name>
    ```

- Deploy an application:
    ```sh
    minctl app deploy --path <app-manifest-path>
    ```

- Delete an application:
    ```sh
    minctl app delete --path <app-manifest-path>
    ```

- Forward a port:
    ```sh
    minctl app port-forward --deployment <deployment-name>
    ```

For a full list of command and options, run `minctl --help`.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Acknowledgments

- The `minctl` team and all contributors
- The Go and Kubernetes communities for their tools and libraries

## Contact

For questions or support, please contact Developer Infrastructure.
