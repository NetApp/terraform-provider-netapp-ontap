data "netapp-ontap_cluster" "cluster" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
}
