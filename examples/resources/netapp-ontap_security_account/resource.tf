resource "netapp-ontap_security_account_resource" "security_account" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "carchitest"
  applications = [{
    application = "http"
    authentication_methods = ["password"]
  }]
  password = "P@ssw0rd"
}
