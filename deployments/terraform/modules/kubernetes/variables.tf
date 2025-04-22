# Variables for the Kubernetes module

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

variable "postgres_pvc_size" {
  description = "Size of the PostgreSQL persistent volume claim"
  type        = string
  default     = "5Gi"
}

variable "trading_replicas" {
  description = "Number of trading app replicas"
  type        = number
  default     = 1
}

variable "trading_image" {
  description = "Docker image for the trading app"
  type        = string
  default     = "trading:local"
}

variable "http_host" {
  description = "Hostname for HTTP ingress"
  type        = string
  default     = "trading.local"
}

variable "grpc_host" {
  description = "Hostname for gRPC ingress"
  type        = string
  default     = "grpc.trading.local"
}