data "netapp-ontap_iscsi_services" "protcols_iscsi_services" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  filter = {
    svm_name = "svm*"
  }
}
