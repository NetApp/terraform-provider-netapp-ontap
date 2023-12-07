data "netapp-ontap_protocols_cifs_local_group_data_source" "protocols_cifs_local_group" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name         = "svm3"
  name             = "Administrators"
}
