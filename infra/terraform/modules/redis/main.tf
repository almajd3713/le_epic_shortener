resource "kubernetes_deployment" "cache" {
  metadata {
    name      = "cache"
    namespace = var.namespace
  }

  spec {
    replicas = 1

    selector {
      match_labels = { app = "redis" }
    }

    template {
      metadata {
        labels = { app = "redis" }
      }
      spec {
        container {
          name  = "redis"
          image = "redis:7-alpine"

          port { container_port = 6379 }

          readiness_probe {
            exec {
              command = ["redis-cli", "ping"]
            }
            initial_delay_seconds = 5
            period_seconds        = 10
            timeout_seconds       = 3
            failure_threshold     = 5
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "cache" {
  metadata {
    name      = "cache"
    namespace = var.namespace
  }
  spec {
    selector = { app = "redis" }
    port {
      port        = 6379
      target_port = 6379
      protocol    = "TCP"
    }
  }
}
