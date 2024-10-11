# data "netapp-ontap_cifs_shares" "protocols_cifs_shares" {
#   # required to know which system to interface with
#   cx_profile_name = "cluster5"
#   filter = {
#     svm_name = "testSVM"
#   }
# }
data "netapp-ontap_cifs_shares" "protocols_cifs_shares" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
  filter = {
    svm_name = "fsx"
  }
}