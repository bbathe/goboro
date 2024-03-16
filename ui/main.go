package ui

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bbathe/goboro/config"
	"github.com/bbathe/goboro/email"
	"github.com/bbathe/goboro/qrz"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var (
	appName = "Go Boro"
	appIcon *walk.Icon

	mainWin  *walk.MainWindow
	runDll32 string

	qrzClient *qrz.Client
)

func init() {
	var err error

	// load app icon
	appIcon, err = walk.Resources.Icon("2")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// full path to rundll32 for launching web browser
	runDll32 = filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
}

// launchQRZPage opens the users default web browser to the qso partners QRZ.com page
func launchQRZPage(call string) error {
	u := "https://www.qrz.com"
	if call != "" {
		u += "/db/" + strings.Replace(call, "%", "", -1)
	}

	err := exec.Command(runDll32, "url.dll,FileProtocolHandler", u).Start()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// goboroWindow creates the main window and begins processing of user input
func GoBoroWindow() error {
	var err error

	var leCall *walk.LineEdit
	var pbQRZ *walk.PushButton
	var pbLookup *walk.PushButton
	var leEmailTo *walk.LineEdit
	var leSubject *walk.LineEdit
	var teBody *walk.TextEdit
	var pbSend *walk.PushButton

	// establish qrz.com session
	qrzClient, err = qrz.NewClient(config.QRZ.Endpoint, config.QRZ.Username, config.QRZ.Password, config.QRZ.Agent)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	tmplSubject, err := template.New("subject").Parse(config.Email.SubjectTemplate)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	tmplBody, err := template.New("body").Parse(config.Email.BodyTemplate)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// goboro main window
	err = declarative.MainWindow{
		AssignTo: &mainWin,
		Title:    appName,
		Icon:     appIcon,
		Visible:  false,
		Font: declarative.Font{
			Family:    "MS Shell Dlg 2",
			PointSize: 10,
		},
		Layout: declarative.VBox{},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.HBox{MarginsZero: true},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Callsign",
							},
							declarative.Composite{
								Layout: declarative.HBox{MarginsZero: true},
								Children: []declarative.Widget{
									declarative.LineEdit{
										Text:     declarative.Bind("Call"),
										CaseMode: declarative.CaseModeUpper,
										AssignTo: &leCall,
										OnKeyPress: func(key walk.Key) {
											if key == walk.KeyReturn {
												pbLookup.SendMessage(win.BM_CLICK, 0, 0)
											}
										},
									},
									declarative.PushButton{
										AssignTo:    &pbLookup,
										Text:        "\U000025B6",
										ToolTipText: "lookup QRZ information",
										MaxSize: declarative.Size{
											Width: 30,
										},
										MinSize: declarative.Size{
											Width: 30,
										},
										Font: declarative.Font{
											Family:    "MS Shell Dlg 2",
											PointSize: 9,
										},
										OnClicked: func() {
											call := strings.TrimSpace(leCall.Text())
											if len(call) > 0 {
												r, err := qrzClient.CallsignLookup(call)
												if err != nil {
													MsgError(mainWin, err)
													log.Printf("%+v", err)
													return
												}

												// populate email components
												if call == r.Callsign.Call {
													if len(r.Callsign.Email) > 0 {
														leEmailTo.SetText(r.Callsign.Email)

														var s bytes.Buffer
														tmplSubject.Execute(&s, map[string]string{"callsign": r.Callsign.Call})
														leSubject.SetText(s.String())

														var b bytes.Buffer
														tmplBody.Execute(&b, map[string]string{"callsign": r.Callsign.Call})
														teBody.SetText(string(bytes.Replace(b.Bytes(), []byte{'\n'}, []byte{'\r', '\n'}, -1)))
													} else {
														MsgError(mainWin, errors.New("no email address"))
													}
												} else {
													MsgError(mainWin, errors.New("Callsign changed to "+r.Callsign.Email))
												}
											}
										},
									},
									declarative.PushButton{
										AssignTo:    &pbQRZ,
										Text:        "\U0001F310",
										ToolTipText: "visit QRZ.com page",
										MaxSize: declarative.Size{
											Width: 30,
										},
										MinSize: declarative.Size{
											Width: 30,
										},
										Font: declarative.Font{
											Family:    "MS Shell Dlg 2",
											PointSize: 9,
										},
										OnClicked: func() {
											err := launchQRZPage(leCall.Text())
											if err != nil {
												MsgError(mainWin, err)
												log.Printf("%+v", err)
												return
											}
										},
									},
								},
							},
							declarative.Label{
								Text: "To",
							},
							declarative.LineEdit{
								Text:     declarative.Bind("EmailTo"),
								AssignTo: &leEmailTo,
							},
							declarative.Label{
								Text: "Subject",
							},
							declarative.LineEdit{
								Text:     declarative.Bind("Subject"),
								AssignTo: &leSubject,
							},
							declarative.Label{
								Text: "Body",
							},
							declarative.TextEdit{
								Text:     declarative.Bind("Body"),
								AssignTo: &teBody,
							},
							declarative.PushButton{
								AssignTo:    &pbSend,
								Text:        "Send",
								ToolTipText: "send email",
								Font: declarative.Font{
									Family:    "MS Shell Dlg 2",
									PointSize: 9,
								},
								OnClicked: func() {
									//
									// Directory (tenant) ID: 51d5aff2-f2a7-435d-b5aa-a7f3693ebb61
									// Application (client) ID: d391e9e8-8595-41c9-bc74-75e02e887e5c
									// Secret Value: pl68Q~vsIsNnVYCucrX6~HaRgTNj3J.huY1Kyb~Z
									// Secret ID: a7df6321-c556-4340-b57a-470e0cab9346
									//

									c, err := email.NewClient(config.Office365AppRegistration.TenantID, config.Office365AppRegistration.ClientID, config.Office365AppRegistration.Secret)
									if err != nil {
										MsgError(mainWin, err)
										log.Printf("%+v", err)
										return
									}

									err = c.Send(config.Email.UserID, leSubject.Text(), teBody.Text(), leEmailTo.Text())
									if err != nil {
										MsgError(mainWin, err)
										log.Printf("%+v", err)
										return
									}
								},
							},
						},
					},
				},
			},
		},
	}.Create()
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return err
	}

	// set window position based on config
	err = mainWin.SetBounds(config.UI.MainWinRectangle.ToBounds())
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return err
	}

	// save windows position in config during window close
	mainWin.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		config.UI.MainWinRectangle.FromBounds(mainWin.Bounds())
	})

	// make visible
	mainWin.SetVisible(true)

	// start message loop
	mainWin.Run()

	return nil
}
