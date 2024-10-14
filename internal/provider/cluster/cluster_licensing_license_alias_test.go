package cluster_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
)

func TestLicensingLicenseResouceAlias(t *testing.T) {
	testLicense := os.Getenv("TF_ACC_NETAPP_LICENSE")
	name := "FCP"
	credName := "cluster4"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLicensingLicenseResourceConfigAlias("testme"),
				ExpectError: regexp.MustCompile("1115159"),
			},
			{
				Config: testAccLicensingLicenseResourceConfigAlias(testLicense),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_licensing_license_resource.cluster_licensing_license", "name", "insight_balance")),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_cluster_licensing_license_resource.cluster_licensing_license",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", name, credName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_licensing_license_resource.cluster_licensing_license", "name", "insight_balance")),
			},
		},
	})
}

func testAccLicensingLicenseResourceConfigAlias(key string) string {
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

resource "netapp-ontap_cluster_licensing_license_resource" "cluster_licensing_license" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  keys = ["%s"]
}
`, host, admin, password, key)
}
