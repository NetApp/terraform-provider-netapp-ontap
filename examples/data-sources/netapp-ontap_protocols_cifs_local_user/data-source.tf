data "netapp-ontap_protocols_cifs_local_user_data_source" "protocols_cifs_local_user" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "svm1"
  name = "testme"
}
