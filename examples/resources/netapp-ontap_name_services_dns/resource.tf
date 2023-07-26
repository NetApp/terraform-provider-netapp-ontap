resource "netapp-ontap_name_services_dns_resource" "name_services_dns" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  svm_name = "ansibleSVM_cifs"
  name_servers = ["1.1.1.1", "2.2.2.2"]
  dns_domains = ["foo.bar.com", "boo.bar.com"]
}
