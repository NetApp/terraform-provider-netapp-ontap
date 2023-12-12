data "netapp-ontap_protocols_cifs_local_users_data_source" "protocols_cifs_local_users" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "user1"
    svm_name = "svm*"
  }
}
