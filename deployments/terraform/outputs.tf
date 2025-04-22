# Outputs from the Kubernetes deployment

output "trading_service_http_url" {
  description = "URL for the trading HTTP service"
  value       = "${module.trading_app.trading_service_http_nodeport}"
}

output "trading_service_grpc_url" {
  description = "URL for the trading gRPC service"
  value       = "${module.trading_app.trading_service_grpc_nodeport}"
}

output "postgres_service_name" {
  description = "Name of the PostgreSQL service"
  value       = module.trading_app.postgres_service_name
}