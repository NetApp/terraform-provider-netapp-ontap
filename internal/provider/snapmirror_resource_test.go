package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSnapmirrorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existant Vol
			{
				Config:      testAccSnapmirrorResourceBasicConfig("tf_peer:testme", "terraform:testme"),
				ExpectError: regexp.MustCompile("6619337"),
			},
			// Create snapmirror and read
			{
				Config: testAccSnapmirrorResourceBasicConfig("tf_peer:snap_source2", "terraform:snap_dest2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_resource.example", "destination_endpoint.path", "terraform:snap_dest2"),
				),
			},
			// Update a policy
			{
				Config: testAccSnapmirrorResourceUpdateConfig("tf_peer:snap_source", "terraform:snap_dest", "MirrorAndVault"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_resource.example", "policy.name", "MirrorAndVault"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_resource.example", "destination_endpoint.path", "terraform:snap_dest"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_snapmirror_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "terraform:snap_dest", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_resource.example", "destination_endpoint.path", "terraform:snap_dest"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSnapmirrorResourceBasicConfig(sourceEndpoint string, destinationEndpoint string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_snapmirror_resource" "example" {
  cx_profile_name = "cluster4"
  source_endpoint = {
    path = "%s"
  }
  destination_endpoint = {
    path = "%s"
  }
}`, host, admin, password, sourceEndpoint, destinationEndpoint)
}

func testAccSnapmirrorResourceUpdateConfig(sourceEndpoint string, destinationEndpoint string, policy string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_snapmirror_resource" "example" {
  cx_profile_name = "cluster4"
  source_endpoint = {
    path = "%s"
  }
  destination_endpoint = {
    path = "%s"
  }
  policy = {
	name = "%s"
  }
}`, host, admin, password, sourceEndpoint, destinationEndpoint, policy)
}
