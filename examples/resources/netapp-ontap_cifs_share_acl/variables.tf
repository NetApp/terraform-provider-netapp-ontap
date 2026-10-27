# Terraform will prompt for values, unless a tfvars file is present.
variable "username" {
    type = string
    default = null
}
variable "password" {
    type = string
    sensitive = true
    default = null
}
variable "validate_certs" {
    type = bool
    default = false
}

# Alternate method: For certificate-based auth, see docs/guides/ssl_certificate_authentication.md
