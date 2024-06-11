data "netapp-ontap_cluster_schedules" "cluster_schedules" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    type = "interval"
  }
}
