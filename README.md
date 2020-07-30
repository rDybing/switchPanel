# Switch Panel

A tiny little tool to enable the Logitech (Flight Sim) Switch Panel to work with any game, by intercepting the USB serial stream from it and translate into keystrokes.

For my own purposes specifically for use with:

- IL2
- Elite Dangerous

To enable this, when starting the app, it'll load in a config file containing the keystrokes it will output. By default it'll load in the config file 'keys0.json' but by setting an option on starting the app, any other keys file can be loaded. These files however must follow the format 'keys<number>.json'. So by the command 'panel -4', it'll load in the file 'keys4.json'

This should allow existing owners of the Logitech Switch Panel to use it with newer games that Logitech do not provide plug-ins for - like for instance the new Flight Simulator from Microsoft.

## Build

Will require libusb-1.0 or newer installed on the system. For Windows (or Mac) builds, please reference the guide from the gousb library README:

https://github.com/google/gousb

Linux should have this installed already, and if not is easily accesible from your distros package manager. Again reference the above link.

Non standard libraries used in this project (go get -v)

- github.com/google/gousb

---

## USB info

From probing the USB devices (in linux) I fished out this info regarding the Logitech Switch Panel:

```
001.005 06a3:0d67 Pro Flight Switch Panel (Saitek PLC)
  Protocol: (Defined at Interface level)
  Configuration 1:
  --------------
  Interface 0 alternate setting 0 (available endpoints: [0x81(1,IN)])
    Human Interface Device (No Subclass) None
    ep #1 IN (address 0x81) interrupt - undefined usage [63 bytes]
```

There should be also be an OUT endpoint, to control the landing gear status lights, but alas I've not found it.

---

**Contact:**

location   | name/handle
-----------|---------
github:    | rDybing
twitter:   | @DybingRoy
Linked In: | Roy Dybing

---

## Releases

- Version format: [major release].[new feature(s)].[bugfix patch-version]
- Date format: yyyy-mm-dd

#### v.1.0.0: Date to be determined
Details TBA

---

## License: MIT

Full license text found in LICENCE.md

**Copyright © 2020 Roy Dybing** 

---

ʕ◔ϖ◔ʔ