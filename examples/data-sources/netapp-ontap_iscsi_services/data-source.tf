data "netapp-ontap_iscsi_services" "protcols_iscsi_services" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  filter = {
    svm_name = "svm*"
  }
}
