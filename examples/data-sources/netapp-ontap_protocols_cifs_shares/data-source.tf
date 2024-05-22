data "netapp-ontap_protocols_cifs_shares" "protocols_cifs_shares" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    svm_name = "testSVM"
  }
}
