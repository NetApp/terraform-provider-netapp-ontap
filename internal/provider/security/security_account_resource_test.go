package security_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
)

func TestAccSecurityAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test updating a resource
			{
				Config: testAccSecurityAccountResourceConfigCreate("tf_acc_test", "password123"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "name", "tf_acc_test"),
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "password", "password123"),
				),
			},
			// Test updating a resource with comment and locked
			{
				Config: testAccSecurityAccountResourceConfig("tf_acc_test", "update", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "name", "tf_acc_test"),
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "comment", "update"),
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "locked", "true"),
				),
			},
			// Test updating a resource with application and secondAuthenticationMethod
			{
				Config: testAccSecurityAccountResourceConfigUpdateAndCheckIdempotency("tf_acc_test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "name", "tf_acc_test"),
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "applications.0.application", "http"),
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "applications.1.application", "ontapi"),
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "applications.2.application", "ssh"),
				),
			},
			// Test updating a resource with application and secondAuthenticationMethod
			{
				Config:             testAccSecurityAccountResourceConfigUpdateAndCheckIdempotency("tf_acc_test"),
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "name", "tf_acc_test"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_security_account.security_account",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "vsadmin", "tfsvm", "cluster2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account.security_account", "name", "vsadmin"),
				),
			},
		},
	})
}

func testAccSecurityAccountResourceConfig(name string, comment string, locked bool) string {
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
      name = "cluster2"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_account" "security_account" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "%s"
  applications = [{
    application = "http"
    authentication_methods = ["password"]
  }]
  comment = "%s"
  locked = %t
  password = "password123"
}
`, host, admin, password, name, comment, locked)
}

func testAccSecurityAccountResourceConfigUpdateAndCheckIdempotency(name string) string {
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
      name = "cluster2"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_account" "security_account" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "%s"
  applications = [{
    application = "ssh"
    second_authentication_method = "none"
    authentication_methods = ["password"],
  },
  {
    application = "ontapi"
    second_authentication_method = "none"
    authentication_methods = ["password"],
  },
  {
    application = "http"
    second_authentication_method = "none"
    authentication_methods = ["password"],
  }]
  password = "password123"
}
`, host, admin, password, name)
}

func testAccSecurityAccountResourceConfigCreate(name string, accpassword string) string {
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
      name = "cluster2"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_account" "security_account" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "%s"
  applications = [{
    application = "http"
    authentication_methods = ["password"]
  }]
  password = "%s"
}
`, host, admin, password, name, accpassword)
}
