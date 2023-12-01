data "netapp-ontap_security_account_data_source" "security_accounts" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  owner = {
    name = "ansibleSVM"
  }
  name = "vsadmin"
}
