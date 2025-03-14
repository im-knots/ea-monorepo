resource "helm_release" "external_dns" {
  name       = "external-dns"
  repository = "https://kubernetes-sigs.github.io/external-dns"
  chart      = "external-dns"
  namespace  = "external-dns"
  
  create_namespace = true

#   values = [file("${path.module}/helm-values.yaml")]
}