resource "kubernetes_namespace" "kong" {
  metadata {
    name = "kong"
  }
}

resource "helm_release" "kong" {
  name              = "kong-ingress"
  repository        = "https://charts.konghq.com"
  chart             = "ingress"
  namespace         = kubernetes_namespace.kong.metadata[0].name

  wait    = false

  values = [file("${path.module}/helm-values.yaml")]

}