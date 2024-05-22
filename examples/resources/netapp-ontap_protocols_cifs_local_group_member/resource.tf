resource "netapp-ontap_protocols_cifs_local_group_member" "protocols_cifs_local_group_member" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  group_name = "SVM_SERVER\\testme"
  svm_name = "test3"
  member = "SVM_SERVER\\test"
}
