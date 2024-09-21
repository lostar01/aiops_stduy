module "k3s" {
  source                   = "xunleii/k3s/module"
  k3s_version              = "v1.28.11+k3s2"
  generate_ca_certificates = true
  drain_timeout            = "30s"
  global_flags = [
    "--tls-san ${var.local_ip}",
    "--write-kubeconfig-mode 644",
    "--disable=traefik",
    "--kube-controller-manager-arg bind-address=0.0.0.0",
    "--kube-proxy-arg metrics-bind-address=0.0.0.0",
    "--kube-scheduler-arg bind-address=0.0.0.0"
  ]
  k3s_install_env_vars = {}
 
 servers = {
    "k3s" = {
      ip = var.local_ip
      connection = {
        timeout  = "60s"
        type     = "ssh"
        host     = var.local_ip
        user     = var.username
        password = var.password
      }
    }
  }
 
}

resource "local_sensitive_file" "kubeconfig" {
  content  = module.k3s.kube_config
  filename = "${path.module}/config.yaml"
}

resource "null_resource" "config_kubeconfig" {
  provisioner "remote-exec" {
    connection {
      type     = "ssh"
      host     = var.local_ip
      user     = var.username
      password = var.password
    }

    inline = [
      "mkdir ~/.kube",
      "until [ -f /etc/rancher/k3s/k3s.yaml ]; do cp /etc/rancher/k3s/k3s.yaml ~/.kube/config; done",
    ]
  }
  depends_on = [module.k3s]
}
