package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProtocolsCIFSSharesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existant SVM
			{
				Config:      testAccProtocolsCIFSSharesResourceConfig("non-existant", "terraformTest4"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Read testing
			{
				Config: testAccProtocolsCIFSSharesResourceConfig("tfsvm", "acc_test_cifs_share"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "name", "acc_test_cifs_share"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "comment", "this is a comment"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "continuously_available", "false"),
				),
			},
			{
				Config: testAccProtocolsCIFSSharesResourceConfigUpdate("tfsvm", "acc_test_cifs_share"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "name", "acc_test_cifs_share"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "comment", "update comment"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "continuously_available", "true"),
				),
			},
			{
				Config: testAccProtocolsCIFSSharesResourceConfigUpdateAddACL("tfsvm", "acc_test_cifs_share"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "name", "acc_test_cifs_share"),
				),
			},
			{
				Config: testAccProtocolsCIFSSharesResourceConfigUpdateDeleteACL("tfsvm", "acc_test_cifs_share"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "name", "acc_test_cifs_share"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_cifs_shares.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "acc_test_cifs_shares_import", "tfsvm", "clustercifs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_shares.example", "name", "acc_test_cifs_share_import"),
				),
			},
		},
	})
}

func testAccProtocolsCIFSSharesResourceConfig(svm, shareName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "clustercifs"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_cifs_shares" "example" {
	cx_profile_name = "clustercifs"
  	name = "%s"
  	svm_name = "%s"
	path = "/acc_test_cifs_share_volume"
	acls = [
		{
	  		"permission": "read",
	  		"type": "windows",
	  		"user_or_group": "BUILTIN\\Administrators"
		}
	]
 	comment = "this is a comment"
}`, host, admin, password, shareName, svm)
}

func testAccProtocolsCIFSSharesResourceConfigUpdate(svm, volName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "clustercifs"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_cifs_shares" "example" {
  cx_profile_name = "clustercifs"
  name = "%s"
  svm_name = "%s"
  path = "/acc_test_cifs_share_volume"
  acls = [
	{
		"permission": "full_control",
		"type": "windows",
		"user_or_group": "BUILTIN\\Administrators"
  	}
  ]
  comment = "update comment"
  continuously_available = true
}`, host, admin, password, volName, svm)
}

func testAccProtocolsCIFSSharesResourceConfigUpdateAddACL(svm, volName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "clustercifs"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_cifs_shares" "example" {
  cx_profile_name = "clustercifs"
  name = "%s"
  svm_name = "%s"
  path = "/acc_test_cifs_share_volume"
  acls = [
	  {
			"permission": "read",
			"type": "windows",
			"user_or_group": "Everyone"
	  },
	  {
		"permission": "full_control",
		"type": "windows",
		"user_or_group": "BUILTIN\\Administrators"
  	}
  ]
  comment = "update comment"
  continuously_available = true
}`, host, admin, password, volName, svm)
}

func testAccProtocolsCIFSSharesResourceConfigUpdateDeleteACL(svm, volName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "clustercifs"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_cifs_shares" "example" {
  cx_profile_name = "clustercifs"
  name = "%s"
  svm_name = "%s"
  path = "/acc_test_cifs_share_volume"
  acls = [
	  {
			"permission": "read",
			"type": "windows",
			"user_or_group": "Everyone"
	  }
  ]
  comment = "update comment"
  continuously_available = true
}`, host, admin, password, volName, svm)
}
