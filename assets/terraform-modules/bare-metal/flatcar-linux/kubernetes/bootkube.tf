# Self-hosted Kubernetes assets (kubeconfig, manifests)
module "bootkube" {
  source = "../../../bootkube"

  cluster_name                    = var.cluster_name
  api_servers                     = var.api_servers
  api_servers_external            = [var.k8s_domain_name]
  api_servers_ips                 = var.api_servers_ips
  etcd_servers                    = var.etcd_servers_domains
  etcd_endpoints                  = var.etcd_servers_ips
  asset_dir                       = var.asset_dir
  network_mtu                     = var.network_mtu
  network_ip_autodetection_method = var.network_ip_autodetection_method
  pod_cidr                        = var.pod_cidr
  service_cidr                    = var.service_cidr
  cluster_domain_suffix           = var.cluster_domain_suffix
  enable_reporting                = var.enable_reporting
  enable_aggregation              = var.enable_aggregation
  kube_apiserver_extra_flags      = var.kube_apiserver_extra_flags

  certs_validity_period_hours = var.certs_validity_period_hours

  # Disable the self hosted kubelet.
  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet

  bootstrap_tokens     = [local.controller_bootstrap_token, local.worker_bootstrap_token]
  enable_tls_bootstrap = true
  encrypt_pod_traffic  = var.encrypt_pod_traffic

  ignore_x509_cn_check = var.ignore_x509_cn_check

  conntrack_max_per_core = var.conntrack_max_per_core
}
