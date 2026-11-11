data "netapp-ontap_svm_audit" "svm_audit_config" {
  cx_profile_name = "hw-cluster"
  svm_name = "svm1"
}
