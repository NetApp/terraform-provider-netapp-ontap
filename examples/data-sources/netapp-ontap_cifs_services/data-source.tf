data "netapp-ontap_cifs_services" "protocols_cifs_services" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  filter = {
    svm_name = "test3"
  }
}
