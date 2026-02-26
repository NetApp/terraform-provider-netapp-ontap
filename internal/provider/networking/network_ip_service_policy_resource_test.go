package networking_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccServicePolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccServicePolicyResourceBasicConfig("tf_acc_svm", "tf_acc_ip_service_policy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "name", "tf_acc_ip_service_policy"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "scope", "svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "ipspace", "Default"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "services.#", "3"),
				),
			},
			// update and read
			// update services
			{
				Config: testAccServicePolicyResourceUpdateConfig("tf_acc_svm", "tf_acc_ip_service_policy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "name", "tf_acc_ip_service_policy"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "scope", "svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "ipspace", "Default"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "services.#", "0"),
				),
			},
			// update and read
			// rename
			{
				Config: testAccServicePolicyResourceUpdateConfig("tf_acc_svm", "tf_acc_service_policy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "name", "tf_acc_service_policy"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "scope", "svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "ipspace", "Default"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "services.#", "0"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_network_ip_service_policy.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_service_policy", "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "name", "tf_acc_service_policy"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "scope", "svm"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "ipspace", "Default"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_service_policy.example", "services.#", "0"),
				),
			},
		},
	})
}

func testAccServicePolicyResourceBasicConfig(svmName string, name string) string {
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

resource "netapp-ontap_network_ip_service_policy" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  services = [
    "data_nfs",
    "data_cifs",
    "management_ssh"
  ]
  ipspace = "Default"
}`, host, admin, password, svmName, name)
}

func testAccServicePolicyResourceUpdateConfig(svmName string, name string) string {
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

resource "netapp-ontap_network_ip_service_policy" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  services = []
  ipspace = "Default"
}`, host, admin, password, svmName, name)
}
