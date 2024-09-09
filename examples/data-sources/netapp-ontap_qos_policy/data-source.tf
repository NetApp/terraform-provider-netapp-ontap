data "netapp-ontap_qos_policy" "qos_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "test"
  svm_name = "terraform"
}
