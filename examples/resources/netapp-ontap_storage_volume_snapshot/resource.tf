resource "netapp-ontap_storage_volume_snapshot_resource" "example" {
  cx_profile_name = "cluster4"
  name = "snaptest"
  volume = {
    name = "carchi_test_root"
  }
  svm = {
    name = "carchi-test"
  }
}
