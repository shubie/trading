# Kubernetes module for trading application

# App ConfigMap
resource "kubernetes_config_map" "app_config" {
  metadata {
    name = "app-config"
  }

  data = {
    "config.yaml" = <<-EOT
      binance:
        wss_url: wss://stream.binance.com:9443/ws
        symbols: [BTCUSDT, ETHUSDT, PEPEUSDT]
      grpc:
        port: 50057
      storage:
        postgres:
          dsn: postgres://${var.postgres_user}:${var.postgres_password}@postgres:5432/${var.postgres_db}?sslmode=disable
      buffers:
        tick_chan: 1000
        candle_chan: 500
      health:
        data_timeout: 5m
        port: 8080
    EOT
  }
}

# Postgres PVC
resource "kubernetes_persistent_volume_claim" "postgres_pvc" {
  metadata {
    name = "postgres-pvc"
  }
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = var.postgres_pvc_size
      }
    }
  }
}

# Postgres Deployment
resource "kubernetes_deployment" "postgres" {
  metadata {
    name = "postgres"
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "postgres"
      }
    }

    template {
      metadata {
        labels = {
          app = "postgres"
        }
      }

      spec {
        container {
          name  = "postgres"
          image = "timescale/timescaledb:latest-pg14"

          env {
            name = "POSTGRES_USER"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postgres_credentials.metadata[0].name
                key  = "username"
              }
            }
          }

          env {
            name = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postgres_credentials.metadata[0].name
                key  = "password"
              }
            }
          }

          env {
            name = "POSTGRES_DB"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postgres_credentials.metadata[0].name
                key  = "database"
              }
            }
          }

          port {
            container_port = 5432
          }

          volume_mount {
            name       = "postgres-data"
            mount_path = "/var/lib/postgresql/data"
          }
        }

        volume {
          name = "postgres-data"
          persistent_volume_claim {
            claim_name = kubernetes_persistent_volume_claim.postgres_pvc.metadata[0].name
          }
        }
      }
    }
  }

  # This prevents Terraform from recreating the deployment when only certain fields change
  lifecycle {
    ignore_changes = [
      spec[0].template[0].spec[0].container[0].image,
    ]
  }
}

# Postgres Service
resource "kubernetes_service" "postgres" {
  metadata {
    name = "postgres"
  }

  spec {
    selector = {
      app = kubernetes_deployment.postgres.metadata[0].name
    }

    port {
      port        = 5432
      target_port = 5432
    }

    type = "ClusterIP"
  }
}

# Trading App Deployment
resource "kubernetes_deployment" "trading" {
  metadata {
    name = "trading"
  }

  spec {
    replicas = var.trading_replicas

    selector {
      match_labels = {
        app = "trading"
      }
    }

    template {
      metadata {
        labels = {
          app = "trading"
        }
      }

      spec {
        container {
          name             = "trading"
          image            = var.trading_image
          image_pull_policy = "Never"

          port {
            container_port = 8080
          }

          port {
            container_port = 50057
          }

          volume_mount {
            name       = "config-volume"
            mount_path = "/app/configs/config.yaml"
            sub_path   = "config.yaml"
          }
        }

        volume {
          name = "config-volume"
          config_map {
            name = kubernetes_config_map.app_config.metadata[0].name
          }
        }
      }
    }
  }

  depends_on = [kubernetes_deployment.postgres]
}

# Trading Service
resource "kubernetes_service" "trading" {
  metadata {
    name = "trading"
  }

  spec {
    selector = {
      app = kubernetes_deployment.trading.metadata[0].name
    }

    port {
      name       = "http"
      port       = 8080
      node_port  = 30080
    }

    port {
      name       = "grpc"
      port       = 50057
      node_port  = 30057
    }

    type = "LoadBalancer"
  }
}

# HTTP Ingress
resource "kubernetes_ingress_v1" "trading_http_ingress" {
  metadata {
    name = "trading-http-ingress"
    annotations = {
      "kubernetes.io/ingress.class"            = "nginx"
      "nginx.ingress.kubernetes.io/ssl-redirect" = "false"
      "nginx.ingress.kubernetes.io/use-regex"  = "true"
    }
  }

  spec {
    rule {
      host = var.http_host
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = kubernetes_service.trading.metadata[0].name
              port {
                number = 8080
              }
            }
          }
        }
      }
    }
  }
}

# gRPC Ingress
resource "kubernetes_ingress_v1" "trading_grpc_ingress" {
  metadata {
    name = "trading-grpc-ingress"
    annotations = {
      "kubernetes.io/ingress.class"               = "nginx"
      "nginx.ingress.kubernetes.io/backend-protocol" = "GRPC"
      "nginx.ingress.kubernetes.io/ssl-redirect"   = "false"
    }
  }

  spec {
    rule {
      host = var.grpc_host
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = kubernetes_service.trading.metadata[0].name
              port {
                number = 50057
              }
            }
          }
        }
      }
    }
  }
}

# PostgreSQL secrets
resource "kubernetes_secret" "postgres_credentials" {
  metadata {
    name = "postgres-credentials"
  }

  data = {
    username = var.postgres_user
    password = var.postgres_password
    database = var.postgres_db
  }

  type = "Opaque"
  
  lifecycle {
    # Either prevent Terraform from recreating the secret if it exists
    # prevent_destroy = true
    # Or ignore changes to the data (if you want to manage changes outside Terraform)
    # ignore_changes = [data]
  }
}