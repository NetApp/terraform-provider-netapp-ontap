data "netapp-ontap_cluster_peer_data_source" "cluster_peers" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "ontapcluster"
}
