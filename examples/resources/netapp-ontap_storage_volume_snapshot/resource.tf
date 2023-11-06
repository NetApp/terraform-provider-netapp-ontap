resource "netapp-ontap_storage_volume_snapshot_resource" "example" {
  cx_profile_name = "cluster4"
  name = "snaptest"
  volume_name =  "carchi_test_root"
  svm_name = "carchi-test"
}
