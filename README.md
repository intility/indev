
# IDPCTL

`idpctl` is a command-line interface (CLI) tool designed for managing Intility Developer Platform resources. With `idpctl`, you can perform a variety of operations such as creating and deleting clusters, deploying and managing applications, and more. It is built with simplicity and ease of use in mind, making Kubernetes operations more accessible.

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

To install `idpctl`, follow these steps:

1. Clone the repository:

    ```sh
    git clone git@gitlab.intility.com:developer-infrastructure/platform-2.0/idpctl.git
    cd idpctl
    ```

2. Build from source:

    ```sh
    go build -o idpctl main.go
    ```

3. (Optional) Move the binary to a location in your PATH:

    ```sh
    mv idpctl /usr/local/bin/idpctl
    ```

### Usage

Here are some of the commonly used `idpctl` commands:

- Create a cluster:
    ```sh
    idpctl cluster create
    ```

- List clusters:
    ```sh
    idpctl cluster list
    ```

- Delete a cluster:
    ```sh
    idpctl cluster delete --name <cluster-name>
    ```

- Deploy an application:
    ```sh
    idpctl app deploy --path <app-manifest-path>
    ```

- Delete an application:
    ```sh
    idpctl app delete --path <app-manifest-path>
    ```

- Forward a port:
    ```sh
    idpctl app port-forward --deployment <deployment-name>
    ```

For a full list of command and options, run `idpctl --help`.

## Telemetry

`idpctl` includes a telemetry feature that helps improve the tool by collecting anonymous usage data. 
The telemetry system gathers information such as command usage, performance metrics, and error reports. 
This data is crucial for identifying common issues, understanding user behavior, and prioritizing new features.

### What We Collect

- Command usage: Which commands are run, along with flag usage.
- Performance metrics: Response times for commands and other performance-related metrics.
- Error reports: Unhandled errors or exceptions that occur during the use of the tool.

### Anonymity and Privacy

We are fully committed to ensuring user privacy and anonymity. 
The telemetry system only collects non-personally identifiable information. 
Additionally, data is stored securely and in compliance with relevant data protection regulations.

### Opting Out

Telemetry is enabled by default to help us improve `idpctl`. However, respecting user choice is paramount, 
and you can opt-out of telemetry at any time. To disable telemetry, set the environment variable 
`DO_NOT_TRACK` to `1`.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Acknowledgments

- The `idpctl` team and all contributors
- The Go and Kubernetes communities for their tools and libraries

## Contact

For questions or support, please contact Developer Infrastructure.
