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

resource "kubernetes_manifest" "request_authentication_httpbin" {
  manifest = {
    "apiVersion" = "security.istio.io/v1"
    "kind"       = "RequestAuthentication"
    "metadata" = {
      "name"      = "ea-gateway-request-auth"
      "namespace" = var.namespace
    }
    "spec" = {
      "selector" = {
        "matchLabels" = {
          "jwt" = "true"
        }
      }
      "jwtRules" = [
        {
          "issuer"  = "eru-labs-jwt-issuer"
          "jwksUri" = "${var.jwks_uri}/api/jwks"
        }
      ]
    }
  }
}

resource "kubernetes_manifest" "authorization_policy_httpbin" {
  manifest = {
    "apiVersion" = "security.istio.io/v1"
    "kind"       = "AuthorizationPolicy"
    "metadata" = {
      "name"      = "ea-gateway-auth-policy"
      "namespace" = var.namespace
    }
    "spec" = {
      "selector" = {
        "matchLabels" = {
          "jwt" = "true"
        }
      }
      "rules" = [
        {
          "from" = [
            {
              "source" = {
                "requestPrincipals" = ["*"]
              }
            }
          ]
        }
      ]
    }
  }
}

# resource "kubernetes_manifest" "authorization_policy_default_deny" {
#   manifest = {
#     "apiVersion" = "security.istio.io/v1"
#     "kind"       = "AuthorizationPolicy"
#     "metadata" = {
#       "name"      = "ea-gateway-default-deny"
#       "namespace" = var.namespace
#     }
#     "spec" = {
#       "selector" = {
#         "matchLabels" = {
#           "jwt" = "true"
#         }
#       }
#       "action" = "DENY"
#       "rules" = [{}] # matches all traffic
#     }
#   }
# }
