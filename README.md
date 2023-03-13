# temper

[![Latest Release](https://img.shields.io/github/release/taigrr/temper.svg?style=for-the-badge)](https://github.com/taigrr/temper/releases)
[![Software License](https://img.shields.io/badge/license-0BSD-blue.svg?style=for-the-badge)](/LICENSE)
[![Go ReportCard](https://goreportcard.com/badge/github.com/taigrr/temper?style=for-the-badge)](https://goreportcard.com/report/taigrr/temper)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](https://pkg.go.dev/github.com/taigrr/temper)

A zero-dependency library to read USB TEMPer thermometers on Linux.

## Configuration

On Linux you need to set up some udev rules to be able to access the device as
a non-root/regular user.
Edit `/etc/udev/rules.d/99-temper.rules` and add these lines:

```
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1a86", ATTRS{idProduct}=="e025", GROUP="plugdev", SYMLINK+="temper%n"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7401", GROUP="plugdev", SYMLINK+="temper%n"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7402", GROUP="plugdev", SYMLINK+="temper%n"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1130", ATTRS{idProduct}=="660c", GROUP="plugdev", SYMLINK+="temper%n"
```
Note that there are many versions of the TEMPer USB and your
`idVendor` and `idProduct` ATTRs may differ.

Make sure your user is part of the `plugdev` group and reload the rules with
`sudo udevadm control --reload-rules`.
Unplug and replug the device.

## Example Code

There are examples on how to use this repo in [examples/main.go](/examples/main.go)
