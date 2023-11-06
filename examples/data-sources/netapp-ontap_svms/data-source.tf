data "netapp-ontap_svms_data_source" "svms" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
      name = "ansibleSVM"
  }
}
