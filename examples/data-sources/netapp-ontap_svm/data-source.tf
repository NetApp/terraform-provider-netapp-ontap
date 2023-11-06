data "netapp-ontap_svm_data_source" "svm" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "ansibleSVM"
}