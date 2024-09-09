data "netapp-ontap_nfs_export_policy_rule" "rule" {
  cx_profile_name = "cluster4"
  svm_name = "svm0"
  export_policy_name = "export_policy"
  index = 1
}