data "netapp-ontap_security_login_message" "security_login_message" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "ansibleSVM"
}
