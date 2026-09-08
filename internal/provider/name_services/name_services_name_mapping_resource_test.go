package name_services_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNameServicesNameMappingResource(t *testing.T) {
	svmName := "tf_acc_svm"
	direction := "unix_win"
	index := 4
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNameServicesNameMappingResourceConfig(svmName, direction, index, "test_pattern", "user", "10.254.101.112/28"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "svm_name", svmName),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "direction", direction),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "index", "4"),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "pattern", "test_pattern"),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "replacement", "user"),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "client_match", "10.254.101.112/28"),
				),
			},
			{
				Config: testAccNameServicesNameMappingResourceConfig(svmName, direction, index, "test_pattern_updated", "root", "10.254.101.112/28"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "svm_name", svmName),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "pattern", "test_pattern_updated"),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "replacement", "root"),
				),
			},
			{
				ResourceName:  "netapp-ontap_name_services_name_mapping.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%d,%s", svmName, direction, index, "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "svm_name", svmName),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "direction", direction),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_name_mapping.example", "index", "4"),
				),
			},
		},
	})
}

func testAccNameServicesNameMappingResourceConfig(svmName, direction string, index int, pattern, replacement, clientMatch string) string {
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

resource "netapp-ontap_name_services_name_mapping" "example" {
  cx_profile_name = "hw-cluster"
  svm_name        = "%s"
  direction       = "%s"
  index           = %d
  pattern         = "%s"
  replacement     = "%s"
  client_match    = "%s"
}
`, host, admin, password, svmName, direction, index, pattern, replacement, clientMatch)
}
