package protocols_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccS3BucketResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccS3BucketResourceBasicConfig("tf_acc_svm", "tf-acc-s3-bucket1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "name", "tf-acc-s3-bucket1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "size", "102005473280"),
				),
			},
			// update and read
			// update policy, comment
			{
				Config: testAccS3BucketResourceUpdateConfig("tf_acc_svm", "tf-acc-s3-bucket1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "name", "tf-acc-s3-bucket1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "size", "102005473280"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "comment", "acc test bucket updated"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.sid", "AccessToUser"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.effect", "allow"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.principals.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.principals.0", "tf_acc_user1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.resources.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.resources.0", "*"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.actions.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "policy.statements.0.actions.0", "GetObject"),
				),
			},
			// update and read
			// add cors_rules
			{
				Config: testAccS3BucketResourceUpdateCORSConfig("tf_acc_svm", "tf-acc-s3-bucket1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "name", "tf-acc-s3-bucket1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "size", "102005473280"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "cors_rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("netapp-ontap_s3_bucket.example", "cors_rules.*", map[string]string{
						"max_age_seconds": "1",
					}),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.0.allowed_headers.*", "Content-Type"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.0.allowed_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.0.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.0.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.0.allowed_origins.*", "*"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.1.allowed_headers.*", "Content-Type"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.1.allowed_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.1.allowed_origins.*", "example.com"),
				),
			},
			// update and read
			// remove cors_rules
			{
				Config: testAccS3BucketResourceCORSConfigRemoveRule("tf_acc_svm", "tf-acc-s3-bucket1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "name", "tf-acc-s3-bucket1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "size", "102005473280"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "cors_rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("netapp-ontap_s3_bucket.example", "cors_rules.*", map[string]string{
						"max_age_seconds": "1",
					}),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.*.allowed_headers.*", "Content-Type"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.*.allowed_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.*.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.*.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_s3_bucket.example", "cors_rules.*.allowed_origins.*", "*"),
				),
			},
			// update and read
			// set cors_rules to empty list to remove all rules
			{
				Config: testAccS3BucketResourceDeleteCORSConfig("tf_acc_svm", "tf-acc-s3-bucket1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "name", "tf-acc-s3-bucket1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "size", "102005473280"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "cors_rules.#", "0"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_s3_bucket.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf-acc-s3-bucket", "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "name", "tf-acc-s3-bucket"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_bucket.example", "size", "102005473280"),

				),
			},
		},
	})
}

func testAccS3BucketResourceBasicConfig(svmName string, name string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
	{
	  name = "hw-cluster"
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  validate_certs = false
	},
  ]
}

resource "netapp-ontap_s3_bucket" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  size = 102005473280
}`, host, admin, password, svmName, name)
}

func testAccS3BucketResourceUpdateConfig(svmName string, name string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "hw-cluster"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_s3_bucket" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  size = 102005473280
  comment = "acc test bucket updated"
  policy = {
    statements = [
      {
        sid = "AccessToUser"
        resources = [
          "*",
        ]
        actions = [
          "GetObject"
        ]
        effect = "allow"
        conditions = [
          {
            operator = "string_like"
            usernames = [
              "user1"
            ]
          }
        ]
        principals = [
          "tf_acc_user1"
        ]
      }
    ]}
}`, host, admin, password, svmName, name)
}

func testAccS3BucketResourceUpdateCORSConfig(svmName string, name string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "hw-cluster"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_s3_bucket" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  size = 102005473280
  cors_rules = [
	{
	  allowed_headers = ["Content-Type"]
	  allowed_methods = ["GET","POST","DELETE"]
	  allowed_origins = ["*"]
	  max_age_seconds = 1
	},
	{
	  allowed_headers = ["Content-Type"]
      allowed_methods = ["GET"]
	  allowed_origins = ["example.com"]
      max_age_seconds = 1
	}
  ]
}`, host, admin, password, svmName, name)
}

func testAccS3BucketResourceCORSConfigRemoveRule(svmName string, name string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "hw-cluster"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_s3_bucket" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  size = 102005473280
  cors_rules = [
	{
	  allowed_headers = ["Content-Type"]
	  allowed_methods = ["GET","POST","DELETE"]
	  allowed_origins = ["*"]
	  max_age_seconds = 1
	}
  ]
}`, host, admin, password, svmName, name)
}

func testAccS3BucketResourceDeleteCORSConfig(svmName string, name string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "hw-cluster"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_s3_bucket" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  size = 102005473280
  cors_rules = []
}`, host, admin, password, svmName, name)
}
