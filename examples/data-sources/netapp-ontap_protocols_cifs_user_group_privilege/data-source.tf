data "netapp-ontap_protocols_cifs_user_group_privilege_data_source" "protocols_cifs_user_group_privilege" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "user1"
  svm_name = "ansibleSVM"
}
