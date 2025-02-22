resource "helm_release" "alloy" {
  name       = "alloy"
  namespace  = "monitoring"
  repository = "https://grafana.github.io/helm-charts"
  chart      = "alloy"

  values = [file("${path.module}/alloy-helm-values.yaml")]
}

resource "helm_release" "prometheus" {
  name       = "prometheus"
  namespace  = "monitoring"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "prometheus"

  values = [file("${path.module}/prometheus-helm-values.yaml")]
}


resource "kubernetes_persistent_volume" "loki_pv" {
  metadata {
    name = "loki-pv"
  }

  spec {
    capacity = {
      storage = "10Gi"
    }
    access_modes = ["ReadWriteOnce"]
    persistent_volume_source {
      host_path {
        path = "/mnt/data/loki"
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "loki_pvc" {
  metadata {
    name      = "loki-pvc"
    namespace = "monitoring"
  }

  spec {
    access_modes = ["ReadWriteOnce"]

    resources {
      requests = {
        storage = "10Gi"
      }
    }
  }
}

resource "helm_release" "loki" {
    name             = "loki"
    repository       = "https://grafana.github.io/helm-charts"
    chart            = "loki"
    namespace        = "monitoring"
    create_namespace = false

    values = [file("${path.module}/loki-helm-values.yaml")]

    depends_on = [ 
        helm_release.alloy,
        kubernetes_persistent_volume.loki_pv,
        kubernetes_persistent_volume_claim.loki_pvc
    ]
}

resource "helm_release" "grafana" {
    name             = "grafana"
    repository       = "https://grafana.github.io/helm-charts"
    chart            = "grafana"
    namespace        = "monitoring"
    create_namespace = false

    values = [file("${path.module}/grafana-helm-values.yaml")]
    depends_on = [ 
        helm_release.alloy,
        helm_release.loki
    ]
}