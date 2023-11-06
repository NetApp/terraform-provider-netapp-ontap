data "netapp-ontap_name_services_dnss_data_source" "name_services_dnss" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    svm_name = "svm*"
    dns_domains = "netapp*"
    name_servers = "10.193.115*"
  }
}
