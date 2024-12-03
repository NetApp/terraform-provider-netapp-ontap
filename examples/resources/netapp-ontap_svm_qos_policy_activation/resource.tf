resource "netapp-ontap_svm_qos_policy_activation" "example" {
  cx_profile_name = "svacluster"
  svm_id          = "b366fb39-8d2d-11ef-9ca8-00a0b8bc0407" # svm02
  qos_policy_id   = "beb86559-8d2d-11ef-9ca8-00a0b8bc0407" # performance-svm02
}
