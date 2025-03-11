resource "helm_release" "app_postgresql" {
  name             = "postgresql"
  chart            = "bitnami/postgresql"
  version         = "16.5.0"
  namespace        = var.namespace
  create_namespace = false

 values = [file("${path.module}/helm-values.yaml")]
}
