data "netapp-ontap_network_ip_interfaces" "cluster_scoped" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name  = "*"
    scope = "cluster"
  }
}

data "netapp-ontap_network_ip_interfaces" "svm_scoped" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name     = "*"
    svm_name = "svm*"
    scope    = "svm"
  }
}
