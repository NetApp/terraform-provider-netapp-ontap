data "netapp-ontap_storage_volume_snapshot_data_source" "snapshot" {
  cx_profile_name = "cluster3"
  name = "weekly.2022-11-06_0015"
  volume_uuid = "4c3835b8-5aa9-11ed-817f-005056b3aeb3"
}
