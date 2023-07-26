data "netapp-ontap_cluster_schedule_data_source" "cluster_schedule" {
  cx_profile_name = "cluster2"
  # name = "Application Templates ASUP Dump"
  name = "mylistcron"
}
