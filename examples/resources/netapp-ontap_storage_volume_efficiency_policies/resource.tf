resource "netapp-ontap_volume_efficiency_policies" "storage_volume_efficiency_policies" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
  svm = {
    name = "terraform"
  }
  type = "scheduled"
  schedule = {
    name = "hourly"
  }
  duration = 5
  qos_policy = "background"
  comment = "test112"
}
