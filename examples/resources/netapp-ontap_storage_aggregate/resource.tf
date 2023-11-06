resource "netapp-ontap_storage_aggregate_resource" "example" {
  cx_profile_name = "cluster4"
  name = "test_aggr"
  node = "swenjun-vsim1"
  disk_count = 5
  disk_size = 1
  disk_size_unit= "gb"
  is_mirrored = false
  raid_type = "raid4"
  snaplock_type = "compliance"
  encryption = true
}
