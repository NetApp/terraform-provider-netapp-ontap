data "netapp-ontap_protocols_nfs_export_policies_data_source" "export_policies" {
  cx_profile_name = "cluster4"
  filter = {
    #name = "default"
    svm_name = "svm*"
  }
}