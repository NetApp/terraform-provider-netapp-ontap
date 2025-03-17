data "netapp-ontap_iscsi_service" "protcols_iscsi_service" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  svm_name        = "svm02"
}
