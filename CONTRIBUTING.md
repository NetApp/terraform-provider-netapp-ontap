# Contributing
Thank you  for your interest in contributing to the Terraform Provider for ONTAP! 🎉

We appreciate that you want to take the time to contribute! Please follow these steps before submitting your PR.

# Creating a Pull Request
1. Please search [existing issues](https://github.com/NetApp/terraform-provider-netapp-ontap/issues) to determine if an issue already exists for what you intend to contribute.
2. If the issue does not exist, [create a new one](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/new/choose) that explains the bug or feature request.
3. Let us know in the issue that you plan on creating a pull request for it, by selecting the option when creating an issue. This helps us to keep track of the pull request and make sure there isn't duplicate effort.
4. It's better to wait for feedback from someone on Netapp's Terraform Team development team before writing code. We don't have an SLA for our feedback, but we will do our best to respond in a timely manner (at a minimum, to give you an idea if you're on the right track and that you should proceed, or not).
5. Sign and submit [NetApp's Contributor License Agreement (CCLA)](https://netapp.tap.thinksmart.com/prod/Portal/ShowWorkFlow/AnonymousEmbed/3d2f3aa5-9161-4970-997d-e482b0b033fa). You must sign and submit the Corporate Contributor License Agreement (CCLA) in order to contribute.
* For Project name, select "Terraform Provider for ONTAP"
* For Project Website, enter "https://github.com/NetApp/terraform-provider-netapp-ontap"
6. If you've made it this far, have written the code that solves your issue, and addressed the review comments, then feel free to create your pull request.

Important: NetApp will NOT look at the PR or any of the code submitted in the PR if the CCLA is not on file with NetApp Legal.

## Requirement for new Resources/Data Sources
If your building a new Resouces or Data Source for the ONTAP we have a few requirements that need to be met before we can accept the PR.
* documentation in the /docs directory
* example in the /examples directory
* ACCtest in the /internal/provider directory
  * we have a GitHub Self Hosted Action that will run the ACCtest on an internal ONTAP VSIM, if there is anything that need to be set up for the ACCtest to run please let us know in the PR.
* Pass all existing GitHub actions test


# Netapp's Terraform Team's Commitment
While we truly appreciate your efforts on pull requests, we cannot make any commitment for when, or if, your PR will be included in the Trident project. Here are a few reasons why:
* There are many factors involved in integrating new code into this project, including things like: support for a wide variety of NetApp backends, proper adherence to our existing and/or upcoming architecture, sufficient functional and/or scenario tests across all backends, etc. In other words, while your bug fix or feature may be perfect as a standalone patch, we must ensure that your code works and can be accomodated for in all of our use cases, configurations, backends, installer, support matrix, etc.
* The Terraform Team must plan our resources to integrate your code into our code base and CI platform, and depending on the complexity of your PR, we may or may not have the resources available to make it happen in a timely fashion. As much as we would like to, and as much as we like to be "agile", we still have a roadmap of features that comes from Product Management. We will do our best, but it doesn't always happen quickly.
* Sometimes the PR just doesn't fit into our future plans. As stated above, we have a roadmap of things, and you don't necessarily know what's coming down the pike. It's possible that a PR you submit doesn't align with our upcoming plans, thus we won't be able to use it. It's not personal.

Thank you for considering to contribute to the Terraform Provider for ONTAP!

