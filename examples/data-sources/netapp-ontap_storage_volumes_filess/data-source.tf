data "netapp-ontap_storage_volumes_filess_data_source" "storage_volumes_filess" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  # filter = {}
}
