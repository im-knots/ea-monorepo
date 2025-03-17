resource "helm_release" "external_dns" {
  name       = "external-dns"
  repository = "https://kubernetes-sigs.github.io/external-dns"
  chart      = "external-dns"
  namespace  = "external-dns"
  
  create_namespace = true

  values = [file("${path.module}/helm-values.yaml")]

  set {
    name = "extraArgs[0]"
    value = "--google-project=${var.mgmt_project}"
  }

  set {
    name = "extraArgs[1]"
    value = "--txt-owner-id=${var.mgmt_project}-${var.cluster_name}"
  }

  set {
    name = "extraArgs[2]"
    value = "--zone-id-filter=${var.public_zone_id}"
  }

  set {
    name = "domainFilters[1]"
    value = var.public_domain
  }

  set {
    name = "serviceAccount.annotations.iam\\.gke\\.io/gcp-service-account"
    value = var.dns_service_account
  }

}