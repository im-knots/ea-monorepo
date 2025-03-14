resource "helm_release" "external_dns" {
  name       = "external-dns"
  repository = "https://kubernetes-sigs.github.io/external-dns"
  chart      = "external-dns"
  namespace  = "external-dns"
  
  create_namespace = true

  values = [file("${path.module}/helm-values.yaml")]

  dynamic "set" {
    for_each = concat(
      [
        { name : "extraArgs[0]", value : "--google-project=${var.project}"},
        { name : "extraArgs[1]", value : "--txt-owner-id=${var.project}-${var.cluster_name}" },
        { name : "extraArgs[2]", value : "--zone-id-filter=${var.public_zone_id}" },
        { name : "domainFilters[1]", value : var.public_domain },
        { name : "serviceAccount.annotations.iam\\.gke\\.io/gcp-service-account", value : "135550390903-compute@developer.gserviceaccount.com" },
      ]
    )
    content {
      name  = set.key
      value = set.value
      type  = "auto"
    }
  }
}