data "netapp-ontap_protocols_cifs_local_group_members_data_source" "protocols_cifs_local_group_members" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "test3"
  group_name = "testme"
  # filter = {
    # "name" = "testme"
    # "svm_name" = "test3"
  # }
}
