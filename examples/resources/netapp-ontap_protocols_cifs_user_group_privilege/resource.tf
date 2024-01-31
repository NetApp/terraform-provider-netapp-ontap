resource "netapp-ontap_protocols_cifs_user_group_privilege_resource" "protocols_cifs_user_group_privilege" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "test3"
  svm_name = "test3"
  privileges = ["SeTcbPrivilege", "SeSecurityPrivilege"]
}
