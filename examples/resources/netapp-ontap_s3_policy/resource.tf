# Example without statements
resource "netapp-ontap_s3_policy" "basic_policy" {
  cx_profile_name = "cluster5"
  name = "test_policy1"
  svm_name = "tf_acc_svm"
  comment = "S3 policy comment for basic policy"
}

# Example with statements including wildcard
resource "netapp-ontap_s3_policy" "full_policy" {
  cx_profile_name = "cluster5"
  name = "test_policy2"
  svm_name = "tf_acc_svm"
  comment = "S3 policy comment for full policy"
  
  statements = [
    {
      sid       = "AllowFullAccess"
      effect    = "allow"
      actions   = ["PutObject"]
      resources = ["bucket2/*"]
    },
    {
      sid       = "AllowWriteAccess"
      effect    = "allow" 
      actions   = ["*"]
      resources = ["bucket1/*", "bucket1", "bucket2/*"]
    },
    {
      sid       = "DenyAdminActions"
      effect    = "deny"
      actions   = ["DeleteBucket", "CreateBucket"]
      resources = ["*"]
    },
    {
      sid       = "DenyAdminActionsBucket2"
      effect    = "deny"
      actions   = ["CreateBucket", "DeleteBucket"]
      resources = ["bucket2/*"]
    }
  ]
}
