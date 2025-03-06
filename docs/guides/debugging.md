---
page_title: "Enable Debugging in Terraform"
subcategory: ""
description: |-
---

# Enable Debugging in Terraform

When working with Terraform, you might encounter errors or unexpected behavior. To help diagnose and troubleshoot these issues, you can enable detailed logging by setting the `TF_LOG` environment variable.

## Setting the `TF_LOG` Environment Variable

The `TF_LOG` environment variable controls the verbosity of Terraform's logs. You can set it to different levels depending on the amount of detail you need. The available log levels are:

- `TRACE`: Very detailed logs, including internal Terraform operations.
- `DEBUG`: Detailed logs useful for debugging.
- `INFO`: General information about Terraform operations.
- `WARN`: Warnings about potential issues.
- `ERROR`: Only error messages.

### How to Set `TF_LOG`

All you need to do is set the environment variable while running Terraform operations, such as `plan`, `apply`, or `destroy`.

```terraform
TF_LOG=debug terraform apply
```

#### On Linux or macOS
You can set the `TF_LOG` environment variable in your terminal session:

```sh
export TF_LOG=DEBUG
```

To make this change persistent across terminal sessions, you can add the above line to your shell’s configuration file (e.g., .bashrc, .zshrc).

#### On Windows
You can set the `TF_LOG` environment variable in the Command Prompt or PowerShell:

```cmd
set TF_LOG=DEBUG
```

To make this change persistent, you can add the environment variable through the System Properties:

Once `TF_LOG` is set, run your Terraform commands as usual. The logs will be printed to the standard output (stdout). You can redirect these logs to a file for easier analysis:

```terraform
terraform apply > terraform.log
```

### Disabling Logging

After you have finished debugging, you can disable logging by unsetting the `TF_LOG` environment variable:

#### On Linux or macOS
```sh
unset TF_LOG
```

#### On Windows
```cmd
set TF_LOG=
```

### Best Practices for Using `TF_LOG`

- **Use Appropriate Levels**: Start with `DEBUG` for most issues, and only escalate to `TRACE` when you need the most detailed logs.
- **Turn Off Logging After Use**: Once you’ve debugged the issue, be sure to unset `TF_LOG` to avoid unnecessary log output in future operations.
- **Use Logs in Conjunction with Docs**: Always refer back to [Terraform's documentation](https://developer.hashicorp.com/terraform/plugin/log/managing) alongside the logs for a better understanding of the processes and errors you're seeing.

## Using the `TF_LOG_PATH` Environment Variable

To save the debug output to a file instead of displaying it in the console, you can use the `TF_LOG_PATH` environment variable to specify a file path for the logs:

```sh
export TF_LOG_PATH=./terraform-debug.log
```

This command directs the log output to a file named `terraform-debug.log` in the current directory, making it easier to review and share logs.
