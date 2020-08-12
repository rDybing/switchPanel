package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/gousb"
	"github.com/micmonay/keybd_event"
)

// keyMapT contains the Panel to Key definitions
type keymapT struct {
	IsLinux  bool
	Rotary   []keyT // 0 through 4
	Switches []keyT // 0 through 12
	Gear     keyT
}

type keyT struct {
	Active bool
	KeyOn  []int
	KeyOff []int
}

func main() {
	fmt.Println("*********************************************************")
	fmt.Println("*                                                       *")
	fmt.Println("*         Logitech Switch Panel Generic Driver          *")
	fmt.Println("*                     Version: 1.0                      *")
	fmt.Println("*  Source Code: https://github.com/rDybing/switchPanel  *")
	fmt.Println("*            MIT License © 2020 Roy Dybing              *")
	fmt.Println("*                                                       *")
	fmt.Println("*********************************************************")

	keymap, err := getKeymap()
	if err != nil {
		log.Fatalf("Could not open keymap file: %v\n", err)
	}
	fmt.Println("Keymap file loaded!")
	keyBonding, err := keymap.initKeys()
	if err != nil {
		log.Fatalf("Could not initialize keyboard emulation")
	}
	go keymap.initUSB(keyBonding)
	fmt.Println("All good to go, type quit + return to exit!")
	var input string
	quit := false
	for !quit {
		fmt.Scanf("%s\n", &input)
		input = stripNewline(input)
		switch input {
		case "quit":
			quit = true
		}
	}
}

func stripNewline(in string) string {
	// strip newline
	in = strings.Replace(in, "\n", "", -1)
	// if on windows - also strip CR
	out := strings.Replace(in, "\r", "", -1)
	return out
}

func getKeymap() (keymapT, error) {
	// get any arguments deciding what file with keymap definitions to load
	var fileNumber string
	var keymap keymapT
	if len(os.Args[1:]) > 0 {
		fileNumber = os.Args[1]
		if _, err := strconv.Atoi(fileNumber); err != nil {
			return keymap, fmt.Errorf("Error, argument must be a number, eg. sPanel.exe 3: %v", err)
		}
	} else {
		fileNumber = "0"
	}
	err := keymap.loadKeyMap(fileNumber)
	if err != nil {
		return keymap, err
	}
	return keymap, err
}

// loadKeyMap loads up a given keymap from a JSON file
func (km *keymapT) loadKeyMap(in string) error {
	if in == "" {
		in = "0"
	}
	fileName := "keys" + in + ".json"
	fmt.Println(fileName)
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("Error, could not load keymap file: %s\n%v", fileName, err)
	}
	if err := json.Unmarshal(f, &km); err != nil {
		return fmt.Errorf("Error, could not unmarshal %s: %v", fileName, err)
	}
	return nil
}

func (km keymapT) initUSB(kb keybd_event.KeyBonding) {
	// open up usb connection
	ctx := gousb.NewContext()
	defer ctx.Close()
	dev, err := ctx.OpenDeviceWithVIDPID(0x06a3, 0x0d67)
	defer dev.Close()
	if err != nil || dev == nil {
		log.Fatalf("OpenDevice failed - ensure it is connected: %v\n", err)
	}
	fmt.Printf("Device opened: %v\n", dev)
	if err := dev.SetAutoDetach(true); err != nil {
		log.Fatalf("Could not detach device: %v", err)
	}
	fmt.Println("Device auto-detached")
	// grab device, interface and endpoint
	cfg, err := dev.Config(1)
	if err != nil {
		log.Fatalf("Opening %s.Config(1) failed: %v\n", dev, err)
	}
	fmt.Printf("Device config read: %v\n", cfg)
	defer cfg.Close()
	intf, err := cfg.Interface(0, 0)
	if err != nil {
		log.Fatalf("Opening %s.Interface(0, 0) failed: %v\n", cfg, err)
	}
	fmt.Println("Interface opened")
	defer intf.Close()
	epIn, err := intf.InEndpoint(1)
	if err != nil {
		log.Fatalf("Opening %s.InEndpoint(1) failed: %v\n", intf, err)
	}
	counter := 0
	for {
		buf := make([]byte, epIn.Desc.MaxPacketSize)
		inBytes, err := epIn.Read(buf)
		if err != nil {
			fmt.Printf("Read returned an error: %v\n", err)
		}
		if inBytes == 0 {
			log.Fatalf("IN endpoint 1 returned 0 bytes of data.\n")
		}
		var outBytes [3]uint8
		for i := 0; i < 3; i++ {
			outBytes[i] = uint8(buf[i])
		}
		km.getPanelSwitch(&kb, outBytes)
		counter++
	}
}

