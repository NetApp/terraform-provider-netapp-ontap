data "netapp-ontap_dns" "dns" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "svm0"
}
