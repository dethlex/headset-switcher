package main

import (
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dethlex/headset-switcher/headset"
	"github.com/dethlex/headset-switcher/icons"
	"github.com/slytomcat/systray"
)

const (
	dIndexMac = iota + 1
	dIndexName
)

const (
	sIndexSink = iota
	sIndexMac
	sIndexState
	sIndexCount
)

type devices struct {
	menuItem *systray.MenuItem
	headsets headset.HeadsetMap
}

func (d *devices) Reset() {
	d.headsets = make(headset.HeadsetMap)
	d.menuItem.RemoveSubmenu()
}

func (d *devices) Add(name string, h *headset.Headset) {
	d.headsets[name] = h
	d.menuItem.AddSubmenuItem(name, false)
}

func (d *devices) Get(name string) *headset.Headset {
	return d.headsets[name]
}

func (d *devices) Len() int {
	return len(d.headsets)
}

type selectedDevice struct {
	menuItem *systray.MenuItem
	headset  *headset.Headset
}

type menu struct {
	selectedDevice selectedDevice
	devices        devices
	refresh        *systray.MenuItem // TODO: make autorefresh
	exit           *systray.MenuItem
}

func (m *menu) selectDevice(name string) {
	d := m.devices.Get(name)
	if d == nil {
		return
	}
	m.selectedDevice.menuItem.SetTitle(d.GetStateName())
	m.selectedDevice.menuItem.Enable()
	SetIcon(d.GetState())
	m.selectedDevice.headset = d
}

func (m *menu) refreshHeadsets() {
	m.devices.Reset()

	out, err := exec.Command("bluetoothctl", "paired-devices").Output()

	if err != nil {
		log.Fatal(err)
	}

	// iterate over all founded paired devices
	for _, dev := range strings.Split(string(out), "\n") {

		if dev == "" {
			continue
		}

		dInfo := strings.Split(dev, " ")
		dName := dInfo[dIndexName]
		dMac := dInfo[dIndexMac]

		device := headset.NewHeadset(dName, dMac)

		exist, state := evalState(device.GetSinkName())

		if !exist {
			continue
		}

		device.SetState(state)

		m.devices.Add(dName, device)

		// set defaut device (basically first)
		if m.selectedDevice.headset == nil {
			m.devices.menuItem.Enable()
			m.selectDevice(dName)
		}
	}

	if m.devices.Len() == 0 {
		m.selectedDevice.headset = nil
		m.selectedDevice.menuItem.SetTitle("No device found")
		m.selectedDevice.menuItem.Disable()
		m.devices.menuItem.Disable()
		SetIcon(headset.SUnknown)
	}

}

func (m *menu) changeState() {
	if m.selectedDevice.headset == nil {
		return
	}
	var newState headset.State
	if m.selectedDevice.headset.GetState() == headset.SSpeak {
		newState = headset.SListen
	} else if m.selectedDevice.headset.GetState() == headset.SListen {
		newState = headset.SSpeak
	} else {
		log.Fatal("Unknown state")
		return
	}
	_, err := exec.Command("pacmd", "set-card-profile", m.selectedDevice.headset.GetCardName(), newState.String()).Output()
	if err != nil {
		log.Fatal(err)
		return
	}

	SetIcon(newState)
	m.selectedDevice.headset.SetState(newState)
	m.selectedDevice.menuItem.SetTitle(m.selectedDevice.headset.GetStateName())
}

func SetIcon(state headset.State) {
	switch state {
	case headset.SSpeak:
		systray.SetIcon(icons.IconSpeak)
	case headset.SListen:
		systray.SetIcon(icons.IconListen)
	default:
		systray.SetIcon(icons.IconDisabled)
	}
}

func main() {
	systray.Run(onRun, onExit)
}

//TODO: add checkers for system commands
func onRun() {
	log.Println("Run")
	icons.CreateIcons()
	SetIcon(headset.SUnknown)

	var m menu

	// main device
	m.selectedDevice.menuItem = systray.AddMenuItem("No device found", "Click to change state")
	m.selectedDevice.menuItem.Disable()

	systray.AddSeparator()

	// list of paired headphones
	m.devices.menuItem = systray.AddMenuItem("\u200B\u2060"+"Paired devices", "")
	m.devices.menuItem.Disable()

	m.refresh = systray.AddMenuItem("Refresh", "")

	m.refreshHeadsets()

	systray.AddSeparator()

	m.exit = systray.AddMenuItem("Exit", "Exit from program")

	go func() {
		defer systray.Quit()
		for {
			select {

			case <-m.selectedDevice.menuItem.ClickedCh:
				m.refreshHeadsets()
				m.changeState()

			case dName := <-m.devices.menuItem.ClickedCh:
				if !strings.HasPrefix(dName, "\u200B\u2060") {
					m.refreshHeadsets()
					m.selectDevice(dName)
				}

			case <-m.refresh.ClickedCh:
				m.refreshHeadsets()

			case <-m.exit.ClickedCh:
				return
			}
		}
	}()

}

func evalState(sinkName string) (exist bool, state headset.State) {
	pacmd := exec.Command("pacmd", "list-sinks")
	grep := exec.Command("grep", sinkName)

	pipe, _ := pacmd.StdoutPipe()
	defer pipe.Close()

	grep.Stdin = pipe

	pacmd.Start()

	res, _ := grep.Output()

	name := regexp.MustCompile("<(.*?)>").FindString(string(res))
	name = strings.Trim(name, "<")
	name = strings.Trim(name, ">")

	elements := strings.Split(name, ".")

	if len(elements) < sIndexCount {
		exist = false
	} else {
		state = headset.ToState(elements[sIndexState])
		exist = state != headset.SUnknown
	}

	return
}

func onExit() {
	icons.DeleteIcons()
	log.Println("Exit")
}
