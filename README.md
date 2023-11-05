# K8s Secret Replicator

K8s Secret Replicator is used to replicate one source secret to all namespaces in a Kubernetes cluster.

This is useful for example for replicating image pull secret to all namespaces without knowing which namespaces will exist in advance.

When content of source secret is changed, secrets with the same name will also be updated in all namespaces.

## Usage

First create or prepare docker-registry secret. Here is example:
```bash
kubectl create secret \
               docker-registry \
               my-secret-to-replicate \
               --docker-server=https://index.docker.io/v1/ \
               --docker-username=user \
               --docker-password=password
```

### Production

Install via helm chart or use your favorite continuous deployment tool:

```bash
helm upgrade \
     --install \
     --create-namespace \
     --version 0.1.0 \
     --namespace kube-system \
     --set app.sourceSecretName=my-secret-to-replicate \
     --set app.sourceSecretNamespace=kube-system \
     --wait \
     k8s-secret-replicator \
     chart
```

Official image is built and pushed to dockerhub https://hub.docker.com/repository/docker/lirt/k8s-secret-replicator/.

### Development

In production in-cluster config will be consumed. For testing you can set kubeconfig to point to a cluster where you want to test it.

```bash
export KUBECONFIG="~/.kube/configs/my-awesome-cluster.yaml"
export SOURCE_SECRET_NAME=my-secret-to-replicate
export SOURCE_SECRET_NAMESPACE=kube-system
go run main.go
```

## Build

### Go

```bash
go build
```

### Docker

```bash
docker buildx build \
              --platform linux/amd64,linux/arm,linux/arm64 \
              --tag lirt/k8s-secret-replicator:v0.1.0 \
              --no-cache \
              --push \
              .
```
