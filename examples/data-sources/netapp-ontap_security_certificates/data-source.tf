# retrieving certificates installed on a specific SVM
data "netapp-ontap_security_certificates" "security_certificates1" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    scope    = "svm"
    svm_name = "tfsvm"
  }
}

# retrieving all certificates installed at cluster-scope
data "netapp-ontap_security_certificates" "security_certificates2" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    scope = "cluster"
  }
}

# retrieving certificate using its common_name and type
data "netapp-ontap_security_certificates" "security_certificates3" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    common_name = "tfsvm"
    type        = "server"
  }
}