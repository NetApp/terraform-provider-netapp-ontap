# data "netapp-ontap_cifs_local_users" "protocols_cifs_local_users" {
#   # required to know which system to interface with
#   cx_profile_name = "cluster4"
#   filter = {
#     name = "user1"
#     svm_name = "svm*"
#   }
# }

data "netapp-ontap_cifs_local_users" "protocols_cifs_local_users" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
  filter = {
    # name = "user1"
    svm_name = "fs*"
  }
}