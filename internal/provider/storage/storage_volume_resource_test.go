package storage_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existent SVM
			{
				Config:      testAccStorageVolumeResourceConfig("non-existent", "terraformTest4"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Test bad volume name
			{
				Config:      testAccStorageVolumeResourceConfig("non-existent", "name-cant-have-dashes"),
				ExpectError: regexp.MustCompile("917888"),
			},
			// restore_to is update-only, throws error when used during create.
			{
				Config:      testAccStorageVolumeResourceConfigWithRestore("tf_acc_svm", "tf_acc_volume_restore_snapshot"),
				ExpectError: regexp.MustCompile("Invalid restore_to usage during create"),
			},
			// Create and read volume
			{
				Config: testAccStorageVolumeResourceConfig("tf_acc_svm", "tf_acc_volume_1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume.example", "name", "tf_acc_volume_1"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "style", "flexvol"),
					resource.TestCheckNoResourceAttr("netapp-ontap_volume.example", "volname"),
					resource.TestCheckResourceAttr("netapp-ontap_volume.example", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_volume.example", "tags.*", "env:test"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_volume.example", "tags.*", "tier:gold"),
				),
			},
      // Update volume
			{
				Config: testAccStorageVolumeResourceConfigUpdate("tf_acc_svm", "tf_acc_volume_1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume.example", "name", "tf_acc_volume_1"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "style", "flexvol"),
					resource.TestCheckResourceAttr("netapp-ontap_volume.example", "nas.group_id", "10"),
					resource.TestCheckNoResourceAttr("netapp-ontap_volume.example", "volname"),
					resource.TestCheckResourceAttr("netapp-ontap_volume.example", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_volume.example", "tags.*", "env:prod"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_volume.example", "tags.*", "owner:qa"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_volume.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_volume_1", "tf_acc_svm", "cluster5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume.example", "name", "tf_acc_volume_1"),
				),
			},
      // Create a snaplock volume
      {
        Config: testAccStorageVolumeResourceConfigSnaplock("tf_acc_svm", "tf_acc_volume_snaplock"),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "name", "tf_acc_volume_snaplock"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "snaplock.type", "compliance"),
        ),
      },
      // Update the snaplock volume
      {
        Config: testAccStorageVolumeResourceConfigSnaplockUpdate("tf_acc_svm", "tf_acc_volume_snaplock"),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "name", "tf_acc_volume_snaplock"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "snaplock.type", "compliance"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "snaplock.append_mode_enabled", "true"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "snaplock.autocommit_period", "P30M"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "snaplock.retention.default", "P1Y"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "snaplock.retention.maximum", "P20Y"),
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "snaplock.retention.minimum", "P1Y"),
        ),
      },
      // Import the snaplock volume
      {
        ResourceName:  "netapp-ontap_volume.example",
        ImportState:   true,
        ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_volume_snaplock", "tf_acc_svm", "cluster5"),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("netapp-ontap_volume.example", "name", "tf_acc_volume_snaplock"),
        ),
      },
		},
	})
}

