data "netapp-ontap_nfs_service" "protcols_nfs_services" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  svm_name = "ansibleSVM"
}
