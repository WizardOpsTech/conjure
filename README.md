# conjure
CLI tool for generating common DevOps templates

# Bash script to generate framework for go cli tool in the future

go mod init $project_name
go get -u github.com/spf13/viper@latest
go get -u github.com/spf13/cobra@latest
cobra-cli init --viper

echo "cli project initilized. Start using cobra-cli add <command> to add commands."

template file types .tf .json .yaml
bundle types kubernetes terraform

future support - opentofu, containerfile

## Getting started

- Creating templates
- Creating bundles
- Listing
- Generating from templates
- Local template and bundle registry
- Remote template and bundle registry ( github )
- interactive vs noninteractive

## What next

phase 4 - values.yaml file support
generate bundles
testing