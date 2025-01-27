resource "helm_release" "name" {
  name              = "ollama"
  repository        = "https://otwld.github.io/ollama-helm"
  chart             = "ollama"
  namespace         = "ea-platform-ollama"
  create_namespace  = true

  values = [file("${path.module}/helm-values.yaml")]

}