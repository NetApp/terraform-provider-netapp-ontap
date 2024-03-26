data "netapp-ontap_protocols_cifs_share_data_source" "protocols_cifs_share" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name = "share1"
  svm_name = "testSVM"
}
