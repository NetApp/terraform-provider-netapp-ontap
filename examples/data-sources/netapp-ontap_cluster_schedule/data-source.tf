data "netapp-ontap_cluster_schedule" "cluster_schedule" {
  cx_profile_name = "cluster4"
  # name = "Application Templates ASUP Dump"
  name = "Balanced Placement Model Cache Update"
}
