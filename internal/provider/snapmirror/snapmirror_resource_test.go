package snapmirror_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSnapmirrorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existant Vol
			{
				Config:      testAccSnapmirrorResourceBasicConfig("tf_peer:tf_acc_svm", "terraform:tf_acc_svm"),
				ExpectError: regexp.MustCompile("13304105"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSnapmirrorResourceBreakResyncOptions(t *testing.T) {
	sourcePath := os.Getenv("TF_ACC_SNAPMIRROR_SOURCE_PATH")
	destinationPath := os.Getenv("TF_ACC_SNAPMIRROR_DESTINATION_PATH")
	if sourcePath == "" || destinationPath == "" {
		t.Skip("set TF_ACC_SNAPMIRROR_SOURCE_PATH and TF_ACC_SNAPMIRROR_DESTINATION_PATH to run break/resync acceptance test")
	}

	resourceName := "netapp-ontap_snapmirror.example"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create relationship with initialize and custom timeout.
			{
				Config: testAccSnapmirrorResourceBreakResyncConfig(sourcePath, destinationPath, "snapmirrored", false, false, true, 600),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "source_endpoint.path", sourcePath),
					resource.TestCheckResourceAttr(resourceName, "destination_endpoint.path", destinationPath),
					resource.TestCheckResourceAttr(resourceName, "initialize", "true"),
					resource.TestCheckResourceAttr(resourceName, "transferring_time_out", "600"),
				),
			},
			// Break relationship with force=true.
			{
				Config: testAccSnapmirrorResourceBreakResyncConfig(sourcePath, destinationPath, "broken_off", true, false, true, 600),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "broken_off"),
					resource.TestCheckResourceAttr(resourceName, "force", "true"),
					resource.TestCheckResourceAttr(resourceName, "quick_resync", "false"),
				),
			},
			// Resync relationship with quick_resync=true.
			{
				Config: testAccSnapmirrorResourceBreakResyncConfig(sourcePath, destinationPath, "snapmirrored", false, true, true, 600),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "snapmirrored"),
					resource.TestCheckResourceAttr(resourceName, "quick_resync", "true"),
				),
			},
		},
	})
}

func testAccSnapmirrorResourceBasicConfig(sourceEndpoint string, destinationEndpoint string) string {
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
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_snapmirror" "example" {
  cx_profile_name = "cluster4"
  source_endpoint = {
    path = "%s"
  }
  destination_endpoint = {
    path = "%s"
  }
}
`, host, admin, password, sourceEndpoint, destinationEndpoint)
}

func testAccSnapmirrorResourceBreakResyncConfig(sourceEndpoint string, destinationEndpoint string, state string, force bool, quickResync bool, initialize bool, transferringTimeOut int) string {
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
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_snapmirror" "example" {
  cx_profile_name = "cluster4"
  source_endpoint = {
    path = "%s"
  }
  destination_endpoint = {
    path = "%s"
  }
  initialize            = %t
  state                 = "%s"
  force                 = %t
  quick_resync          = %t
  transferring_time_out = %d
}
`, host, admin, password, sourceEndpoint, destinationEndpoint, initialize, state, force, quickResync, transferringTimeOut)
}
