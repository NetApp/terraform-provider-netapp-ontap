data "netapp-ontap_protocols_cifs_service" "protocols_cifs_service" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
}
