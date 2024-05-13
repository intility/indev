
# ICPCTL

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
    git clone git@gitlab.intility.com:developer-infrastructure/platform-2.0/icpctl.git
    cd icpctl
    ```

2. Build from source:

    ```sh
    go build -o icpctl main.go
    ```

3. (Optional) Move the binary to a location in your PATH:

    ```sh
    mv icpctl /usr/local/bin/icpctl
    ```

### Usage

Here are some of the commonly used Minctl commands:

- Create a cluster:
    ```sh
    icpctl cluster create
    ```

- List clusters:
    ```sh
    icpctl cluster list
    ```

- Delete a cluster:
    ```sh
    icpctl cluster delete --name <cluster-name>
    ```

- Deploy an application:
    ```sh
    icpctl app deploy --path <app-manifest-path>
    ```

- Delete an application:
    ```sh
    icpctl app delete --path <app-manifest-path>
    ```

- Forward a port:
    ```sh
    icpctl app port-forward --deployment <deployment-name>
    ```

For a full list of command and options, run `icpctl --help`.

## Telemetry

Minctl includes a telemetry feature that helps improve the tool by collecting anonymous usage data. 
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

Telemetry is enabled by default to help us improve Minctl. However, respecting user choice is paramount, 
and you can opt-out of telemetry at any time. To disable telemetry, set the environment variable 
`DO_NOT_TRACK` to `1`.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Acknowledgments

- The `icpctl` team and all contributors
- The Go and Kubernetes communities for their tools and libraries

## Contact

For questions or support, please contact Developer Infrastructure.
