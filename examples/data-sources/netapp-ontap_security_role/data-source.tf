data "netapp-ontap_security_role_data_source" "security_role" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "vsadmin"
  svm_name = "acc_test"
}
