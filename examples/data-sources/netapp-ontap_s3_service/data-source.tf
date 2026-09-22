# retrieving S3 server configurationf for specified SVM
data "netapp-ontap_s3_service" "s3_server_config" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "svm1"
}
