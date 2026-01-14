<p align="center">
  <img src="assets/idp-logo.svg" alt="indev logo" width="100">
  <h1 align="center">indev</h1>
</p>

<p align="center">
  A command-line interface for managing Intility Developer Platform resources.
</p>

<p align="center">
  <a href="https://github.com/intility/indev/releases"><img src="https://img.shields.io/github/v/release/intility/indev" alt="GitHub Release"></a>
<img src="https://img.shields.io/badge/go-%3E%3D1.25-blue" alt="Go Version">
</p>

<p align="center">
<a href="https://github.com/intility/indev/actions/workflows/ci.yml"><img src="https://github.com/intility/indev/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build Status"></a>
</p>

---

## Installation

### Homebrew (macOS)

```sh
brew install intility/tap/indev
```

### Windows / Linux

Download the latest release for your platform from the [GitHub Releases](https://github.com/intility/indev/releases) page. Extract the archive and add the binary to your PATH.

### From Source

Requires Go 1.25 or higher.

```sh
git clone https://github.com/intility/indev.git
cd indev
go build -o indev ./cmd/indev
```

## Prerequisites

Some commands require additional tools to be installed:

| Command | Requirement |
|---------|-------------|
| `indev cluster login` | [OpenShift CLI (oc)](https://developers.intility.com/docs/getting-started/first-steps/deploy-first-application/?h=oc#install-openshift-cli) |

## Usage

### Authentication

Log in to the Intility Developer Platform:

```sh
indev login
```

Log out:

```sh
indev logout
```

View your account information:

```sh
indev account show
```

### Cluster Management

Create a new Kubernetes cluster:

```sh
indev cluster create
```

You can also specify options directly:

```sh
indev cluster create --name my-cluster --preset balanced --nodes 4
```

With autoscaling enabled:

```sh
indev cluster create --name my-cluster --preset performance --enable-autoscaling --min-nodes 2 --max-nodes 6
```

Available node presets: `minimal`, `balanced`, `performance`

List your clusters:

```sh
indev cluster list
```

Get details for a specific cluster:

```sh
indev cluster get --name <cluster-name>
```

Check cluster status:

```sh
indev cluster status --name <cluster-name>
```

Log in to a cluster (requires `oc`):

```sh
indev cluster login --name <cluster-name>
```

Open the cluster in the web console:

```sh
indev cluster open --name <cluster-name>
```

Delete a cluster:

```sh
indev cluster delete --name <cluster-name>
```

### Team Management

List teams:

```sh
indev team list
```

Create a new team:

```sh
indev team create --name <team-name>
```

Add a member to a team:

```sh
indev team member add --team <team-name> --user <user-email>
```

Remove a member from a team:

```sh
indev team member remove --team <team-name> --user <user-email>
```

### User Management

List users:

```sh
indev user list
```

### Shell Completions

Shell completions are installed automatically via Homebrew. For manual installation, run `indev completion --help` for instructions.

## Telemetry

`indev` collects anonymous usage data to help improve the tool. This includes command usage, performance metrics, and error reports. No personally identifiable information is collected.

To opt out, set the environment variable:

```sh
export DO_NOT_TRACK=1
```

## Contributing

This project is not currently accepting external contributions. However, we welcome feedback from the community. If you encounter a bug or have a feature request, please [open an issue](https://github.com/intility/indev/issues).
