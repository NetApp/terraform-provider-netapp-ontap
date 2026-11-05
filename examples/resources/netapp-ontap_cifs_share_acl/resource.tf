resource "netapp-ontap_cifs_share_acl" "protocols_cifs_share_acl" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name        = "AnsibleSVM"
  share_name      = "test_share"
  user_or_group   = "Everyone"
  type            = "windows"
  permission      = "no_access"
}
