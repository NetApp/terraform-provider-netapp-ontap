data "netapp-ontap_cluster_schedules_data_source" "cluster_schedules" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    type = "interval"
  }
}
