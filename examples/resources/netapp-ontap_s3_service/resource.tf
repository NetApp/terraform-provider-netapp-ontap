resource "netapp-ontap_s3_service" "s3_server_config" {
  cx_profile_name = "hw-cluster"
  svm_name = "svm1"
  name = "svm1.example.com"
  enabled = false
  comment = "not enabled"
  certificate_name = "svm1_1888D5C7BA7A2302"
}
