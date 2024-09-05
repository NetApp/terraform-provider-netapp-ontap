data "netapp-ontap_security_certificates" "security_certificates1" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    scope = "svm"
    svm_name = "tfsvm"
  }
}

data "netapp-ontap_security_certificates" "security_certificates2" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    scope = "cluster"
  }
}
