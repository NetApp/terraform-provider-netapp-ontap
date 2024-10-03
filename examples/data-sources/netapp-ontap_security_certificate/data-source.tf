# retrieving a certificate using its unique name
data "netapp-ontap_security_certificate" "security_certificate1" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name            = "tfsvm_17B9B4C1696136FC"
}
