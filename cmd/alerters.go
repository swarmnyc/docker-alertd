package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"gopkg.in/gomail.v2"
	"github.com/pkg/errors"
)

// Alerter is the interface which will handle alerting via different methods such as email
// and twitter/slack
type Alerter interface {
	Valid() error
	Alert(a *Alert) error
}

// Email implements the Alerter interface and sends emails
type Email struct {
	SMTP     string
	Username string
	Password string
	Port     string
	From     string
	To       []string
	Subject  string
}

// Alert sends an email alert
func (e Email) Alert(a *Alert) error {
	subject := e.Subject + ": "
	for i := range a.SubjectAddendums {
		// add addendums to the subject
		subject += fmt.Sprintf("%s ", a.SubjectAddendums[i])
		if i == 2 { // subjects cannot be too long, stop if it is at position 3
			subject += fmt.Sprintf("...")
		}
	}

	to := strings.Join(e.To, ",")

	m := gomail.NewMessage()
    
	// Set the main email part to use HTML.
	m.SetBody("text/html", a.DumpEmail())

	// Construct the message headers, including a Configuration Set and a Tag.
	m.SetHeaders(map[string][]string{
			"From": {e.From},
			"To": {to},
			"Subject": {subject},
	})

	// Send the email.
	port, _ := strconv.Atoi(e.Port)
	d := gomail.NewPlainDialer(e.SMTP, port, e.Username, e.Password)

	if err := d.DialAndSend(m); err != nil {
		return errors.Wrap(err, "error sending email")
	}

	log.Println("alert email sent")

	return nil
}

// Valid returns true if the email settings are complete
func (e Email) Valid() error {
	errString := []string{}

	if reflect.DeepEqual(Email{}, e) {
		return nil // assume that email alerts were omitted
	}

	if e.SMTP == "" {
		errString = append(errString, ErrEmailNoSMTP.Error())
	}

	if len(e.To) < 1 {
		errString = append(errString, ErrEmailNoTo.Error())
	}

	if e.From == "" {
		errString = append(errString, ErrEmailNoFrom.Error())
	}

	if e.Username == "" {
		errString = append(errString, ErrEmailNoUser.Error())
	}

	if e.Password == "" {
		errString = append(errString, ErrEmailNoPass.Error())
	}

	if e.Port == "" {
		errString = append(errString, ErrEmailNoPort.Error())
	}

	if e.Subject == "" {
		errString = append(errString, ErrEmailNoSubject.Error())
	}

	if len(errString) == 0 {
		return nil
	}

	delimErr := strings.Join(errString, ", ")
	err := errors.New(delimErr)

	return errors.Wrap(err, "email settings validation fail")
}

// Slack contains all the info needed to connect to a slack channel
type Slack struct {
	WebhookURL string
}

// Valid returns an error if slack settings are invalid
func (s Slack) Valid() error {
	errString := []string{}

	if reflect.DeepEqual(Slack{}, s) {
		return nil // assume that slack was omitted
	}

	if s.WebhookURL == "" {
		errString = append(errString, ErrSlackNoWebHookURL.Error())
	}

	if len(errString) == 0 {
		return nil
	}

	delimErr := strings.Join(errString, ", ")
	err := errors.New(delimErr)

	return errors.Wrap(err, "slack settings validation fail")
}

// Alert sends the alert to a slack channel
func (s Slack) Alert(a *Alert) error {
	alerts := a.Dump()

	json := fmt.Sprintf("{\"text\": \"%s\"}", alerts)
	body := bytes.NewReader([]byte(json))
	resp, err := http.Post(s.WebhookURL, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("sent alert to slack")
	return nil
}

// Pushover contains all info needed to push a notification to Pushover api
type Pushover struct {
	APIToken string
	UserKey  string
	APIURL   string
}

// Valid returns an error if pushover settings are invalid
func (p Pushover) Valid() error {
	errString := []string{}

	if reflect.DeepEqual(Pushover{}, p) {
		return nil // assume that pushover was omitted
	}

	if p.APIToken == "" {
		errString = append(errString, ErrPushoverAPIToken.Error())
	}

	if p.UserKey == "" {
		errString = append(errString, ErrPushoverUserKey.Error())
	}

	if p.APIURL == "" {
		errString = append(errString, ErrPushoverAPIURL.Error())
	}

	if len(errString) == 0 {
		return nil
	}

	delimErr := strings.Join(errString, ", ")
	err := errors.New(delimErr)

	return errors.Wrap(err, "pushover settings validation fail")
}

// Alert sends the alert to Pushover API
func (p Pushover) Alert(a *Alert) error {
	alerts := a.Dump()

	parsedBody := fmt.Sprintf("token=%s&user=%s&message=%s", p.APIToken, p.UserKey,
		url.QueryEscape(alerts))
	body := bytes.NewBufferString(parsedBody)

	resp, err := http.Post(p.APIURL, "application/x-www-form-urlencoded", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("sent alert to pushover")
	return nil
}
