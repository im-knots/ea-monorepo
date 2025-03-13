resource "kubernetes_namespace" "istio_system" {
  metadata {
    name = "istio-system"
  }
}

resource "null_resource" "istio_namespace" {
  triggers = {
    namespace = kubernetes_namespace.istio_system.metadata[0].name
  }

  provisioner "local-exec" {
    command = "kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml --context minikube"
  }
}

resource "helm_release" "istio_base" {
  name              = "istio-base"
  repository        = "https://istio-release.storage.googleapis.com/charts"
  chart             = "base"
  namespace         = kubernetes_namespace.istio_system.metadata[0].name

  wait    = false

  depends_on = [ 
    kubernetes_namespace.istio_system
  ]

  set {
    name = "profile"
    value = "platform-minikube"
  }
}

resource "helm_release" "istiod" {
  name              = "istiod"
  repository        = "https://istio-release.storage.googleapis.com/charts"
  chart             = "istiod"
  namespace         = kubernetes_namespace.istio_system.metadata[0].name

  wait    = false

  depends_on = [ 
    helm_release.istio_base
  ]

  set {
    name = "profile"
    value = "platform-minikube"
  }
}
