
# ---------------------------------------------------------------------------
# Application
# ---------------------------------------------------------------------------

variable "app_port" {
  type    = string
  default = "8080"
}

variable "base_url" {
  type    = string
  default = "http://shortener.local"
}

variable "environment" {
  type    = string
  default = "production"
}

variable "log_level" {
  type    = string
  default = "info"
}

variable "allowed_origins" {
  type    = string
  default = "http://shortener.local"
}

variable "trusted_proxies" {
  type    = string
  default = ""
}

# ---------------------------------------------------------------------------
# Database secrets
# ---------------------------------------------------------------------------

variable "database_url" {
  type      = string
  sensitive = true
  default   = "postgres://postgres:password@db:5432/postgres"
}

variable "postgres_user" {
  type      = string
  sensitive = true
  default   = "postgres"
}

variable "postgres_password" {
  type      = string
  sensitive = true
  default   = "password"
}

variable "postgres_db" {
  type    = string
  default = "postgres"
}

# ---------------------------------------------------------------------------
# Infrastructure
# ---------------------------------------------------------------------------

variable "ingress_host" {
  type    = string
  default = "shortener.local"
}

variable "storage_class" {
  type        = string
  default     = "local-path"
  description = "StorageClass for the Postgres PVC. Use 'standard' for Minikube, 'local-path' for k3s."
}

# ---------------------------------------------------------------------------
# Images
# ---------------------------------------------------------------------------

variable "api_image" {
  type    = string
  default = "ghcr.io/almajd3713/le_epic_shortener/api:latest"
}

variable "frontend_image" {
  type    = string
  default = "ghcr.io/almajd3713/le_epic_shortener/frontend:latest"
}
