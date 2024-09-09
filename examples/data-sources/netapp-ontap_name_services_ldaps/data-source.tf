data "netapp-ontap_name_services_ldaps" "name_services_ldaps" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  # filter = {}
}
