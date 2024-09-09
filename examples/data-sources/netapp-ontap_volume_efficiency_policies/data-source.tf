data "netapp-ontap_volume_efficiency_policies" "volume_efficiency_policies" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  filter = {
    svm = "terraform"
  }
}
