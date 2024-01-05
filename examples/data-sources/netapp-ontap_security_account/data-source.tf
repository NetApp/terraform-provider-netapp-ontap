data "netapp-ontap_security_account_data_source" "security_accounts" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  scope = "cluster"
  name = "admin"
}

data "netapp-ontap_security_account_data_source" "security_accounts2" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "admin"
}

data "netapp-ontap_security_account_data_source" "security_accounts3" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  owner = {
    name = "carchi-test"
  }
  name = "vsadmin"
}