# temper

[![Latest Release](https://img.shields.io/github/release/taigrr/temper.svg?style=for-the-badge)](https://github.com/taigrr/temper/releases)
[![Software License](https://img.shields.io/badge/license-0BSD-blue.svg?style=for-the-badge)](/LICENSE)
[![Go ReportCard](https://goreportcard.com/badge/github.com/taigrr/temper?style=for-the-badge)](https://goreportcard.com/report/taigrr/temper)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](https://pkg.go.dev/github.com/taigrr/temper)

A zero-dependency library to read USB TEMPer thermometers on Linux.

## Installation

Make sure you have a working Golang environment (Go 1.12 or higher is required).
See the [install instructions](http://golang.org/doc/install.html).

To install temper-cli, simply run:
	`go get github.com/taigrr/temper/cli`

## Configuration

On Linux you need to set up some udev rules to be able to access the device as
a non-root/regular user.
Edit `/etc/udev/rules.d/99-temper.rules` and add these lines:

```
ACTION=="add", SUBSYSTEM=="hidraw", ATTR{idVendor}=="0c45", ATTR{idProduct}=="7401", MODE:="666", GROUP="plugdev", SYMLINK+="temper%n"
```
Note that there are many versions of the TEMPer USB and your
`idVendor` and `idProduct` ATTRs may differ.

Make sure your user is part of the `plugdev` group and reload the rules with
`sudo udevadm control --reload-rules`.
Unplug and replug the device.


