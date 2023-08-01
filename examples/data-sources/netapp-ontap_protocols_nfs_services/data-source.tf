data "netapp-ontap_protocols_nfs_services_data_source" "protocols_nfs_services" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    svm_name = "ansibleV*"
  }
}
