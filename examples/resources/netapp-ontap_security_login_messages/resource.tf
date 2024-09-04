
resource "netapp-ontap_security_login_messages" "msg_import_cluster" {
  banner               = "test banner"
  cx_profile_name      = "cluster4"
  message              = "test message"
  scope                = "cluster"
  show_cluster_message = true
}

resource "netapp-ontap_security_login_messages" "msg_import_svm" {
  banner               = "test banner"
  cx_profile_name      = "cluster4"
  message              = "test message"
  scope                = "svm"
  show_cluster_message = true
  svm_name             = "svm5"
}
