data "netapp-ontap_snapshot_policies" "storage_snapshot_policies" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "ansible*"
  }
}
