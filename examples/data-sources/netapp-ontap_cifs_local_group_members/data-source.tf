data "netapp-ontap_cifs_local_group_members" "protocols_cifs_local_group_members" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "test3"
  group_name = "testme"
}
