data "netapp-ontap_protocols_cifs_services_data_source" "protocols_cifs_services" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  filter = {
    svm_name = "test3"
  }
}
