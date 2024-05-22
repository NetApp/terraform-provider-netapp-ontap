data "netapp-ontap_cluster_peer" "cluster_peer" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "ontapcluster"
}
