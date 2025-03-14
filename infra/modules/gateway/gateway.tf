resource "kubernetes_manifest" "ea_gateway" {
  manifest = {
    "apiVersion" = "gateway.networking.k8s.io/v1"
    "kind"       = "Gateway"
    "metadata" = {
      "name"      = "ea-gateway"
      "namespace" = var.namespace
    }
    "spec" = {
      "gatewayClassName" = "istio"
      "listeners" = [
        {
          "name"     = "http"
          "hostname" = var.gateway_host
          "port"     = 80
          "protocol" = "HTTP"
          "allowedRoutes" = {
            "namespaces" = { "from" = "All" }
          }
        }
      ]
    }
  }
}
