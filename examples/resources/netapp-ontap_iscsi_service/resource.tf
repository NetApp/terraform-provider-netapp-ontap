resource "netapp-ontap_iscsi_service" "protocols_iscsi_service" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  svm_name        = "svm02"
  enabled         = true
  # target = {
  #   alias = "test"
  # }
}
