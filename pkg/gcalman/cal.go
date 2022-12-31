// Gcalman - Google calendar manager
// Author: Arv
// Wrapper over Google calendar SDK for calender management

package gcalman

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GcalMan interface {
	CreateEvent(calId string, event *calendar.Event) (*CreateEventResp, error)
	BuildCalEventObject(calEvent *GcalEvent) *calendar.Event
	GenerateAccessTokenFromAuthToken(authTokenPath string, preferredAccessTokenPath string) error
	PrintEvents(calId string)
}

type gcalman struct {
	OauthToken  string
	AccessToken string
	CalServ     *calendar.Service
}

type GcalEvent struct {
	Title          string
	Description    string
	AttendeeEmails []string
	Start          string
	End            string
	Location       string
	// This sends a calendar invite email to user
	SendInvite bool
	// By setting this flag, the event will appear in the calender of user as "AcceptedEvent"
	AcceptedEvent bool
}

type CreateEventResp struct {
	Id          string
	HtmlLink    string
	HangoutLink string
}

/*
Fetching Access Token
*********************

  - Oauth token must be ready in advance, it needs to be fetched from Google developer console,
    https://developers.google.com/calendar/api/quickstart/go

  - Then you need to got "Enable API Wizard" to get your credentials, which you need to pass in authTokenPath

  - This function Generates Auth URL in STDOUT with instructions

  - Once AccessToken is provided as STDIN, it will be stored in the provided preferredAccessTokenPath

  - Once the tokens are saved to disk, we can Initialize this library using Init function
*/
func (g *gcalman) GenerateAccessTokenFromAuthToken(authTokenPath string, preferredAccessTokenPath string) error {
	file, err := ioutil.ReadFile(authTokenPath)
	if err != nil {
		return err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(file, calendar.CalendarScope)
	if err != nil {
		return err
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+"authorization code: \n%v\n", authURL)

	var authCode string
	fmt.Printf("Paste the Token here: ")

	if _, err := fmt.Scan(&authCode); err != nil {
		return err
	}

	token, err := config.Exchange(context.TODO(), authCode)

	if err != nil {
		return err
	}

	if err = saveToken(preferredAccessTokenPath, token); err != nil {
		return err
	}

	return nil
}

/*
Initialize GcalMan Object

- Details about fetching the auth/access token is provided in the documentation
- Params: File path of oauthtoken and accesstoken
*/
func Init(oauthToken string, accessToken string) (GcalMan, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile(oauthToken)
	if err != nil {
		return nil, err
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, err
	}

	token, err := tokenFromFile(accessToken)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx, token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &gcalman{OauthToken: oauthToken, AccessToken: accessToken, CalServ: srv}, nil
}

// Fetches event from Google calendar by default from primary calendar ID
func (g *gcalman) PrintEvents(calId string) {

	if calId == "" {
		calId = "primary"
	}

	t := time.Now().Format(time.RFC3339)
	events, err := g.CalServ.Events.List(calId).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}
}

// calendar.Event struct is very complex with lot of features, it was abstracted
// away by our GcalEvent struct, which is simple, After building this calendar.Event,
// it is possible to call the CreateEvent function.
// Create Event Object in GoogelCalendar format.
// Example Time format "2022-03-31T10:45:26.371Z"
func (g *gcalman) BuildCalEventObject(calEvent *GcalEvent) *calendar.Event {
	startTime := &calendar.EventDateTime{
		DateTime: calEvent.Start,
	}

	endTime := &calendar.EventDateTime{
		DateTime: calEvent.End,
	}

	var attendees []*calendar.EventAttendee

	var eventStatus = ""
	if calEvent.AcceptedEvent {
		eventStatus = "accepted"
	}

	for _, val := range calEvent.AttendeeEmails {
		attendee := &calendar.EventAttendee{
			Email:          val,
			ResponseStatus: eventStatus,
		}
		attendees = append(attendees, attendee)
	}
	event := &calendar.Event{
		Attendees:   attendees,
		Summary:     calEvent.Title,
		Description: calEvent.Description,
		Start:       startTime,
		End:         endTime,
		Location:    calEvent.Location,
		Reminders: &calendar.EventReminders{
			UseDefault: calEvent.SendInvite,
		},
	}
	return event
}

// Create event in Google Calender, required to pass
func (g *gcalman) CreateEvent(calId string, event *calendar.Event) (*CreateEventResp, error) {

	eventDetails := &CreateEventResp{}

	if calId == "" {
		calId = "primary"
	}

	eventCall := g.CalServ.Events.Insert(calId, event)

	eventCall.SendUpdates("all")

	res, err := eventCall.ConferenceDataVersion(1).Do()
	if err != nil {
		return nil, err
	}

	eventDetails.Id = res.Id
	eventDetails.HangoutLink = res.HangoutLink
	eventDetails.HtmlLink = res.HtmlLink

	return eventDetails, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = json.NewEncoder(f).Encode(token); err != nil {
		return err
	}

	return nil
}
