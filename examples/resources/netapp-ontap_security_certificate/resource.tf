# creating a certificate
resource "netapp-ontap_security_certificate" "create_certificate" {
  cx_profile_name = "cluster5"
  name            = "svm2_ca_cert1_unique"
  common_name     = "svm2_ca_cert1"
  type            = "root_ca"
  svm_name        = "svm2"
  expiry_time     = "P365DT"
}

# signing a certificate
resource "netapp-ontap_security_certificate" "sign_certificate" {
  cx_profile_name = "cluster5"
  name            = "svm2_ca_cert1_unique"
  common_name     = "svm2_ca_cert1"
  type            = "root_ca"
  svm_name        = "svm3"
  expiry_time     = "P90DT"
  signing_request = <<-EOT
-----BEGIN CERTIFICATE REQUEST-----
signing-request
-----END CERTIFICATE REQUEST-----
EOT
}

# installing a certificate
resource "netapp-ontap_security_certificate" "install_certificate" {
  cx_profile_name = "cluster5"
  common_name     = "svm3_cert1"
  type            = "server"
  svm_name        = "svm3"
  expiry_time     = "P90DT"
  public_certificate = <<-EOT
-----BEGIN CERTIFICATE-----
certificate
-----END CERTIFICATE-----
EOT

  private_key = <<-EOT
-----BEGIN PRIVATE KEY-----
private-key
-----END PRIVATE KEY-----
EOT
}
