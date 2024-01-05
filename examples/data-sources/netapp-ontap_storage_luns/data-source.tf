data "netapp-ontap_storage_luns_data_source" "storage_luns" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  # filter = {}
}
