data "netapp-ontap_san_lun-maps" "protocols_san_lun-mapss" {
  # required to know which system to interface with
  cx_profile_name = "fsx"
  # filter = {}
}
