data "netapp-ontap_name_services_dns" "name_services_dns" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "svm0"
}
