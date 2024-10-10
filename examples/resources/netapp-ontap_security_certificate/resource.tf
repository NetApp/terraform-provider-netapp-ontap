# creating a cluster-scoped certificate
resource "netapp-ontap_security_certificate" "create_certificate1" {
  cx_profile_name = "cluster5"
  name            = "test_ca_cert1"
  common_name     = "test_ca_cert"
  type            = "root_ca"
  expiry_time     = "P365DT"
}

# creating a certificate
resource "netapp-ontap_security_certificate" "create_certificate2" {
  cx_profile_name = "cluster5"
  name            = "tfsvm_ca_cert1"
  common_name     = "tfsvm_ca_cert"
  type            = "root_ca"
  svm_name        = "tfsvm"
  expiry_time     = "P365DT"
}

# signing a certificate
resource "netapp-ontap_security_certificate" "sign_certificate" {
  cx_profile_name = "cluster5"
  name            = "tfsvm_ca_cert1"
  common_name     = "tfsvm_ca_cert"
  type            = "root_ca"
  svm_name        = "svm1"  # SVM on which the signed certificate will exist
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
  common_name     = "svm1_cert1"
  type            = "server"
  svm_name        = "svm1"
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
