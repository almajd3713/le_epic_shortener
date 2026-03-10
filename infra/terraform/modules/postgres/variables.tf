variable "namespace" {
  type = string
}

variable "postgres_user" {
  type      = string
  sensitive = true
}

variable "postgres_db" {
  type = string
}

variable "storage_class" {
  type    = string
  default = "local-path"
}
