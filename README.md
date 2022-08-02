# Vivino Quick-Conf

## Purpose

Quick-Conf is a custom-built utility tool that allows you to quickly and easily pull, modify and push configuration files for go-api for non-production environments.
It's meant as an easy alternative to the Capistrano scripts that otherwise require a working Ruby installation and a not insignificant number of Ruby Gems.

While the Capistrano solution does the job just fine, Quick-Conf is an alternative for engineers who find it unwieldy to maintain a working Ruby installation when all they need from that eco-system is the ability to edit configuration files.

## Features

* Comes with unicreds built-in (you won't need to install unicreds separately)
* No installation requirements
* Launches your primary $EDITOR automatically for editing purposes
* Forwards the configuration file contents to stdout if live editing is not requested
* Pushes configuration file only in case of changes
* Supports "projects" defined in a simple TOML config

## Installation

* Method 1: `go get github.com/mkock/vivino_quickconf`, then run `make`
* Method 2: Download one of the binaries for Mac or Linux

## Usage

# Live edit

`qconf -project=<name> -edit`

locates the project with the given name in qconf.toml and uses unicreds to fetch the relevant configuration file from DynamoDB. Subsequently launches your editor and pushes the modified configuration file after editing.

# Pull and pipe to stdout

You don't have to edit the configuration file. Instead, you can pipe it to stdout (or elsewhere) without any pushes being made:

`qconf -project=<name> > testing.config`

# Pull to local file

You can pull a project into a local file (the file will be created for you):

`qconf -project=<name> -pull=<filename>`

# Push from local file

If you've modified a project locally, you can push the local file:

`qconf -project=<name> -push=<filename>`

You will get an error if the provided file does not exist, or is empty.

## Configuration

If your TOML file is not called qconf.toml, or if it's located elsewhere on your file system, you can tell qconf where to find it:

`qconf -conf=path/to/qconf.toml ...`

Alternatively, you can put it in `~/.config/qconf/qconf.toml`, and it will be automatically detected.

The qconf.toml file is *not* included with this source code! The reasons should be obvious. Please request this file via secure means from a Vivino colleague.
