data "netapp-ontap_network_ip_service_policy" "svm_scoped" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name            = "test_service_policy1"
  svm_name        = "svm1"
}

data "netapp-ontap_network_ip_service_policy" "cluster_scoped" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name            = "default-management"
  scope           = "cluster"
  ipspace         = "Default"
}
