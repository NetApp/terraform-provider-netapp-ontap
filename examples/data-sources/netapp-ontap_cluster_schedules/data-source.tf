data "netapp-ontap_cluster_schedules" "cluster_schedules" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
  filter = {
    type = "interval"
  }
}
