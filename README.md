# golib

Collection of various modules that i use in my go projects, 

## 1.  Gcalman[pkg/gcalman/cal.go]
- A Wrapper over Google calendar SDK, Google calendar API is too much verbose for beginners, intention was to simplify it by writing a layer above it to meet simple use cases.
- Install: `go get -u github.com/arvryna/golib/pkg/gcalman`

```
Usage:

func main() {
	gc := gcalman.Init("auth-token.json", "access-token.json")

    // Fetching list of events in the primary calendar
	gc.GetEvents("primary")

	event := gcalman.GcalEvent{
		Title:          "This is a test event",
		Description:    "This is a new desc",
		AttendeeEmails: []string{"email1@example.com", "email2@example.com"},
		Start:          "2022-04-05T18:45:26.371+03:00",
		End:            "2022-04-05T19:45:26.371+03:00",
		SendInvite:     true,
		Location:       "Warsawa",
		AcceptedEvent:  true,
	}

    // Create an Event in primary calendar
	gc.CreateEvent("primary", gc.BuildCalEventObject(&event))

}
```

## 2. FileUtils[In-progress]

- List of utils
- Install: `go get -u github.com/arvryna/golib/pkg/utils`