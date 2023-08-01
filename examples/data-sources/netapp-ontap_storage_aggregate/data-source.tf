data "netapp-ontap_storage_aggregate_data_source" "storage_aggregate" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "aggr1"
}
