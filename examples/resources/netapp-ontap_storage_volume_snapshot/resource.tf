resource "netapp-ontap_storage_volume_snapshot_resource" "example" {
  cx_profile_name = "cluster3"
  name = "carchi-test-snapshot2"
  volume_uuid = "4c3835b8-5aa9-11ed-817f-005056b3aeb3"

}
