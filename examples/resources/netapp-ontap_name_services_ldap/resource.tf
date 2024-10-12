resource "netapp-ontap_name_services_ldap" "name_services_ldap" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  svm_name = "svm5"
  servers = ["1.2.3.4", "5.6.7.8"]
  query_timeout = 5
  skip_config_validation = true
}