func testAccStorageVolumeResourceConfig(svm, volName string) string {
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
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_volume" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = [
	{name = "NSOL_NetApp_A70_T19U05a_NVME_SSD_1"}
  ]
  encryption = true
  space_guarantee = "none"
  snapshot_policy = "default-1weekly"
  style = "flexvol"
  space = {
	size = 30
	size_unit = "mb"
	percent_snapshot_space = 10
    logical_space = {
      enforcement = true
      reporting = true
    }
  }
  tiering = {
  	policy_name = "none"
  }
  nas = {
    export_policy_name = "default"
    group_id = 1
    user_id = 2
    unix_permissions = "100"
    security_style = "mixed"
	  junction_path = "/testacc"
  }
  autosize = {
    minimum = 20
    maximum = 60
    shrink_threshold = 10
    grow_threshold = 90
    mode = "off"
    size_unit = "mb"
  }
  tags = ["env:test", "tier:gold"]
}`, host, admin, password, volName, svm)
}

// testAccStorageVolumeResourceConfigWithRestore is used to test restoring a volume from a snapshot.
// It is update-only because restore_to cannot be used during create.
func testAccStorageVolumeResourceConfigWithRestore(svm, volName string) string {
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
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_volume" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = [
	{name = "NSOL_NetApp_A70_T19U05a_NVME_SSD_1"}
  ]
  encryption = true
  space_guarantee = "none"
  snapshot_policy = "default-1weekly"
  space = {
	size = 30
	size_unit = "mb"
	percent_snapshot_space = 10
    logical_space = {
      enforcement = true
      reporting = true
    }
  }
  tiering = {
    policy_name = "none"
  }
  nas = {
    export_policy_name = "default"
    group_id = 1
    user_id = 2
    unix_permissions = "100"
    security_style = "mixed"
	  junction_path = "/testacc"
  }
  autosize = {
    minimum = 20
    maximum = 60
    shrink_threshold = 10
    grow_threshold = 90
    mode = "off"
    size_unit = "mb"
  }
  restore_to = {
    snapshot = {
      name = "snap1"
    }
  }
  snapshot_locking_enabled = true
}`, host, admin, password, volName, svm)
}

// testAccStorageVolumeResourceConfigUpdate updates percent_snapshot_space from 10 to 20
// and group_id from 1 to 10, and size from 30 to 40
func testAccStorageVolumeResourceConfigUpdate(svm, volName string) string {
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
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_volume" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = [
	{name = "NSOL_NetApp_A70_T19U05a_NVME_SSD_1"}
]
  encryption = true
  space_guarantee = "none"
  snapshot_policy = "default-1weekly"
  style = "flexvol"
  space = {
	size = 40
	size_unit = "mb"
	percent_snapshot_space = 20
    logical_space = {
      enforcement = true
      reporting = true
    }
  }
  tiering = {
  	policy_name = "none"
  }
  nas = {
    export_policy_name = "default"
    group_id = 10
    user_id = 20
    unix_permissions = "755"
    security_style = "mixed"
	junction_path = "/testacc"
  }
  autosize = {
    minimum = 25
    maximum = 65
    shrink_threshold = 15
    grow_threshold = 95
    mode = "grow"
    size_unit = "mb"
  }
  snapshot_locking_enabled = false
  tags = ["env:prod", "owner:qa"]
}`, host, admin, password, volName, svm)
}

// testAccStorageVolumeResourceConfigSnaplock creates a snaplock volume
func testAccStorageVolumeResourceConfigSnaplock(svm, volName string) string {
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
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_volume" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = [
	  {
      name = "NSOL_NetApp_A70_T19U05a_NVME_SSD_1"
    }
  ]
  encryption = true
  space = {
	  size = 30
	  size_unit = "mb"
	  junction_path = "/testacc"
  }
  snaplock = {
    type = "compliance"
  }
}`, host, admin, password, volName, svm)
}

// testAccStorageVolumeResourceConfigSnaplockUpdate updates the snaplock volume
func testAccStorageVolumeResourceConfigSnaplockUpdate(svm, volName string) string {
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
        name = "cluster5"
        hostname = "%s"
        username = "%s"
        password = "%s"
        validate_certs = false
      },
    ]
  }
  resource "netapp-ontap_volume" "example" {
    cx_profile_name = "cluster5"
    name = "%s"
    svm_name = "%s"
    aggregates = [
      {
        name = "NSOL_NetApp_A70_T19U05a_NVME_SSD_1"
      }
    ]
    encryption = true
    space = {
      size = 30
      size_unit = "mb"
      junction_path = "/testacc"
    }
    snaplock = {
      type = "compliance"
      append_mode_enabled = true
      autocommit_period = "P30M"
      retention = {
        default = "P1Y"
        maximum = "P20Y"
        minimum = "P1Y"
      }
    }
}`, host, admin, password, volName, svm)
}
