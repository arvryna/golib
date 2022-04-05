// Gcalman - Google calendar manager
// Author: Arv
// Wrapper over Google calendar SDK for calender management

package gcalman

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GcalMan interface {
	GetEvents(calId string)
	BuildCalEventObject(calEvent *GcalEvent) *calendar.Event
	CreateEvent(calId string, event *calendar.Event) error
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
	// This sends adds an entry in the users's calendar and also sends email to the user's Inbox
	SendInvite bool
	// By setting this flag, the event will appear in the calender of user as "AcceptedEvent"
	AcceptedEvent bool
}

// Oauth token must be ready in advance, it can be fetched from
// Google developer console, more instructions in the docs
// This function Generates Auth URL in STDOUT with instructions
// Once AccessToken is provided as STDIN, it will be stored in the preferredAccessTokenPath
// Once the tokens are saved to disk, we can Initialize this library using Init function
func GenerateAccessTokenFromAuthToken(authTokenPath string, preferredAccessTokenPath string) {
	b, err := ioutil.ReadFile(authTokenPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	saveToken(preferredAccessTokenPath, tok)
}

// Params: File path of oauthtoken and accesstoken
// Details about fetching the auth/access token is provided in the documentation
func Init(oauthToken string, accessToken string) GcalMan {
	ctx := context.Background()
	b, err := ioutil.ReadFile(oauthToken)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	token, err := tokenFromFile(accessToken)
	if err != nil {
		log.Fatalf("Unable to read Token from file: %v", err)
	}

	client := config.Client(ctx, token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	} else {
		log.Println("Gcalman Sucessfully Initialized !")
	}

	return &gcalman{OauthToken: oauthToken, AccessToken: accessToken, CalServ: srv}
}

// Fetches event from Google calendar by default from primary calendar ID
func (g *gcalman) GetEvents(calId string) {

	if calId == "" {
		calId = "primary"
	}

	t := time.Now().Format(time.RFC3339)
	events, err := g.CalServ.Events.List(calId).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
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
func (g *gcalman) CreateEvent(calId string, event *calendar.Event) error {

	if calId == "" {
		calId = "primary"
	}

	eventCall := g.CalServ.Events.Insert(calId, event)
	eventCall.SendUpdates("all")
	res, err := eventCall.ConferenceDataVersion(1).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	} else {
		fmt.Println("Create event success")
	}

	log.Println(res, res.HangoutLink, res.HtmlLink, res.Id)
	return err
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
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
