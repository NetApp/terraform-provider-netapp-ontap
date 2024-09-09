data "netapp-ontap_security_accounts" "security_accounts" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "admin"
  }
}

data "netapp-ontap_security_accounts" "security_accounts2" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "a*"
  }
}

data "netapp-ontap_security_accounts" "security_accounts3" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "vsadmin"
  }
}

data "netapp-ontap_security_accounts" "security_accounts4" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "vsadmin"
    svm_name = "carchi-test"
  }
}

