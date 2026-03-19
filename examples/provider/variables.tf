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
variable "client_cert_file" {
    type = string
    default = null
}
variable "client_key_file" {
    type = string
    default = null
}
variable "ca_cert_file" {
    type = string
    default = null
}