# data "netapp-ontap_nfs_services" "protocols_nfs_services" {
#   # required to know which system to interface with
#   cx_profile_name = "cluster4"
#   filter = {
#     svm_name = "ansibleV*"
#   }
# }
data "netapp-ontap_nfs_services" "protocols_nfs_services" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
  filter = {
    svm_name = "fsx*"
  }
}