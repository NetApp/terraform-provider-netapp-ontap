# Terraform will prompt for values, unless a tfvars file is present.
variable "username" {
    type = string
}
variable "password" {
    type = string
    sensitive = true
}
variable "validate_certs" {
    type = bool
}

# Alternate method: For certificate-based auth, see docs/guides/ssl_certificate_authentication.md
