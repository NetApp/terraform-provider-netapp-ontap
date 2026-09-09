# retrieving audit configuration for all SVMs in the cluster
data "netapp-ontap_svm_audits" "all_audit_configs" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
}

# retrieving audit configuration for matched SVMs
data "netapp-ontap_svm_audits" "matched_audit_configs" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm*"
  }
}
