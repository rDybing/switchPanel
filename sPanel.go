package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/gousb"
)

// keyMapT contains the Panel to Key definitions
type keymapT struct {
	Rotary [][]int // 0 through 4
	Rocker [][]int // 0 through 12
	Gear   [][]int // 0 and 1
}

func main() {
	var input string
	quit := false
	keymap, err := getKeymap()
	if err != nil {
		log.Fatalf("Could not open keymap file: %v\n", err)
	}
	fmt.Println("Keymap file loaded!")
	go initUSB(keymap)
	time.Sleep(1 * time.Second)
	fmt.Println("All good to go, type quit + return to exit!")
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
	keymap, err := loadKeyMap(fileNumber)
	if err != nil {
		return keymap, err
	}
	return keymap, err
}

// loadKeyMap loads up a given keymap from a JSON file
func loadKeyMap(in string) (keymapT, error) {
	var km keymapT
	if in == "" {
		in = "0"
	}
	fileName := "keys" + in + ".json"
	fmt.Println(fileName)
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		return km, fmt.Errorf("Error, could not load keymap file: %s\n%v", fileName, err)
	}
	if err := json.Unmarshal(f, &km); err != nil {
		return km, fmt.Errorf("Error, could not unmarshal %s: %v", fileName, err)
	}
	return km, nil
}

func initUSB(keymap keymapT) {
	// open up usb connection
	ctx := gousb.NewContext()
	defer ctx.Close()
	dev, err := ctx.OpenDeviceWithVIDPID(0x06a3, 0x0d67)
	defer dev.Close()
	if err != nil {
		log.Fatalf("OpenDevice failed: %v\n", err)
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
	for {
		buf := make([]byte, epIn.Desc.MaxPacketSize)
		inBytes, err := epIn.Read(buf)
		if err != nil {
			fmt.Printf("Read returned an error: %v\n", err)
		}
		if inBytes == 0 {
			log.Fatalf("IN endpoint 1 returned 0 bytes of data.\n")
		}
		fmt.Printf("Bytes received: %v\n", buf)
	}
}
