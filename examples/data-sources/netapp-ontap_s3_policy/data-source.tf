data "netapp-ontap_s3_policy" "protocols_s3_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name = "test_policy_1"
  svm_name = "tf_acc_svm"
}
