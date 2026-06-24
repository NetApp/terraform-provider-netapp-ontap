package name_services_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNameServicesUnixGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccNameServicesUnixGroupResourceBasicConfig("tf_acc_svm", "tf_acc_unix_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "name", "tf_acc_unix_group"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "id", "99"),
				),
			},
			// Update and read
			// update id
			{
				Config: testAccNameServicesUnixGroupResourceUpdateConfig("tf_acc_svm", "tf_acc_unix_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "name", "tf_acc_unix_group"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "id", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.#", "0"),
				),
			},
			// Update and read
			// add users
			{
				Config: testAccNameServicesUnixGroupResourceAddUsersConfig("tf_acc_svm", "tf_acc_unix_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "name", "tf_acc_unix_group"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "id", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.0", "unix_user1"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.1", "unix_user2"),
				),
			},
			// Update and read
			// update users - remove one user, add another
			{
				Config: testAccNameServicesUnixGroupResourceUpdateUsersConfig("tf_acc_svm", "tf_acc_unix_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "name", "tf_acc_unix_group"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "id", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.0", "unix_user1"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.1", "unix_user3"),
				),
			},
			// Update and read
			// update users - remove all users
			{
				Config: testAccNameServicesUnixGroupResourceRemoveUsersConfig("tf_acc_svm", "tf_acc_unix_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "name", "tf_acc_unix_group"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "id", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.#", "0"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_unix_group.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_unix_group", "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "name", "tf_acc_unix_group"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "id", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_group.example", "users.#", "0"),
				),
			},
		},
	})
}

func testAccNameServicesUnixGroupResourceBasicConfig(svmName string, name string) string {
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

resource "netapp-ontap_unix_group" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  id = 99
}`, host, admin, password, svmName, name)
}

func testAccNameServicesUnixGroupResourceUpdateConfig(svmName string, name string) string {
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

resource "netapp-ontap_unix_group" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  id = 100
}`, host, admin, password, svmName, name)
}

func testAccNameServicesUnixGroupResourceAddUsersConfig(svmName string, name string) string {
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

resource "netapp-ontap_unix_group" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  id = 100
  users = ["unix_user1", "unix_user2"]
}`, host, admin, password, svmName, name)
}

func testAccNameServicesUnixGroupResourceUpdateUsersConfig(svmName string, name string) string {
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

resource "netapp-ontap_unix_group" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  id = 100
  users = ["unix_user1", "unix_user3"]
}`, host, admin, password, svmName, name)
}

func testAccNameServicesUnixGroupResourceRemoveUsersConfig(svmName string, name string) string {
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

resource "netapp-ontap_unix_group" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  id = 100
  users = []
}`, host, admin, password, svmName, name)
}
