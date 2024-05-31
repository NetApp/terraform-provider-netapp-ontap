resource "netapp-ontap_nfs_export_policy" "example" {
  cx_profile_name = "cluster4"
  svm_name = "carchi-test"
  name = "exportpolicytest"
}
