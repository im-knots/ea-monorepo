resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = "monitoring"
  }
}

resource "helm_release" "prometheus" {
  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "prometheus"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  values = [file("${path.module}/helm-values/prometheus-helm-values.yaml")]
}

resource "helm_release" "alloy" {
  name       = "alloy"
  repository = "https://grafana.github.io/helm-charts"
  chart      = "alloy"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  values = [file("${path.module}/helm-values/alloy-helm-values.yaml")]
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
    namespace = kubernetes_namespace.monitoring.metadata[0].name
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
    namespace        = kubernetes_namespace.monitoring.metadata[0].name

    values = [file("${path.module}/helm-values/loki-helm-values.yaml")]

    depends_on = [ 
        kubernetes_persistent_volume.loki_pv,
        kubernetes_persistent_volume_claim.loki_pvc
    ]
}

resource "helm_release" "grafana" {
    name             = "grafana"
    repository       = "https://grafana.github.io/helm-charts"
    chart            = "grafana"
    namespace        = kubernetes_namespace.monitoring.metadata[0].name

    values = [file("${path.module}/helm-values/grafana-helm-values.yaml")]
}

// =====================================
//  Security/Compliance Tools
// =====================================

//  Trivy Operator is a continuous security scanning tool that runs inside the cluster.
//    - It scans container images, Kubernetes configurations, and infrastructure for vulnerabilities & misconfigurations.
//    - Helps with SOC 2 compliance by detecting security risks before they impact workloads.
//    - Outputs vulnerability reports, misconfigurations, and exposed secrets directly into Kubernetes CRDs.
resource "helm_release" "trivy" {
    name             = "trivy"
    repository       = "https://aquasecurity.github.io/helm-charts/"
    chart            = "trivy-operator"
    namespace        = kubernetes_namespace.monitoring.metadata[0].name

    values = [file("${path.module}/helm-values/trivy-helm-values.yaml")]
}

//  Falco is a real-time runtime security tool for Kubernetes.
//    - It monitors system calls and detects suspicious behavior inside running containers.
//    - Uses pre-defined and custom rules to alert on security threats (e.g., unexpected process execution, privilege escalation).
//    - Logs are collected in Loki via Alloy, enabling security monitoring and incident response.
resource "helm_release" "falco" {
  name       = "falco"
  repository = "https://falcosecurity.github.io/charts"
  chart      = "falco"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  values = [file("${path.module}/helm-values/falco-helm-values.yaml")]
}

resource "kubernetes_config_map" "global_dashboards" {
  for_each = fileset("${path.module}/global-dashboards", "*.json")

  metadata {
    name      = "grafana-dashboard-${replace(each.value, ".json", "")}"
    namespace = kubernetes_namespace.monitoring.metadata[0].name
    labels = {
      grafana_dashboard = "1"
    }
  }

  data = {
    "${each.value}" = file("${path.module}/global-dashboards/${each.value}")
  }
}
