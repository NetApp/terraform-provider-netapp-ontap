# data "netapp-ontap_qos_policies" "qos_policies" {
#   # required to know which system to interface with
#   cx_profile_name = "cluster1"
#   filter = {
#     svm_name = "terraform"
#     name = "test2"
#   }
# }

data "netapp-ontap_qos_policies" "qos_policies" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
  filter = {
    svm_name = "fsx"
    name = "*"
  }
}