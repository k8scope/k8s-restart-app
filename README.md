# k8s-restart-app

A simple application that allows to restart services running in K8s. The service that should be allowed to be restarted, must be defined in the configuration file.

## UI

The application provides a simple UI to restart services. The UI is available at the root path of the application.

![UI](docs/ui.png)

## Configuration

The configuration of the application is mostly done through environment variables. The following environment variables are available:

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `LISTEN_ADDRESS` | string | `:8080` | The address the application should listen on. |
| `CONFIG_FILE_PATH` | string | `config.yaml` | The path to the configuration file. |
| `KUBE_CONFIG_PATH` | string | `` | The path to the kubeconfig file. If not specified, the application tries to use the in-cluster config. |
| `WATCH_INTERVAL` | int | `10` | The interval in seconds the application watches for pod, deployment or statefulset changes |

In order to provide a list of services that should be allowed to be restarted, a configuration file must be provided. In that file, the services are defined as follows:

```yaml
services:
  - kind: Deployment # The kind of the service (Deployment, StatefulSet)
    name: my-deployment # The name of the service
    namespace: my-namespace # The namespace the service is running in
```

## API

The application provides a simple API to restart services. The following endpoints are available:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Returns the HTML control page. |
| `/api/v1/service` | GET | Returns a list of services that can be restarted. |
| `/api/v1/service/{kind}/{namespace}/{name}/restart` | POST | Restarts the service with the given kind, namespace and name. |
| `/api/v1/service/{kind}/{namespace}/{name}/status` | GET | Returns the status of the service with the given kind, namespace and name. As websocket stream. |
