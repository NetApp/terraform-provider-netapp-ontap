# retrieving all S3 server configurations
data "netapp-ontap_s3_services" "all_s3_servers" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
}

# retrieving S3 server configurations for matched SVMs
data "netapp-ontap_s3_services" "matched_s3_servers" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm*"
  }
}
