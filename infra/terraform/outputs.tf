output "namespace" {
  description = "The Kubernetes namespace all resources are deployed into."
  value       = kubernetes_namespace.shortener.metadata[0].name
}

output "ingress_host" {
  description = "URL to access the application (add this host to /etc/hosts pointing at your cluster IP)."
  value       = "http://${var.ingress_host}"
}