func (km *keymapT) getPanelSwitch(kb *keybd_event.KeyBonding, b [3]byte) {
	for i := uint(0); i < 8; i++ {
		// byte 0
		if b[0]&(1<<i) != 0 {
			km.setSwitchOn(kb, i)
		} else {
			km.setSwitchOff(kb, i)
		}
		// byte 1
		if b[1]&(1<<i) != 0 {
			if i < 5 {
				km.setSwitchOn(kb, i+8)
			} else {
				if !km.Rotary[i-5].Active {
					km.setRotary(kb, i-5)
				}
			}
		} else {
			if i < 5 {
				km.setSwitchOff(kb, i+8)
			}
		}
		// byte 2
		if b[2]&(1<<i) != 0 {
			// rotary pos 4
			if i == 0 {
				if !km.Rotary[3].Active {
					km.setRotary(kb, 3)
				}
			}
			// rotary pos 5
			if i == 1 {
				if !km.Rotary[4].Active {
					km.setRotary(kb, 4)
				}
			}
			// gear
			if i == 2 {
				km.setGearUp(kb)
			}
			if i == 3 {
				km.setGearDown(kb)
			}
		}
	}
}

func (km *keymapT) setRotary(kb *keybd_event.KeyBonding, i uint) {
	for j := range km.Rotary {
		if km.Rotary[j].Active {
			km.Rotary[j].Active = false
		}
	}
	if !km.Rotary[i].Active {
		km.Rotary[i].Active = true
		fmt.Printf("Rotary Position %d\n", i)
	}

}

func (km *keymapT) setSwitchOn(kb *keybd_event.KeyBonding, i uint) {
	if !km.Switches[i].Active {
		km.Switches[i].Active = true
		if err := pressKeys(kb, km.IsLinux, km.Switches[i].KeyOn); err != nil {
			log.Printf("Woopsie: %v\n", err)
		}
	}
}

func (km *keymapT) setSwitchOff(kb *keybd_event.KeyBonding, i uint) {
	if km.Switches[i].Active {
		km.Switches[i].Active = false
		if err := pressKeys(kb, km.IsLinux, km.Switches[i].KeyOff); err != nil {
			log.Printf("Woopsie: %v\n", err)
		}
	}
}

func (km *keymapT) setGearDown(kb *keybd_event.KeyBonding) {
	if !km.Gear.Active {
		km.Gear.Active = true
		if err := pressKeys(kb, km.IsLinux, km.Gear.KeyOff); err != nil {
			log.Printf("Woopsie: %v\n", err)
		}
	}
}

func (km *keymapT) setGearUp(kb *keybd_event.KeyBonding) {
	if km.Gear.Active {
		km.Gear.Active = false
		if err := pressKeys(kb, km.IsLinux, km.Gear.KeyOn); err != nil {
			log.Printf("Woopsie: %v\n", err)
		}
	}
}

func (km *keymapT) initKeys() (keybd_event.KeyBonding, error) {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return kb, fmt.Errorf("Could not create new key bonding: %v", err)
	}
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
		km.IsLinux = true
	} else {
		km.IsLinux = false
	}
	return kb, nil
}

func pressKeys(kb *keybd_event.KeyBonding, linux bool, keys []int) error {
	var key int
	if len(keys) > 1 {
		key = keys[1]
		if linux {
			switch keys[0] {
			// alt
			case 56:
				kb.HasALT(true)
			// l shift
			case 42:
				kb.HasSHIFT(true)
			// r shift
			case 54:
				kb.HasSHIFTR(true)
			// l ctrl
			case 29:
				kb.HasCTRL(true)
			// r ctrl
			case 97:
				kb.HasCTRLR(true)
			}
		} else {
			switch keys[0] {
			// alt
			case 4113:
				kb.HasALT(true)
			// l shift
			case 4255:
				kb.HasSHIFT(true)
			// r shift
			case 4256:
				kb.HasSHIFTR(true)
			// l ctrl
			case 4257:
				kb.HasCTRL(true)
			// r ctrl
			case 4258:
				kb.HasCTRLR(true)
			}
		}
	} else {
		key = keys[0]
	}
	kb.SetKeys(key)
	if err := kb.Launching(); err != nil {
		return fmt.Errorf("Could not press keys: %v", err)
	}
	kb.Clear()
	return nil
}
