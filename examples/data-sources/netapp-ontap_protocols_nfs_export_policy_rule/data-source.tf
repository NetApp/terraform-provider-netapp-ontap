data "netapp-ontap_protocols_nfs_export_policy_rule_data_source" "rule" {
  cx_profile_name = "cluster4"
  svm_name = "svm0"
  export_policy_name = "export_policy"
  index = 1
}