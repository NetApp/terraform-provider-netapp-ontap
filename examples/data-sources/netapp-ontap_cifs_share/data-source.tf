data "netapp-ontap_cifs_share" "protocols_cifs_share" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name = "share1"
  svm_name = "testSVM"
}
