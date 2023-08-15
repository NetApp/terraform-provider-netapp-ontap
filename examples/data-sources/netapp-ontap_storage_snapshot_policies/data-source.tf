data "netapp-ontap_storage_snapshot_policies_data_source" "storage_snapshot_policies" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "ansible*"
  }
}
