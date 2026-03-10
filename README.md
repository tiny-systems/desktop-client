# TinySystems Desktop Client

A desktop application for building and managing Kubernetes-native workflows visually.

<!-- Screenshot placeholder -->

## What it does

- **Visual flow editor** -- drag-and-drop nodes onto a canvas, connect them with edges, and configure each node through JSON schema-driven forms.
- **Module browser** -- discover, install, and update TinySystems modules directly from the app using Helm.
- **Project and flow management** -- create projects, organize flows, and manage widget dashboards for control ports.
- **Runs locally** -- connects to your Kubernetes cluster via kubeconfig. No cloud account required.

## Prerequisites

- A running **Kubernetes cluster** (local or remote)
- **kubectl** configured with a valid kubeconfig
- **Helm 3** (recommended -- the app uses Helm to install modules)

## Download

Pre-built binaries for macOS, Windows, and Linux are available on the
[GitHub Releases](https://github.com/tiny-systems/desktop-client/releases) page.

Alternatively, visit [tinysystems.io/download](https://tinysystems.io/download).

## Build from source

Requires [Go 1.24+](https://go.dev/dl/) and [Wails v2](https://wails.io/docs/gettingstarted/installation).

```bash
wails build
```

The output binary is placed in `build/bin/`.

## Development

```bash
wails dev
```

This starts a Vite dev server with hot reload for the Vue 3 frontend. The Go backend rebuilds automatically on changes.

## How it works

The desktop client reads your kubeconfig and talks to the Kubernetes API directly -- the same way kubectl does. TinySystems modules are packaged as Helm charts; installing a module deploys it as a pod in your cluster. Each node in a flow corresponds to a `TinyNode` custom resource. The client watches these resources and renders the flow graph in real time using Vue Flow.

## Related repositories

| Repository | Description |
|---|---|
| [module](https://github.com/tiny-systems/module) | Operator SDK for building TinySystems modules |
| [common-module](https://github.com/tiny-systems/common-module) | Core components (cron, inject, signal, ticker, etc.) |
| [http-module](https://github.com/tiny-systems/http-module) | HTTP server and client components |
| [kubernetes-module](https://github.com/tiny-systems/kubernetes-module) | Kubernetes resource management components |
| [communication-module](https://github.com/tiny-systems/communication-module) | Email and messaging components |
| [encoding-module](https://github.com/tiny-systems/encoding-module) | JSON, Base64, and other encoding components |

## Links

- [Documentation](https://docs.tinysystems.io)
- [Cloud Platform](https://tinysystems.io)
- [Community Slack](https://tinysystems.io/slack)

## License

Apache 2.0
