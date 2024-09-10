data "netapp-ontap_volume_efficiency_policy" "volume_efficiency_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "test"
  svm = {
    name = "terraform"
  }
}
