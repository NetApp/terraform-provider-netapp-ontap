<!-- markdownlint-disable first-line-h1 no-inline-html -->
<!-- test -->
<a href="https://netapp.com">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/NTAP_BIG.D.png">
    <source media="(prefers-color-scheme: light)" srcset=".github/NTAP_BIG.png">
    <img src=".github/NTAP_BIG.png" alt="NetApp logo" title="NetApp" align="right" height="50">
  </picture>
</a>

# Terraform ONTAP Provider

[![Discord](https://img.shields.io/discord/855068651522490400)](https://discord.gg/NetApp)
![GitHub last commit (by committer)](https://img.shields.io/github/last-commit/netapp/terraform-provider-netapp-ontap)
![GitHub release (with filter)](https://img.shields.io/github/v/release/NetApp/terraform-provider-netapp-ontap)
![GitHub](https://img.shields.io/github/license/netapp/terraform-provider-netapp-ontap)

The ONTAP Provider allows Terraform to Manage NetApp ONTAP resources.

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) 1.4+
* [Go](https://golang.org/doc/install) 1.21+ (to build the provider plugin)

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Bug Reports & Feature Requests
Click on Issues, we have 6 categories for issues:
1. Report a Bug -- for unexpected error, a crash, or otherwise incorrect behavior.
2. Report a Documentation Error -- for error in our documentation, including typos, mistakes, or outdated information.
3. Request an Enhancement -- For new fields to existing Resources or data sources 
4. Request a New Resource, Data Source, or Service 
5. Other -- Any other issue that is not covered by the above categories.
6. General Questions -- If you have any general question we have a Discord channel for that.

## Contributing
Please read the [Contributing Guide](CONTRIBUTING.md) for details on the process for submitting pull requests to us.
