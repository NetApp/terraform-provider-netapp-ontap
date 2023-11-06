data "netapp-ontap_storage_aggregates_data_source" "storage_aggregates" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "aggr*"
  }
}
