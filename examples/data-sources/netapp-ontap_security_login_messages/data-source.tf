data "netapp-ontap_security_login_messages" "security_login_messages" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    scope = "cluster"
  }
}
