resource "netapp-ontap_svm_qos_policy_activation" "example" {
  cx_profile_name = "svacluster"
  svm = {
    # id = "b366fb39-8d2d-11ef-9ca8-00a0b8bc0407"
    name = "svm02"
  }
  qos_policy = {
    # id   = "beb86559-8d2d-11ef-9ca8-00a0b8bc0407"
    name = "test-svm02"
    # name = "performance-svm02"
  }
}
