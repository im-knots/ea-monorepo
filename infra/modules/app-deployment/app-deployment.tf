resource "helm_release" "app" {
  name       = var.app_name
  chart      = var.chart_path
  namespace  = var.namespace
  
  create_namespace = false
}

