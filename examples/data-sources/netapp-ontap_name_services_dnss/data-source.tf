data "netapp-ontap_dnss" "dnss" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    svm_name = "svm*"
    dns_domains = "netapp*"
    name_servers = "10.193.115*"
  }
}
