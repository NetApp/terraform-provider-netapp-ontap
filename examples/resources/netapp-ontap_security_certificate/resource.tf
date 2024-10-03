resource "netapp-ontap_security_certificate" "sign_certificate" {
  cx_profile_name = "cluster5"
  name            = "svm2_ca_cert1_unique"
  common_name     = "svm2_ca_cert1"
  type            = "root_ca"
  svm_name        = "svm3"
  expiry_time     = "P90DT"
  signing_request = <<-EOT
-----BEGIN CERTIFICATE REQUEST-----
MIICmjCCAYICAQAwFTETMBEGA1UEAxQKc3ZtM19jZXJ0MTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBALR4afBcQLJ0iK041/5Kt1/X4KxKB2g50Ap3PNJw
aPx5KH/0PpxqLM8qqu5nsFVrjNnuUJ+1mnMsKVrVQHPXQqhJvlBmDh8PcR+snQhD
XqU1C/LOdsT1B2f6ezwHsQ0/s1yRXwRnYvbEpnNcq5xGRcwF4UeYZWjdhTDou9FL
qL9zJ0FeQZ/mt21yh9pe2NOtcawFfciljEOa3fEuZu2AMpNis53V3siQRcygBzmK
yC+OuoHIh7BO5Sac1wV6XZOANSGdqdQ+OJUSmh347ArEOBBTxDGPHrH0FFL0kSok
5sCO05eln/JErULyaySsjLW9dzSduZoIGEPweotVqKGGfDECAwEAAaBAMD4GCSqG
SIb3DQEJDjExMC8wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMC
BggrBgEFBQcDATANBgkqhkiG9w0BAQsFAAOCAQEAMCmLaaAET6WrMBrXOsj1tfzi
5zFlSQXo72c2KgaIrYTJ/tDbXFGFpV4f7cKDI3CIBjh4GQ3cNtr9ktg6Aq4cZajr
2cSfIpwNIZZlU/UribZf3Y5F7zN6vxL/Kb31AjpSITyM+Q1hlK/1/w/DdMos7BBk
gGcgyIKyvvKkIea/ik8aJpLBuJIsQbdtQwl8KhgK+btFyOEPWw5BBTItyvYC5K24
rA3/jvzFWCBU5nArNYhaCQSFd/270eVgYewhB7jKs6TX+38uANilF79qv1lnJIyg
43NyspLhk/mfjdZwvPBhRO1IyAcDwgw5X6NshnquxYCvtVHD4qfsS8oXw+D/OQ==
-----END CERTIFICATE REQUEST-----
EOT
}