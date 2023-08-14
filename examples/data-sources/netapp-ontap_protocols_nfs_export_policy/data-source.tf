data "netapp-ontap_protocols_nfs_export_policy_data_source" "export_policy" {
  cx_profile_name = "cluster4"
  svm_name = "automation"
  name = "test"
}