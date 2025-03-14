resource "helm_release" "app" {
  for_each  = var.apps

  name      = each.key
  chart     = each.value.chart
  namespace = var.namespace
  version   = each.value.version

  dynamic "set" {
    for_each = each.value.helm_overrides

    content {
      name  = set.key
      value = set.value
      type  = "auto"
    }
  }

  create_namespace = false
}
