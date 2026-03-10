terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
  backend "local" {}
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

# ---------------------------------------------------------------------------
# Namespace
# ---------------------------------------------------------------------------

resource "kubernetes_namespace" "shortener" {
  metadata {
    name   = "shortener"
    labels = { name = "shortener" }
  }
}

# ---------------------------------------------------------------------------
# Shared ConfigMap  (app env vars)
# ---------------------------------------------------------------------------

resource "kubernetes_config_map" "shortener_config" {
  metadata {
    name      = "shortener-config"
    namespace = kubernetes_namespace.shortener.metadata[0].name
  }

  data = {
    PORT                    = var.app_port
    BASE_URL                = var.base_url
    ENV                     = var.environment
    LOG_LEVEL               = var.log_level
    ALLOWED_ORIGINS         = var.allowed_origins
    TRUSTED_PROXIES         = var.trusted_proxies
    REDIS_URL               = "redis://cache:6379"
    REDIS_MAX_RETRIES       = "5"
    REDIS_MIN_RETRY_BACKOFF = "100ms"
    REDIS_MAX_RETRY_BACKOFF = "1s"
  }
}

# ---------------------------------------------------------------------------
# Shared Secret
# ---------------------------------------------------------------------------

resource "kubernetes_secret" "shortener_secrets" {
  metadata {
    name      = "shortener-secrets"
    namespace = kubernetes_namespace.shortener.metadata[0].name
  }

  data = {
    database_url      = var.database_url
    postgres_user     = var.postgres_user
    postgres_password = var.postgres_password
    postgres_db       = var.postgres_db
  }
}

# ---------------------------------------------------------------------------
# Ingress
# ---------------------------------------------------------------------------

resource "kubernetes_ingress_v1" "shortener" {
  metadata {
    name      = "shortener-ingress"
    namespace = kubernetes_namespace.shortener.metadata[0].name
    annotations = {
      "nginx.ingress.kubernetes.io/ssl-redirect" = "false"
    }
  }
  
  spec {
    rule {
      host = var.ingress_host
      http {
        path {
          path      = "/api"
          path_type = "Prefix"
          backend {
            service {
              name = "api"
              port { number = 8080 }
            }
          }
        }
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "frontend"
              port { number = 80 }
            }
          }
        }
      }
    }
  }

  depends_on = [module.api, module.nginx]
}

# ---------------------------------------------------------------------------
# Modules
# ---------------------------------------------------------------------------

module "postgres" {
  source        = "./modules/postgres"
  namespace     = kubernetes_namespace.shortener.metadata[0].name
  postgres_user = var.postgres_user
  postgres_db   = var.postgres_db
  storage_class = var.storage_class

  depends_on = [kubernetes_secret.shortener_secrets]
}

module "redis" {
  source    = "./modules/redis"
  namespace = kubernetes_namespace.shortener.metadata[0].name
}

module "api" {
  source         = "./modules/api"
  namespace      = kubernetes_namespace.shortener.metadata[0].name
  image          = var.api_image
  configmap_name = kubernetes_config_map.shortener_config.metadata[0].name
  secret_name    = kubernetes_secret.shortener_secrets.metadata[0].name

  depends_on = [module.postgres, module.redis]
}

module "nginx" {
  source    = "./modules/nginx"
  namespace = kubernetes_namespace.shortener.metadata[0].name
  image     = var.frontend_image

  depends_on = [module.api]
}