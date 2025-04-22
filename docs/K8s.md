# Deployment Instructions
This directory contains deployment configurations for the trading application.

## Kubernetes Deployment

The `k8s` directory contains all necessary Kubernetes manifests to deploy the application stack.

### Prerequisites

- Kubernetes cluster (Minikube, kind, or any other Kubernetes cluster)

- kubectl CLI installed and configured

- Docker

### Building the Docker Image

1. Navigate to the project root:

```bash

cd  trading

```

### Building the Docker Image

1. Navigate to the project root:

```bash

cd  trading

```

2. Build the Docker image:

```bash

docker  build  -t  trading:local  .

```

3. If using Minikube, load the image into Minikube:

```bash

minikube  image  load  trading:local

```

Alternatively, you can point your Docker client to Minikube's Docker daemon before building:

```bash

eval $(minikube  docker-env)

docker  build  -t  trading:local  .

```

### Deploying the Application

1. Apply the ConfigMap:

```bash

kubectl  apply  -f  deployments/k8s/app-config.yaml

```

2. Deploy Database:

```bash

kubectl  apply  -f  deployments/k8s/postgres-deployment.yaml

```
3. Deploy the trading application:

```bash

kubectl  apply  -f  deployments/k8s/app-deployment.yaml

```

4. Verify all components are running:

```bash

kubectl  get  deployments

kubectl  get  pods

kubectl  get  services

```

### Accessing the Application

The application exposes two ports:

- HTTP service on port 8080

- gRPC service on port 50057

Since the services are configured with ClusterIP type, use port forwarding to access them:
```bash

# For HTTP

kubectl  port-forward  svc/trading  8080:8080

# For gRPC

kubectl  port-forward  svc/trading  50057:50057

```
To change the service type to NodePort, edit the app-deployment.yaml file and uncomment the nodePort lines, then change the service type to NodePort.

  

### Troubleshooting

If pods are not starting properly, check the pod status and logs:
```bash

kubectl  get  pods

kubectl  describe  pod  <pod-name>

kubectl  logs  <pod-name>

```
### Common issues:
- ErrImageNeverPull: Make sure the image exists in the Kubernetes environment

- CrashLoopBackOff: Check application logs for errors

- Pending: Check if PersistentVolumeClaims are being provisioned

### Cleanup

To remove all deployed resources:

```bash

kubectl  delete  -f  deployments/k8s/app-deployment.yaml

kubectl  delete  -f  deployments/k8s/postgres-deployment.yaml

kubectl  delete  -f  deployments/k8s/app-config.yaml

```