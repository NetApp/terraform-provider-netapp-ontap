data "netapp-ontap_aggregates" "storage_aggregates" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "aggr*"
  }
}
