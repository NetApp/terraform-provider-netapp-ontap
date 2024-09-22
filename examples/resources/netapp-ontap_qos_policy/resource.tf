resource "netapp-ontap_qos_policy" "qos_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "terraform2"
  svm_name = "terraform"
  fixed = {
    max_throughput_iops = 1
  }
}
