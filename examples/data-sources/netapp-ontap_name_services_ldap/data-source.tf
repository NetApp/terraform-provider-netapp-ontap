data "netapp-ontap_name_services_ldap_data_source" "name_services_ldap" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  svm_name = "testme"
}
