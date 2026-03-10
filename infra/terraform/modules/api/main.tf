resource "kubernetes_deployment" "api" {
  metadata {
    name      = "api"
    namespace = var.namespace
  }

  spec {
    replicas = 1

    selector {
      match_labels = { app = "api" }
    }

    template {
      metadata {
        labels = { app = "api" }
      }
      spec {
        container {
          name              = "api"
          image             = var.image
          image_pull_policy = "IfNotPresent"

          port { container_port = 8080 }

          # Bulk env from ConfigMap (all app settings except the secret)
          env_from {
            config_map_ref { name = var.configmap_name }
          }

          # DATABASE_URL comes from the Secret
          env {
            name = "DATABASE_URL"
            value_from {
              secret_key_ref {
                name = var.secret_name
                key  = "database_url"
              }
            }
          }

          readiness_probe {
            http_get {
              path = "/ping"
              port = 8080
            }
            initial_delay_seconds = 10
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "api" {
  metadata {
    name      = "api"
    namespace = var.namespace
  }
  spec {
    selector = { app = "api" }
    port {
      port        = 8080
      target_port = 8080
      protocol    = "TCP"
    }
  }
}
