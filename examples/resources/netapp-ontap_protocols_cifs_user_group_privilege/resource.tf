resource "netapp-ontap_protocols_cifs_user_group_privilege_resource" "protocols_cifs_user_group_privilege_test3" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "test3"
  svm_name = "test3"
  privileges = ["setcbprivilege", "sesecurityprivilege"]
}

resource "netapp-ontap_protocols_cifs_user_group_privilege_resource" "protocols_cifs_user_group_privilege_test" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "test"
  svm_name = "test3"
  privileges = [lower("SeTcbPrivilege"), lower("SeSecurityPrivilege")]
}
