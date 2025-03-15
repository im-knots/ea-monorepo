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
        },
        {
          "name"     = "https"
          "hostname" = var.gateway_host
          "port"     = 443
          "protocol" = "HTTPS"
          "allowedRoutes" = {
            "namespaces" = { "from" = "All" }
          }
          "tls" = {
            "mode" = "Terminate"
            "certificateRefs" = [{
              "kind" = "Secret"
              "name" = "ea-gateway-tls"
            }]
          }
        }
      ]
    }
  }
}

resource "kubernetes_manifest" "ea-gateway-tls" {
  manifest = {
    "apiVersion" = "v1"
    "kind"       = "Secret"
    "metadata" = {
      "name"      = "ea-gateway-tls"
      "namespace" = var.namespace
    }
    "data" = {
      "tls.crt" = filebase64("${path.module}/certs/tls.crt")
      "tls.key" = filebase64("${path.module}/certs/tls.key")
    }
  }
}