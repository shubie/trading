# Trading Application

A design and implementation of a trading chart service in Go that reads tick data from Binance stream API, aggregates the data into OHLC candlesticks, broadcasts the current bar to a streaming server API, and stores complete bars into a database. The service should be deployed to a Kubernetes cluster using Terraform.

![Candle Sticks Output](/docs/images/postman%20screenshot.png "Candle Sticks Output")

## Features

- Real-time market data ingestion from Binance WebSocket API

- Processing of tick patterns for multiple trading pairs

- gRPC API for streaming candlestick data

- Kubernetes deployment support with Terraform

- TimescaleDB for time-series data

- Health check endpoint for monitoring

## Project Structure

![Candle Sticks Output](/docs/images/structure.png "Candle Sticks Output")

## Design Documentation

A brief document explaining your design decisions: how the service handles data ingestion, aggregation, and streaming; Data design, Choice of K8s resources and deployment strategy. [Please see the the design document here](/docs/design-document.md).

## Prerequisites

- Go 1.19+
- Docker
- TimescaleDB or PostgreSQL with TimescaleDB extension (there is a docker compose file for this)
- Kubernetes
- Terraform (for Kubernetes deployment)


## Getting Started

### Local Development

1. Clone the repository:
```bash
git  clone  github.com/shubie/trading
cd  trading
```
2. Install dependencies:

```bash
go  mod  download
```

3. Run TimescaleDB with Docker Compose for local development:

```bash
docker-compose  up  -d
```

4. Build and run the application:

```bash
go  build  -o  bin/trading  ./cmd/main.go
./bin/trading
```


### Configuration

The application can be configured through a YAML configuration file. See configs/config.yaml for available options or check the sample configuration in the Kubernetes ConfigMap.

**Key configuration options**:

- Binance WebSocket URL and trading symbols

- gRPC server port

- Database connection string

- Buffer sizes for internal channels

- Health check settings

### Using the gRPC API

The application provides a gRPC API to stream candlestick data. You can use any gRPC client to connect to it.

Example using grpcurl :

```bash
grpcurl  -plaintext  localhost:50057  candlestick.CandlestickService/StreamCandlesticks
```
You also can use postman to test the gRPC API.

### Running test cases

To run all tests in the project, inside the project root directory and use the Go test command:

```bash
cd trading
go test ./...
```
This command will recursively run all tests in all packages of this project.

If you want to run tests for a specific package, you can specify the package path:
```bash
# For example, to test the binance client package
go test ./internal/binance

# Or for any other package
go test ./internal/aggregator
go test ./internal/server
```


## Terraform Deployment

Terraform implementation are provided in the `deployments/terraform` directory. [Please see the Terraform Deployment README for detailed instructions](/docs/Terraform.md).

## Prerequisites

Ensure you have the following installed and configured:

- Terraform : Make sure Terraform is installed on your machine.
- Minikube : A local Kubernetes cluster.
- kubectl : Kubernetes command-line tool.
- Docker : For building and managing container images.

### ### Terraform Deployment Steps
1. Initialize Terraform
   
   Navigate to the directory containing your Terraform configuration files and initialize Terraform:
   
   ```bash
   cd deployments/terraform
   terraform init
    ```
2. Configure Minikube
   
   Ensure Minikube is running and your kubectl is configured to use the Minikube context 
   ```bash
   minikube start
   kubectl config use-context minikube
    ```
2. Apply Terraform Configuration
   
   Apply the Terraform configuration to deploy the resources defined in your Terraform files:
   
   ```bash
   terraform apply
    ```
   
   During this step, Terraform will prompt you to confirm the changes. Type yes to proceed.
3. Verify Deployment
   
   After Terraform completes the deployment, verify that the resources are correctly deployed:
   
   ```bash
   kubectl get deployments
   kubectl get services
   kubectl get pods
    ```
4. Accessing the Application
   
   You can access the application through these components. Use the following command to get the ingress details:
   
   ```bash
   kubectl get ingress
    ```
   
   You can also check the load balancer status:
   
   ```bash
   kubectl get services
    ```
   
   If you are running on minikube, you can run this commane:
   
   ```bash
   minikube tunnel
    ```
    the trading service should now be accessible at:
     
   ```bash
   HTTP: http://localhost:8080
   GRPC: grpc://localhost:50057


   HTTP: http://<minikube-ip>:8080
   GRPC: grpc://<minikube-ip>:50057
    ```
    
### Cleanup
To remove all deployed resources, use Terraform to destroy the infrastructure:

```bash
terraform destroy
 ```
This will remove all resources created by Terraform, ensuring a clean state.

## Kurbenetes Deployment

Kubernetes manifests are defined in the `deployments/k8s` directory. [Please see the Kurbenetes Deployment README for detailed instructions](/docs/K8s.md).


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


## Health Checks

The application exposes a health check endpoint on port 8080. You can access it at:
```p
http://localhost:8080
```
## License

This project is licensed under the MIT License.