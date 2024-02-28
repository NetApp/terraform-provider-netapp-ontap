resource "netapp-ontap_protocols_cifs_service_resource" "protocols_cifs_service" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
  svm_name = "testsvm"
  ad_domain = {
    domain_name = "test.com"
    organizational_unit = "CN=Computers"
    password = "password"
    user = "user"
  }
  security = {
    advertised_kdc_encryptions = ["des", "rc4"]
    session_security = "none"
    lm_compatibility_level = "lm_ntlm_ntlmv2_krb"
    use_ldaps = false
    use_start_tls = false
    restrict_anonymous = "no_restriction"
    ldap_referral_enabled = false
    try_ldap_channel_binding = false
    smb_signing = false
    smb_encryption = false
    encrypt_dc_connection = false
    aes_netlogon_enabled = false
  }
  netbios = {
    aliases = ["ABC"]  
    wins_servers = ["1.3.7.8", "2.3.4.5"]
    enabled = false
  }
  comment = "test server commant"
  default_unix_user = "defaultuser"
  enabled = true
}
