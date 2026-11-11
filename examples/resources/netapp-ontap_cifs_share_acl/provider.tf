terraform {
  required_providers {
    netapp-ontap = {
      source = "NetApp/netapp-ontap"
      version = "0.0.1"
    }
  }
}


provider "netapp-ontap" {
  # A connection profile defines how to interface with an ONTAP cluster or svm.
  # At least one is required.
  connection_profiles = [
    {
      name = "cluster1"
      hostname = "********219"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster2"
      hostname = "********222"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster3"
      hostname = "********159"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster4"
      hostname = "********108"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster5"
      hostname = "********2"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "clustercifs"
      hostname = "********189"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "hw-cluster"
      hostname = "********50"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "aws-fsx"
      hostname = "********83"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
      aws_lambda = {
        function_name = var.aws_lambda_function_name
        region = var.aws_region
        shared_config_profile = var.aws_shared_config_profile
      }
    },
    {
      name = "gcnv-ontap-mode"
      google_netapp_unified_pool = {
        project_id = "1234567890"
        location = var.location
        storage_pool = var.storage_pool
        # custom_base_url = "https://netapp.googleapis.com/v1beta1"
      }
    }
  ]
}
