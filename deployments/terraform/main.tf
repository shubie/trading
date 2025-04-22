# Main Terraform configuration for Kubernetes resources

terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23.0"
    }
  }
  required_version = ">= 1.0.0"
}

provider "kubernetes" {
  # Configuration for local minikube
  config_path    = "~/.kube/config"
  # If you have multiple contexts in your kubeconfig, you may need to specify which one to use
  # config_context = "minikube"
  
  # For production environments, you might use:
  # host                   = var.kubernetes_host
  # token                  = var.kubernetes_token
  # cluster_ca_certificate = var.kubernetes_cluster_ca_certificate
}

module "trading_app" {
  source = "./modules/kubernetes"

  # Postgres configuration from secrets
  postgres_user     = var.postgres_user     # These would get values from variable files or environment
  postgres_password = var.postgres_password
  postgres_db       = var.postgres_db
  postgres_pvc_size = "5Gi"

  # Trading app configuration
  trading_replicas = 1
  trading_image    = "trading:local"
  
  # Ingress configuration
  http_host = "trading.local"
  grpc_host = "grpc.trading.local"
}