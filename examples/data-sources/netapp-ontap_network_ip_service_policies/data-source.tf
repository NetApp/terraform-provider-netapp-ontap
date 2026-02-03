# retrieving cluster-scoped service policies
data "netapp-ontap_network_ip_service_policies" "example1" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    scope =    "cluster"
  }
}

# retrieving svm-scoped service policies
data "netapp-ontap_network_ip_service_policies" "example2" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"    
  }
}

# retrieving service policies with matching name pattern
data "netapp-ontap_network_ip_service_policies" "example3" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
    name = "test_service_policy*"
  }
}
