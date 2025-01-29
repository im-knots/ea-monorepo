resource "helm_release" "kube_prometheus_stack" {
    name             = "kps"
    repository       = "https://prometheus-community.github.io/helm-charts"
    chart            = "kube-prometheus-stack"
    namespace        = "monitoring"
    create_namespace = true

    values = [file("${path.module}/kps-helm-values.yaml")]
}