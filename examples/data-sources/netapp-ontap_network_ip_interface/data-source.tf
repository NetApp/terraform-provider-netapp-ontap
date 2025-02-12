data "netapp-ontap_network_ip_interface" "cluster_scoped" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "cluster_mgmt"
}

data "netapp-ontap_network_ip_interface" "svm_scoped" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "ipv4-svm02-mgmt.labwi.sva.de"
  svm_name        = "svm02"
}
