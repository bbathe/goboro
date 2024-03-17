# goboro

[![Codetest](https://github.com/bbathe/goboro/actions/workflows/codetest.yml/badge.svg)](https://github.com/bbathe/goboro/actions/workflows/codetest.yml) [![Release](https://github.com/bbathe/goboro/actions/workflows/release.yml/badge.svg)](https://github.com/bbathe/goboro/actions/workflows/release.yml)

## Description

App I use to manage O letter duties.  Currently supports looking up the email associated with a callsign on QRZ.com, creating email content and allowing the user to send it.

## Installation
To install this application:

1. Create the folder `C:\Program Files\goboro`
2. Download the `goboro_windows_amd64.zip` file from the [latest release](https://github.com/bbathe/goboro/releases) and unzip it into that folder
3. Double-click on the `goboro.exe` file to start the application and finish the configuration
4. Create a shortcut somewhere or pin to taskbar to make it easier to start in the future

You can have multiple configuration files and switch between them by using the `config` command line switch:
  ```yaml
  goboro.exe -config oletter.yaml
  ```