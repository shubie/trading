# Deployment Instructions
This directory contains deployment configurations for the trading application.

## Terraform Deployment

The `deployments/terraform` directory contains all necessary Terraform manifests to deploy the application stack.

### Prerequisites

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