resource "null_resource" "minikube_start" {
  provisioner "local-exec" {
    when    = create
    command = "minikube start --driver=docker --container-runtime docker --gpus all --extra-config=kubelet.max-pods=1000 --cni calico"
  }
}

resource "null_resource" "minikube_set_fs_watches" {
  provisioner "local-exec" {
    when    = create
    command = "minikube ssh -- \"sudo sysctl -w fs.inotify.max_user_watches=10000 fs.inotify.max_user_instances=1000 fs.file-max=50000 && ulimit -n 50000\""
  }
  depends_on = [ null_resource.minikube_start ]
}

resource "null_resource" "minikube_enable_addons" {
  provisioner "local-exec" {
    when    = create
    command = "minikube addons enable registry && minikube addons enable metrics-server"
  }
  depends_on = [ null_resource.minikube_start ]
}

resource "null_resource" "minikube_cleanup" {
  provisioner "local-exec" {
    when    = destroy
    command = "minikube delete"
  }
}