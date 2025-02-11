resource "netapp-ontap_svm_qos_policy_activation" "example" {
  cx_profile_name = "cluster4"
  svm = {
    name = "svm02"
  }
  qos_policy = {
    name = "performance-svm02"
  }
}
