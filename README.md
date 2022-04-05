# golib
Collections of various modules that i use in my go projects, 

`Note:` This lib is not ready for protection, tests aren't not written, might be unstable.

## List of Libraries:
- Gcalman[pkg/cal.go]: A Wrapper over Google calendar SDK, Google calendar API is too much verbose for beginners, intention was to simplify it by writing a layer above it to meet simple use cases.
- Install: `go get -u github.com/arvryna/golib/pkg/gcalman`

```
Usage:

func main() {
	gc := gcalman.Init("auth-token.json", "access-token.json")
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

	gc.CreateEvent("primary", gc.BuildCalEventObject(&event))

}
```