resource "netapp-ontap_autosupport" "autosupport" {
  # required to know which system to interface with
  cx_profile_name = "cluster6"
  
  # AutoSupport configuration
  enabled          = true
  transport        = "smtp"
  to_addresses     = ["admin@example.com"]
  from             = "ontap@example.com"
  contact_support  = true
  mail_hosts       = ["smtp.example.com"]
  is_minimal       = false
  ondemand_enabled = true
  smtp_encryption  = "start_tls"
  proxy_url        = "test123.com"
  partner_addresses = ["partner1@example.com"]
  force            = false  # Set to true to force updates if needed
}
