data "netapp-ontap_security_roles" "security_roles" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    svm_name = "acc_test"
    scope = "cluster"
  }
}
