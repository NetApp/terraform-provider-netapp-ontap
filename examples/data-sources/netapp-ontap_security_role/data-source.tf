# data "netapp-ontap_security_role" "security_role" {
#   # required to know which system to interface with
#   cx_profile_name = "cluster4"
#   name = "vsadmin"
#   svm_name = "acc_test"
# }
data "netapp-ontap_security_role" "security_role" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
  name = "vsadmin"
  svm_name = "fsx"
}