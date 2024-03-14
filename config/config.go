package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/lxn/walk"
	"gopkg.in/yaml.v3"
)

var (
	configFile      string
	errNoConfig     = errors.New("no current configuration file")
	msgMissingField = "required configuration missing %s"

	// unwrapped config values
	UI    ui
	QRZ   qrz
	Email email
)

type mainwinrectangle struct {
	X             int `yaml:"topleftx"`
	Y             int `yaml:"toplefty"`
	Width, Height int
}

func (mwr *mainwinrectangle) FromBounds(bounds walk.Rectangle) {
	mwr.X = bounds.X
	mwr.Y = bounds.Y
	mwr.Width = bounds.Width
	mwr.Height = bounds.Height
}

func (mwr *mainwinrectangle) ToBounds() walk.Rectangle {
	return walk.Rectangle{
		X:      mwr.X,
		Y:      mwr.Y,
		Width:  mwr.Width,
		Height: mwr.Height,
	}
}

type ui struct {
	MainWinRectangle mainwinrectangle
}

type qrz struct {
	Endpoint string // the QRZ Versioned URL
	Username string // a valid QRZ user name
	Password string // the correct password for the username
	Agent    string // a string that contains the product name and version of the client program
}

// Validate tests the required qrz fields
// doesn't log errors because you don't have to use qrz
func (q *qrz) Validate() error {
	if q.Endpoint == "" {
		err := fmt.Errorf(msgMissingField, "QRZ Endpoint")
		return err
	}
	if q.Username == "" {
		err := fmt.Errorf(msgMissingField, "QRZ Username")
		return err
	}
	if q.Password == "" {
		err := fmt.Errorf(msgMissingField, "QRZ Password")
		return err
	}
	if q.Agent == "" {
		err := fmt.Errorf(msgMissingField, "QRZ Agent")
		return err
	}

	return nil
}

type email struct {
	ServerPort      string
	From            string
	SubjectTemplate string // QSL Bureau cards for {{ callsign }}
	BodyTemplate    string // QSL Bureau cards for {{ callsign }}
}

// Validate tests the required qrz fields
// doesn't log errors because you don't have to use qrz
func (e *email) Validate() error {
	if e.ServerPort == "" {
		err := fmt.Errorf(msgMissingField, "email ServerPort")
		return err
	}
	if e.From == "" {
		err := fmt.Errorf(msgMissingField, "email From")
		return err
	}
	if e.SubjectTemplate == "" {
		err := fmt.Errorf(msgMissingField, "email SubjectTemplate")
		return err
	}
	if e.BodyTemplate == "" {
		err := fmt.Errorf(msgMissingField, "email BodyTemplate")
		return err
	}

	return nil
}

// Configuration is the application configuration that is serialized/deserialized to file
type Configuration struct {
	UI    ui
	QRZ   qrz
	Email email
}

// Validate tests the required Configuration fields
func (c *Configuration) Validate() error {
	err := c.QRZ.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = c.Email.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// Read loads application configuration from file fname
func Read(fname string) error {
	// save for write later
	configFile = fname

	// #nosec G304
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var c Configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// make sure valid before unwrapping
	err = c.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	UI = c.UI
	QRZ = c.QRZ
	Email = c.Email

	return nil
}

// Write writes application configuration to the same file it was read from
func Write() error {
	if configFile == "" {
		return errNoConfig
	}

	// wrap
	c := Configuration{
		UI:    UI,
		QRZ:   QRZ,
		Email: Email,
	}

	// make sure valid before proceeding
	err := c.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// create YAML to write from Options
	b, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// write out to file
	err = os.WriteFile(configFile, b, 0600)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}
