package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProtocolsFpolicyEventResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error - non-existent SVM
			{
				Config:      testAccProtocolsFpolicyEventResourceConfig("non-existant", "cifs", "acc_fpolicy_event_test"),
				ExpectError: regexp.MustCompile("svm non-existant not found"),
			},
			// Test error - invalid protocol
			{
				Config:      testAccProtocolsFpolicyEventResourceConfigInvalidProtocol("tf_acc_svm", "invalid_protocol", "acc_fpolicy_event_test"),
				ExpectError: regexp.MustCompile("Attribute protocol value must be one of"),
			},
			// Create and read testing - CIFS protocol
			{
				Config: testAccProtocolsFpolicyEventResourceConfig("tf_acc_svm", "cifs", "acc_fpolicy_event_cifs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "name", "acc_fpolicy_event_cifs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "protocol", "cifs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "file_operations.#", "5"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "filters.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "volume_monitoring", "false"),
				),
			},
			// Update and read - add file operations and filters
			{
				Config: testAccProtocolsFpolicyEventResourceConfigUpdate("tf_acc_svm", "cifs", "acc_fpolicy_event_cifs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "name", "acc_fpolicy_event_cifs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "protocol", "cifs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "file_operations.#", "7"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "filters.#", "4"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "volume_monitoring", "true"),
				),
			},
			// Update and read - remove file operations and filters
			{
				Config: testAccProtocolsFpolicyEventResourceConfigUpdateRemove("tf_acc_svm", "cifs", "acc_fpolicy_event_cifs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "name", "acc_fpolicy_event_cifs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "protocol", "cifs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "file_operations.#", "3"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "filters.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "volume_monitoring", "false"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_fpolicy_event.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_event", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "name", "acc_fpolicy_event_cifs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example", "svm_name", "tf_acc_svm"),
				),
			},
		},
	})
}

func TestAccProtocolsFpolicyEventResourceNFS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read testing - NFSv4 protocol
			{
				Config: testAccProtocolsFpolicyEventResourceConfigNFS("tf_acc_svm", "nfsv4", "acc_fpolicy_event_nfs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "name", "acc_fpolicy_event_nfs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "protocol", "nfsv4"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "file_operations.#", "5"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "volume_monitoring", "true"),
				),
			},
			// Update protocol to NFSv3
			{
				Config: testAccProtocolsFpolicyEventResourceConfigNFS("tf_acc_svm", "nfsv3", "acc_fpolicy_event_nfs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "name", "acc_fpolicy_event_nfs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "protocol", "nfsv3"),
				),
			},
			// Test importing NFS resource
			{
				ResourceName:  "netapp-ontap_fpolicy_event.example_nfs",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_event", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "name", "acc_fpolicy_event_nfs"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_event.example_nfs", "protocol", "nfsv3"),
				),
			},
		},
	})
}

func testAccProtocolsFpolicyEventResourceConfig(svmName string, protocol string, name string) string {
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

resource "netapp-ontap_fpolicy_event" "example" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	protocol = "%s"
	file_operations = ["create", "write", "rename", "delete", "close"]
	filters = ["close_with_read", "first_write"]
	volume_monitoring = false
}`, host, admin, password, svmName, name, protocol)
}

func testAccProtocolsFpolicyEventResourceConfigInvalidProtocol(svmName string, protocol string, name string) string {
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

resource "netapp-ontap_fpolicy_event" "example" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	protocol = "%s"
	file_operations = ["create", "write"]
}`, host, admin, password, svmName, name, protocol)
}

func testAccProtocolsFpolicyEventResourceConfigUpdate(svmName string, protocol string, name string) string {
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

resource "netapp-ontap_fpolicy_event" "example" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	protocol = "%s"
	file_operations = ["create", "write", "rename", "delete", "close", "open", "read"]
	filters = ["close_with_read", "first_write", "first_read", "open_with_write_intent"]
	volume_monitoring = true
}`, host, admin, password, svmName, name, protocol)
}

func testAccProtocolsFpolicyEventResourceConfigUpdateRemove(svmName string, protocol string, name string) string {
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

resource "netapp-ontap_fpolicy_event" "example" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	protocol = "%s"
	file_operations = ["create", "write", "delete"]
	filters = ["first_write"]
	volume_monitoring = false
}`, host, admin, password, svmName, name, protocol)
}

func testAccProtocolsFpolicyEventResourceConfigNFS(svmName string, protocol string, name string) string {
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

resource "netapp-ontap_fpolicy_event" "example_nfs" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	protocol = "%s"
	file_operations = ["create", "write", "read", "rename", "delete"]
	volume_monitoring = true
}`, host, admin, password, svmName, name, protocol)
}
