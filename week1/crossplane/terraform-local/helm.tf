provider "helm" {
  kubernetes {
    config_path = "./config.yaml"
  }
}

resource "helm_release" "argo_cd" {
  depends_on       = [module.k3s, local_sensitive_file.kubeconfig]
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  namespace        = "argocd"
  create_namespace = true
}

resource "helm_release" "crossplane" {
  depends_on       = [module.k3s, local_sensitive_file.kubeconfig]
  name             = "crossplane"
  repository       = "https://charts.crossplane.io/stable"
  chart            = "crossplane"
  namespace        = "crossplane-system"
  create_namespace = true

  set {
    name  = "rovider.packages"
    value = "{xpkg.upbound.io/crossplane-contrib/provider-aws:v0.39.0}"
  }
}

