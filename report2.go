package report2

// Creator interface of device creator
type Creator interface {
	СreateDevice(model string) Device // Parameterized Factory Method
	registerDevice(Device)            // Registration of the created device
}

// Device interface of device
type Device interface {
	SetSerial(serial string)
	SetSheet(sheet xlsx.Sheet)
	Send() error // Every device should be usable
}

// ConcreteCreator struct for concrete device creator
type ConcreteCreator struct {
	devices []*Device // Produced devices
}

// CreateDevice method to create concrete device
func (concreteCreator *ConcreteCreator) CreateDevice(model string, app string) Device {
	var device Device

	if app == "Arc" {
		switch model {
		case "TV7":
			//device = &TV7Arc{}
		default:
			log.Fatalln("Unknown device")
		}
	} else {
		switch model {
		case "TV7":
			//device = &TV7Vzl{}
		case "VKT7":
			//device = &VKT7Vzl{}
		case "TSRV030":
			device = &TSRV030Vzl{}
		case "TSRV034":
			//device = &TSRV034Vzl{}
		default:
			log.Fatalln("Unknown device")
		}
	}

	concreteCreator.RegisterDevice(device)

	return device
}

// RegisterDevice unnecessary function for registering devices in the creator
func (concreteCreator *ConcreteCreator) RegisterDevice(device Device) {
	concreteCreator.devices = append(concreteCreator.devices, &device)
}