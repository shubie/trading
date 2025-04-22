# Outputs for the Kubernetes module

output "trading_service_http_nodeport" {
  description = "NodePort for trading HTTP service"
  value       = "http://localhost:${kubernetes_service.trading.spec[0].port[0].node_port}"
}

output "trading_service_grpc_nodeport" {
  description = "NodePort for trading gRPC service"
  value       = "grpc://localhost:${kubernetes_service.trading.spec[0].port[1].node_port}"
}

output "postgres_service_name" {
  description = "Name of the PostgreSQL service"
  value       = kubernetes_service.postgres.metadata[0].name
}