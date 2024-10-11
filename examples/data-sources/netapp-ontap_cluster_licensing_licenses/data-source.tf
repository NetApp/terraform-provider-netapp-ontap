data "netapp-ontap_cluster_licensing_licenses" "cluster_licensing_licenses" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  #filter = {
  #  name = "snapmirror_sy*"
  #}
}
