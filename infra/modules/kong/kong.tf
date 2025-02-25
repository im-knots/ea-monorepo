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

resource "kubernetes_manifest" "global_jwt_auth" {
  manifest = {
    apiVersion = "configuration.konghq.com/v1"
    kind       = "KongClusterPlugin"
    metadata = {
      name = "global-jwt-auth"
      annotations = {
        "kubernetes.io/ingress.class" = "kong"
      }
    }
    plugin = "jwt"
    config = {
      uri_param_names = ["paramName_2.2.x"]
    }
  }
}

resource "kubernetes_manifest" "kong_cluster_plugin_cors" {
  manifest = {
    apiVersion = "configuration.konghq.com/v1"
    kind       = "KongClusterPlugin"
    metadata = {
      name = "global-cors"
      annotations = {
        "kubernetes.io/ingress.class" = "kong"
      }
    }
    plugin = "cors"
    config = {
      origins         = ["*"]
      methods         = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      headers         = ["Accept", "Accept-Version", "Content-Length", "Content-MD5", "Content-Type", "Date", "X-Auth-Token", "Authorization"]
      exposed_headers = ["X-Auth-Token", "Authorization"]
      credentials     = true
      max_age         = 3600
    }
  }
}