module "k3s" {
  source                   = "xunleii/k3s/module"
  k3s_version              = "v1.28.11+k3s2"
  generate_ca_certificates = true
  drain_timeout            = "30s"
  global_flags = [
    "--tls-san ${tencentcloud_instance.aiops-cvm1.public_ip}",
    "--write-kubeconfig-mode 644",
    "--disable=traefik",
    "--kube-controller-manager-arg bind-address=0.0.0.0",
    "--kube-proxy-arg metrics-bind-address=0.0.0.0",
    "--kube-scheduler-arg bind-address=0.0.0.0"
  ]
  k3s_install_env_vars = {}

  servers = {
    "k3s" = {
      ip = tencentcloud_instance.aiops-cvm1.private_ip
      connection = {
        timeout  = "60s"
        type     = "ssh"
        host     = tencentcloud_instance.aiops-cvm1.public_ip
        private_key = file("~/mykey.pem")
        user     = "ubuntu"
      }
    }
  }
  
  depends_on = [tencentcloud_instance.aiops-cvm1]
}

# resource "local_sensitive_file" "kubeconfig" {
#   content  = module.k3s.kube_config
#   filename = "${path.module}/config.yaml"
# }

resource "null_resource" "fetch_kubeconfig" {
  provisioner "remote-exec" {
    connection {
      type     = "ssh"
      host     = tencentcloud_instance.aiops-cvm1.public_ip
      user     = "ubuntu"
      private_key = file("~/mykey.pem")
    }

    inline = [
      "sudo cp /etc/rancher/k3s/k3s.yaml /tmp/k3s.yaml",
      "sudo chown ubuntu:ubuntu /tmp/k3s.yaml",
      "sed -i 's/127.0.0.1/${tencentcloud_instance.aiops-cvm1.public_ip}/g' /tmp/k3s.yaml"
    ]
  }
  depends_on = [module.k3s]
}

resource "null_resource" "download_k3s_yaml" {
  provisioner "local-exec" {
    command = "scp -i ~/mykey.pem -o StrictHostKeyChecking=no ubuntu@${tencentcloud_instance.aiops-cvm1.public_ip}:/tmp/k3s.yaml ${path.module}/config.yaml"
  }
  depends_on = [null_resource.fetch_kubeconfig]
}
