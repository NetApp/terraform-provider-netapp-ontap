package security_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSecurityLoginMessage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Creating security_login_messages test
			{
				Config:      testAccSecurityLoginMessageBasicClusterConfig(),
				ExpectError: regexp.MustCompile("create operation is not supported"),
			},
			// Import and read with cluster
			{
				ResourceName:  "netapp-ontap_security_login_message.example",
				ImportState:   true,
				ImportStateId: "cluster1",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_login_message.example", "scope", "cluster"),
				),
			},
			// Import and read with svm
			{
				ResourceName:  "netapp-ontap_security_login_message.example",
				ImportState:   true,
				ImportStateId: "svm5,cluster1",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_login_message.example", "scope", "svm"),
				),
			},
			// Update a option cannot tested in acc test
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSecurityLoginMessageBasicClusterConfig() string {
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
      name = "cluster1"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_login_message" "example" {
  cx_profile_name = "cluster1"
  message              = "Test cluster only \n message\n on the cluster"
  show_cluster_message = true
}`, host, admin, password)
}
