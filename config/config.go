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
	UI                       ui
	QRZ                      qrz
	Office365AppRegistration office365AppRegistration
	Email                    email
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

type office365AppRegistration struct {
	TenantID string
	ClientID string
	Secret   string
}

// Validate tests the required office365AppRegistration fields
// doesn't log errors because you don't have to use qrz
func (o *office365AppRegistration) Validate() error {
	if o.TenantID == "" {
		err := fmt.Errorf(msgMissingField, "Office365AppRegistration TenantID")
		return err
	}
	if o.ClientID == "" {
		err := fmt.Errorf(msgMissingField, "Office365AppRegistration ClientID")
		return err
	}
	if o.Secret == "" {
		err := fmt.Errorf(msgMissingField, "Office365AppRegistration Secret")
		return err
	}

	return nil
}

type email struct {
	UserID          string // from user, UPN or ObjectID
	SubjectTemplate string // QSL Bureau cards for {{ callsign }}
	BodyTemplate    string // QSL Bureau cards for {{ callsign }}
}

// Validate tests the required email fields
// doesn't log errors because you don't have to use qrz
func (e *email) Validate() error {
	if e.UserID == "" {
		err := fmt.Errorf(msgMissingField, "Email UserID")
		return err
	}
	if e.SubjectTemplate == "" {
		err := fmt.Errorf(msgMissingField, "Email SubjectTemplate")
		return err
	}
	if e.BodyTemplate == "" {
		err := fmt.Errorf(msgMissingField, "Email BodyTemplate")
		return err
	}

	return nil
}

// Configuration is the application configuration that is serialized/deserialized to file
type Configuration struct {
	UI                       ui
	QRZ                      qrz
	Office365AppRegistration office365AppRegistration
	Email                    email
}

// Validate tests the required Configuration fields
func (c *Configuration) Validate() error {
	err := c.QRZ.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = c.Office365AppRegistration.Validate()
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
	Office365AppRegistration = c.Office365AppRegistration
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
		UI:                       UI,
		QRZ:                      QRZ,
		Office365AppRegistration: Office365AppRegistration,
		Email:                    Email,
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
