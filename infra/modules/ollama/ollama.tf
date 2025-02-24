resource "helm_release" "ollama" {
  name              = "ollama"
  repository        = "https://otwld.github.io/ollama-helm"
  chart             = "ollama"
  namespace         = var.namespace

  values = [file("${path.module}/helm-values.yaml")]

}