data "netapp-ontap_protocols_san_lun-maps_data_source" "protocols_san_lun-mapss" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  # filter = {}
}
