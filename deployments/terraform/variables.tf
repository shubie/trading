# Variables for the Kubernetes deployment

variable "kubernetes_host" {
  description = "Kubernetes API server endpoint"
  type        = string
  default     = ""
}

variable "kubernetes_token" {
  description = "Kubernetes authentication token"
  type        = string
  default     = ""
  sensitive   = true
}

variable "kubernetes_cluster_ca_certificate" {
  description = "Kubernetes cluster CA certificate"
  type        = string
  default     = ""
  sensitive   = true
}


variable "postgres_user" {
  description = "PostgreSQL username"
  type        = string
  default     = "postgres"
}

variable "postgres_password" {
  description = "PostgreSQL password"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "postgres_db" {
  description = "PostgreSQL database name"
  type        = string
  default     = "trading"
}