data "netapp-ontap_cluster_peers_data_source" "cluster_peers" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  # filter = {}
}
