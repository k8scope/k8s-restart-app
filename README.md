# k8s-restart-app

A simple application that allows to restart services running in K8s. The service that should be allowed to be restarted, must be defined in the configuration file.

## Configuration

The configuration of the application is mostly done through environment variables. The following environment variables are available:

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `LISTEN_ADDRESS` | string | `:8080` | The address the application should listen on. |
| `CONFIG_FILE` | string | `config.yaml` | The path to the configuration file. |
| `KUBE_CONFIG` | string | `` | The path to the kubeconfig file. If not specified, the application tries to use the in-cluster config. |

In order to provide a list of services that should be allowed to be restarted, a configuration file must be provided. In that file, the services are defined as follows:

```yaml
services:
  - kind: Deployment # The kind of the service (Deployment, StatefulSet)
    name: my-deployment # The name of the service
    namespace: my-namespace # The namespace the service is running in
```
