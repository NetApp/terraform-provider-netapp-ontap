data "netapp-ontap_protocols_nfs_export_policy" "export_policy" {
  cx_profile_name = "cluster4"
  svm_name = "svm4"
  name = "default"
}