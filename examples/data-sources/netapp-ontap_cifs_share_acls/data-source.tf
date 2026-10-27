data "netapp-ontap_cifs_share_acls" "protocols_cifs_share_acls" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "d_svm"
  filter = {
    share_name = "*"
  }
}
