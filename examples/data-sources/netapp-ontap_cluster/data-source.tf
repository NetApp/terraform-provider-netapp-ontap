data "netapp-ontap_cluster_data_source" "cluster" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
}
