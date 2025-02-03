resource "helm_release" "app_mongodb" {
  name             = "mongodb"
  chart            = "bitnami/mongodb"
  namespace        = var.namespace
  create_namespace = false

  set {
    name  = "auth.enabled"
    value = "false"
  }

  # set {
  #     name  = "persistence.storageClass"
  #     value = "standard"
  # }
}
