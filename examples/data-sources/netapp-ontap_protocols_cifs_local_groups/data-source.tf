data "netapp-ontap_protocols_cifs_local_groups_data_source" "protocols_cifs_local_groups" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    svm_name = "svm*"
    name     = "Administrators"
  }
}
