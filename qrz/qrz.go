package qrz

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html/charset"
)

//
// https://www.qrz.com/XML/current_spec.html
//

// Client is our type
type Client struct {
	endpoint string
	username string
	password string
	agent    string

	httpclient *http.Client
	sessionKey string

	// mutex for sessionKey
	m sync.Mutex
}

type qrzSession struct {
	Text   string `xml:",chardata"`
	Key    string `xml:"Key"`
	Count  string `xml:"Count"`
	SubExp string `xml:"SubExp"`
	Error  string `xml:"Error"`
	GMTime string `xml:"GMTime"`
}

type sessionResponse struct {
	Session qrzSession `xml:"Session"`
}

type qrzCallsign struct {
	Text      string `xml:",chardata"`
	Call      string `xml:"call"`
	Aliases   string `xml:"aliases"`
	Dxcc      string `xml:"dxcc"`
	Fname     string `xml:"fname"`
	Name      string `xml:"name"`
	Addr1     string `xml:"addr1"`
	Addr2     string `xml:"addr2"`
	State     string `xml:"state"`
	Zip       string `xml:"zip"`
	Country   string `xml:"country"`
	Ccode     string `xml:"ccode"`
	Lat       string `xml:"lat"`
	Lon       string `xml:"lon"`
	Grid      string `xml:"grid"`
	County    string `xml:"county"`
	Fips      string `xml:"fips"`
	Land      string `xml:"land"`
	Efdate    string `xml:"efdate"`
	Expdate   string `xml:"expdate"`
	PCall     string `xml:"p_call"`
	Class     string `xml:"class"`
	Codes     string `xml:"codes"`
	Qslmgr    string `xml:"qslmgr"`
	Email     string `xml:"email"`
	URL       string `xml:"url"`
	UViews    string `xml:"u_views"`
	Bio       string `xml:"bio"`
	Image     string `xml:"image"`
	Serial    string `xml:"serial"`
	Moddate   string `xml:"moddate"`
	MSA       string `xml:"MSA"`
	AreaCode  string `xml:"AreaCode"`
	TimeZone  string `xml:"TimeZone"`
	GMTOffset string `xml:"GMTOffset"`
	DST       string `xml:"DST"`
	Eqsl      string `xml:"eqsl"`
	Mqsl      string `xml:"mqsl"`
	Cqzone    string `xml:"cqzone"`
	Ituzone   string `xml:"ituzone"`
	Geoloc    string `xml:"geoloc"`
	Attn      string `xml:"attn"`
	Nickname  string `xml:"nickname"`
	NameFmt   string `xml:"name_fmt"`
	Born      string `xml:"born"`
}

type CallsignLookupResponse struct {
	Callsign qrzCallsign `xml:"Callsign"`
	Session  qrzSession  `xml:"Session"`
}

func (clr CallsignLookupResponse) String() string {
	b, err := json.MarshalIndent(clr, "", "  ")
	if err != nil {
		log.Printf("%+v", err)
		return ""
	}
	return "CallsignLookupResponse" + string(b)
}

func NewClient(endpoint, username, password, agent string) (*Client, error) {
	client := &Client{
		endpoint: endpoint,
		username: username,
		password: password,
		agent:    agent,
		httpclient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	err := client.initializeSession()
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return client, nil
}

func (client *Client) initializeSession() error {
	// create session
	b, err := client.makeRequest(url.Values{
		"username": {client.username},
		"password": {client.password},
		"agent":    {client.agent},
	})
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var s sessionResponse
	decoder := xml.NewDecoder(bytes.NewReader(b))
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&s)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// check for error
	if len(s.Session.Error) > 0 {
		err = errors.New(s.Session.Error)
		log.Printf("%+v", err)
		return err
	}

	client.m.Lock()
	client.sessionKey = s.Session.Key
	client.m.Unlock()

	return nil
}

// makeRequest is a helper function to wrap making calls to the QRZ XML Interface
func (client *Client) makeRequest(parameters url.Values) ([]byte, error) {
	// create request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", client.endpoint, parameters.Encode()), nil)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	request.Header.Set("Accept", "application/xml")

	// make request, get response
	var response *http.Response
	response, err = client.httpclient.Do(request)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	defer response.Body.Close()

	// error?
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		err = fmt.Errorf("returned status code %d ", response.StatusCode)
		log.Printf("%+v", err)
		return nil, err
	}

	// get body for caller, if there is something
	var data []byte
	if response.ContentLength != 0 {
		data, err = io.ReadAll(response.Body)
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}
	}

	return data, nil
}

func (client *Client) CallsignLookup(callsign string) (*CallsignLookupResponse, error) {
	var clr CallsignLookupResponse

	request := func() error {
		// form request parameters
		parameters := url.Values{
			"callsign": {callsign},
		}

		// include session
		client.m.Lock()
		if len(client.sessionKey) > 0 {
			parameters.Add("s", client.sessionKey)
		} else {
			err := errors.New("no QRZ session")
			log.Printf("%+v", err)
			client.m.Unlock()
			return err
		}
		client.m.Unlock()

		b, err := client.makeRequest(parameters)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		decoder := xml.NewDecoder(bytes.NewReader(b))
		decoder.CharsetReader = charset.NewReaderLabel
		err = decoder.Decode(&clr)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		return nil
	}

	err := request()
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	// any response from the server that does not contain the Key element indicates
	// that no valid session exists and that a re-login is required to continue
	if len(clr.Session.Key) == 0 {
		log.Println("refreshing session")

		err = client.initializeSession()
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}

		// try again
		err := request()
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}
	}

	// check for session error
	if len(clr.Session.Error) > 0 {
		err = errors.New(clr.Session.Error)
		log.Printf("%+v", err)
		return nil, err
	}

	return &clr, nil
}
