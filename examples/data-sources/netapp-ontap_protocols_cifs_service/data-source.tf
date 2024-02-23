data "netapp-ontap_protocols_cifs_service_data_source" "protocols_cifs_service" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
}
