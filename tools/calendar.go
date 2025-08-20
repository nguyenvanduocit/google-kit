package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/google-kit/services"
	"github.com/nguyenvanduocit/google-kit/util"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v3"
)

func RegisterCalendarTools(s *server.MCPServer) {
	// Unified event management tool
	eventTool := mcp.NewTool("calendar_event",
		mcp.WithDescription("Manage Google Calendar events - create, update, list, or respond to events"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: create, update, list, respond")),
		mcp.WithString("event_id", mcp.Description("ID of the event (required for update/respond actions)")),
		mcp.WithString("summary", mcp.Description("Title of the event (required for create, optional for update)")),
		mcp.WithString("description", mcp.Description("Description of the event")),
		mcp.WithString("start_time", mcp.Description("Start time in RFC3339 format (required for create, optional for update/list)")),
		mcp.WithString("end_time", mcp.Description("End time in RFC3339 format (required for create, optional for update/list)")),
		mcp.WithString("attendees", mcp.Description("Comma-separated list of attendee email addresses")),
		mcp.WithString("time_min", mcp.Description("Start time for search in RFC3339 format (list action, default: now)")),
		mcp.WithString("time_max", mcp.Description("End time for search in RFC3339 format (list action, default: 1 week from now)")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of events to return (list action, default: 10)")),
		mcp.WithString("response", mcp.Description("Your response: accepted, declined, or tentative (respond action)")),
	)
	s.AddTool(eventTool, util.ErrorGuard(calendarEventHandler))


	// Find time slot tool
	findTimeSlotTool := mcp.NewTool("calendar_find_time_slot",
		mcp.WithDescription("Find available time slots based on room or guest availability"),
		mcp.WithString("guests", mcp.Description("Comma-separated list of guest email addresses to check availability")),
		mcp.WithString("room", mcp.Description("Room to filter events by")),
		mcp.WithString("start_date", mcp.Required(), mcp.Description("Start date for searching slots in RFC3339 format")),
		mcp.WithString("end_date", mcp.Required(), mcp.Description("End date for searching slots in RFC3339 format")),
		mcp.WithNumber("duration_minutes", mcp.Required(), mcp.Description("Duration of the meeting in minutes")),
		mcp.WithString("working_hours_start", mcp.Description("Start of working hours (e.g., '09:00', default: 09:00)")),
		mcp.WithString("working_hours_end", mcp.Description("End of working hours (e.g., '17:00', default: 17:00)")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of time slots to return (default: 5)")),
	)
	s.AddTool(findTimeSlotTool, util.ErrorGuard(calendarFindTimeSlotHandler))

	// Get busy times tool
	getBusyTimesTool := mcp.NewTool("calendar_get_busy_times",
		mcp.WithDescription("Get busy time periods for one or multiple users"),
		mcp.WithString("users", mcp.Description("Comma-separated list of user email addresses (leave empty for primary calendar only)")),
		mcp.WithString("start_date", mcp.Required(), mcp.Description("Start date for the search in RFC3339 format")),
		mcp.WithString("end_date", mcp.Required(), mcp.Description("End date for the search in RFC3339 format")),
	)
	s.AddTool(getBusyTimesTool, util.ErrorGuard(calendarGetBusyTimesHandler))
}

var calendarService = sync.OnceValue(func() *calendar.Service {
	ctx := context.Background()

	tokenFile := os.Getenv("GOOGLE_TOKEN_FILE")
	if tokenFile == "" {
		panic("GOOGLE_TOKEN_FILE environment variable must be set")
	}

	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		panic("GOOGLE_CREDENTIALS_FILE environment variable must be set")
	}

	client := services.GoogleHttpClient(tokenFile, credentialsFile)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		panic(fmt.Sprintf("failed to create Calendar service: %v", err))
	}

	return srv
})

func calendarEventHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	action, _ := arguments["action"].(string)
	
	switch action {
	case "create":
		return calendarCreateEventHandler(arguments)
	case "update":
		return calendarUpdateEventHandler(arguments)
	case "list":
		return calendarListEventsHandler(arguments)
	case "respond":
		return calendarRespondToEventHandler(arguments)
	default:
		return mcp.NewToolResultError("Invalid action. Must be one of: create, update, list, respond"), nil
	}
}

func calendarCreateEventHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	summary, _ := arguments["summary"].(string)
	description, _ := arguments["description"].(string)
	startTimeStr, _ := arguments["start_time"].(string)
	endTimeStr, _ := arguments["end_time"].(string)
	attendeesStr, _ := arguments["attendees"].(string)

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid start_time format"), nil
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid end_time format"), nil
	}

	var attendees []*calendar.EventAttendee
	if attendeesStr != "" {
		for _, email := range strings.Split(attendeesStr, ",") {
			attendees = append(attendees, &calendar.EventAttendee{Email: email})
		}
	}

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
		Attendees: attendees,
	}

	createdEvent, err := calendarService().Events.Insert("primary", event).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create event: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created event with ID: %s", createdEvent.Id)), nil
}

func calendarListEventsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	timeMinStr, ok := arguments["time_min"].(string)
	if !ok || timeMinStr == "" {
		timeMinStr = time.Now().Format(time.RFC3339)
	}

	timeMaxStr, ok := arguments["time_max"].(string)
	if !ok || timeMaxStr == "" {
		timeMaxStr = time.Now().AddDate(0, 0, 7).Format(time.RFC3339) // 1 week from now
	}

	maxResults, ok := arguments["max_results"].(float64)
	if !ok {
		maxResults = 10
	}

	events, err := calendarService().Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMinStr).
		TimeMax(timeMaxStr).
		MaxResults(int64(maxResults)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list events: %v", err)), nil
	}

	eventsList := make([]map[string]interface{}, 0)

	for _, item := range events.Items {
		start, _ := time.Parse(time.RFC3339, item.Start.DateTime)
		end, _ := time.Parse(time.RFC3339, item.End.DateTime)

		eventInfo := map[string]interface{}{
			"id":      item.Id,
			"summary": item.Summary,
			"start":   start.Format("2006-01-02 15:04"),
			"end":     end.Format("2006-01-02 15:04"),
		}

		if item.Description != "" {
			eventInfo["description"] = item.Description
		}

		eventsList = append(eventsList, eventInfo)
	}

	result := map[string]interface{}{
		"count":  len(events.Items),
		"events": eventsList,
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal events: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}

func calendarUpdateEventHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	eventID, _ := arguments["event_id"].(string)
	summary, _ := arguments["summary"].(string)
	description, _ := arguments["description"].(string)
	startTimeStr, _ := arguments["start_time"].(string)
	endTimeStr, _ := arguments["end_time"].(string)
	attendeesStr, _ := arguments["attendees"].(string)

	event, err := calendarService().Events.Get("primary", eventID).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get event: %v", err)), nil
	}

	if summary != "" {
		event.Summary = summary
	}
	if description != "" {
		event.Description = description
	}
	if startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return mcp.NewToolResultError("Invalid start_time format"), nil
		}
		event.Start.DateTime = startTime.Format(time.RFC3339)
	}
	if endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return mcp.NewToolResultError("Invalid end_time format"), nil
		}
		event.End.DateTime = endTime.Format(time.RFC3339)
	}
	if attendeesStr != "" {
		var attendees []*calendar.EventAttendee
		for _, email := range strings.Split(attendeesStr, ",") {
			attendees = append(attendees, &calendar.EventAttendee{Email: email})
		}
		event.Attendees = attendees
	}

	updatedEvent, err := calendarService().Events.Update("primary", eventID, event).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update event: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully updated event with ID: %s", updatedEvent.Id)), nil
}

func calendarRespondToEventHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	eventID, _ := arguments["event_id"].(string)
	response, _ := arguments["response"].(string)

	event, err := calendarService().Events.Get("primary", eventID).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get event: %v", err)), nil
	}

	for _, attendee := range event.Attendees {
		if attendee.Self {
			attendee.ResponseStatus = response
			break
		}
	}

	_, err = calendarService().Events.Update("primary", eventID, event).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update event response: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully responded '%s' to event with ID: %s", response, eventID)), nil
}

func calendarFindTimeSlotHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	guestsStr, _ := arguments["guests"].(string)
	room, _ := arguments["room"].(string)
	startDateStr, _ := arguments["start_date"].(string)
	endDateStr, _ := arguments["end_date"].(string)
	durationMinutes, _ := arguments["duration_minutes"].(float64)
	workingHoursStart, _ := arguments["working_hours_start"].(string)
	workingHoursEnd, _ := arguments["working_hours_end"].(string)
	maxResults, _ := arguments["max_results"].(float64)

	if workingHoursStart == "" {
		workingHoursStart = "09:00"
	}
	if workingHoursEnd == "" {
		workingHoursEnd = "17:00"
	}
	if maxResults <= 0 {
		maxResults = 5
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid start_date format"), nil
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid end_date format"), nil
	}

	// Get all calendars to check (primary + guests)
	calendarsToCheck := []string{"primary"}
	if guestsStr != "" {
		for _, guest := range strings.Split(guestsStr, ",") {
			calendarsToCheck = append(calendarsToCheck, strings.TrimSpace(guest))
		}
	}

	// Collect all busy times with details
	allBusyTimes := make([]timeSlot, 0)
	busyDetails := make([]busyTime, 0)
	
	for _, calendarId := range calendarsToCheck {
		// Always use event listing to get details
		events, err := calendarService().Events.List(calendarId).
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(startDate.Format(time.RFC3339)).
			TimeMax(endDate.Format(time.RFC3339)).
			OrderBy("startTime").
			Do()
		
		if err != nil {
			continue // Skip this calendar if we can't access it
		}

		for _, event := range events.Items {
			// Filter by room if specified
			if room != "" && !strings.Contains(strings.ToLower(event.Location), strings.ToLower(room)) {
				continue
			}

			if event.Start.DateTime != "" && event.End.DateTime != "" {
				start, _ := time.Parse(time.RFC3339, event.Start.DateTime)
				end, _ := time.Parse(time.RFC3339, event.End.DateTime)
				
				allBusyTimes = append(allBusyTimes, timeSlot{Start: start, End: end})
				
				// Collect event details
				organizer := ""
				if event.Organizer != nil {
					organizer = event.Organizer.Email
				}
				
				busyDetails = append(busyDetails, busyTime{
					Start:      start,
					End:        end,
					Summary:    event.Summary,
					Organizer:  organizer,
					CalendarId: calendarId,
				})
			}
		}
	}

	// Merge overlapping busy times
	mergedBusyTimes := mergeTimeSlots(allBusyTimes)

	// Find available slots
	availableSlots := findAvailableSlots(
		startDate,
		endDate,
		mergedBusyTimes,
		time.Duration(durationMinutes)*time.Minute,
		workingHoursStart,
		workingHoursEnd,
		int(maxResults),
	)

	// Format results
	result := map[string]interface{}{
		"available_slots": make([]map[string]string, 0),
		"duration_minutes": durationMinutes,
		"busy_times": make([]map[string]string, 0),
	}

	if guestsStr != "" {
		result["guests_checked"] = guestsStr
	}
	if room != "" {
		result["room_filter"] = room
	}

	// Add available slots
	for _, slot := range availableSlots {
		slotInfo := map[string]string{
			"start": slot.Start.Format("2006-01-02 15:04"),
			"end":   slot.End.Format("2006-01-02 15:04"),
			"day":   slot.Start.Format("Monday"),
		}
		result["available_slots"] = append(result["available_slots"].([]map[string]string), slotInfo)
	}

	// Add busy time details
	for _, busy := range busyDetails {
		busyInfo := map[string]string{
			"start":     busy.Start.Format("2006-01-02 15:04"),
			"end":       busy.End.Format("2006-01-02 15:04"),
			"summary":   busy.Summary,
			"organizer": busy.Organizer,
		}
		
		// Add calendar info to identify whose calendar it is
		if busy.CalendarId == "primary" {
			busyInfo["calendar"] = "Your calendar"
		} else {
			busyInfo["calendar"] = busy.CalendarId
		}
		
		result["busy_times"] = append(result["busy_times"].([]map[string]string), busyInfo)
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}

type timeSlot struct {
	Start time.Time
	End   time.Time
}

type busyTime struct {
	Start       time.Time
	End         time.Time
	Summary     string
	Organizer   string
	CalendarId  string
}

func mergeTimeSlots(slots []timeSlot) []timeSlot {
	if len(slots) == 0 {
		return slots
	}

	// Sort slots by start time
	for i := 0; i < len(slots); i++ {
		for j := i + 1; j < len(slots); j++ {
			if slots[i].Start.After(slots[j].Start) {
				slots[i], slots[j] = slots[j], slots[i]
			}
		}
	}

	merged := []timeSlot{slots[0]}
	for i := 1; i < len(slots); i++ {
		last := &merged[len(merged)-1]
		if slots[i].Start.Before(last.End) || slots[i].Start.Equal(last.End) {
			// Overlapping or adjacent, merge them
			if slots[i].End.After(last.End) {
				last.End = slots[i].End
			}
		} else {
			// No overlap, add as new slot
			merged = append(merged, slots[i])
		}
	}

	return merged
}

func findAvailableSlots(startDate, endDate time.Time, busySlots []timeSlot, duration time.Duration, workStart, workEnd string, maxResults int) []timeSlot {
	availableSlots := make([]timeSlot, 0)
	
	// Parse working hours
	workStartHour, workStartMin := parseTimeString(workStart)
	workEndHour, workEndMin := parseTimeString(workEnd)

	currentDate := startDate
	for currentDate.Before(endDate) && len(availableSlots) < maxResults {
		// Set working hours for current day
		dayStart := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), workStartHour, workStartMin, 0, 0, currentDate.Location())
		dayEnd := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), workEndHour, workEndMin, 0, 0, currentDate.Location())

		// Skip weekends
		if currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Ensure we don't go before the start date
		if dayStart.Before(startDate) {
			dayStart = startDate
		}
		// Ensure we don't go after the end date
		if dayEnd.After(endDate) {
			dayEnd = endDate
		}

		// Find free slots in this day
		currentTime := dayStart
		for currentTime.Add(duration).Before(dayEnd) || currentTime.Add(duration).Equal(dayEnd) {
			slotEnd := currentTime.Add(duration)
			
			// Check if this slot conflicts with any busy time
			isAvailable := true
			for _, busySlot := range busySlots {
				if (currentTime.Before(busySlot.End) && slotEnd.After(busySlot.Start)) {
					// Conflict found
					isAvailable = false
					// Move current time to the end of the busy slot
					if busySlot.End.After(currentTime) {
						currentTime = busySlot.End
					}
					break
				}
			}

			if isAvailable {
				availableSlots = append(availableSlots, timeSlot{Start: currentTime, End: slotEnd})
				if len(availableSlots) >= maxResults {
					break
				}
				// Move to next potential slot (30 minute increments)
				currentTime = currentTime.Add(30 * time.Minute)
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return availableSlots
}

func parseTimeString(timeStr string) (hour, minute int) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 9, 0 // Default to 9:00
	}
	
	if _, err := fmt.Sscanf(parts[0], "%d", &hour); err != nil {
		hour = 9 // Default hour
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &minute); err != nil {
		minute = 0 // Default minute
	}
	return hour, minute
}

func calendarGetBusyTimesHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	usersStr, _ := arguments["users"].(string)
	startDateStr, _ := arguments["start_date"].(string)
	endDateStr, _ := arguments["end_date"].(string)

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid start_date format"), nil
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid end_date format"), nil
	}

	// Determine calendars to check
	calendarsToCheck := []string{"primary"}
	if usersStr != "" {
		calendarsToCheck = []string{}
		for _, user := range strings.Split(usersStr, ",") {
			calendarsToCheck = append(calendarsToCheck, strings.TrimSpace(user))
		}
	}

	// Collect busy times from all calendars
	busyDetails := make([]busyTime, 0)
	
	for _, calendarId := range calendarsToCheck {
		events, err := calendarService().Events.List(calendarId).
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(startDate.Format(time.RFC3339)).
			TimeMax(endDate.Format(time.RFC3339)).
			OrderBy("startTime").
			Do()
		
		if err != nil {
			// Skip calendars we can't access but include error info
			busyDetails = append(busyDetails, busyTime{
				Summary:    fmt.Sprintf("Error accessing calendar: %s", err.Error()),
				CalendarId: calendarId,
			})
			continue
		}

		for _, event := range events.Items {
			if event.Start.DateTime != "" && event.End.DateTime != "" {
				start, _ := time.Parse(time.RFC3339, event.Start.DateTime)
				end, _ := time.Parse(time.RFC3339, event.End.DateTime)
				
				// Get organizer info
				organizer := ""
				if event.Organizer != nil {
					if event.Organizer.DisplayName != "" {
						organizer = event.Organizer.DisplayName
					} else {
						organizer = event.Organizer.Email
					}
				}
				
				busyDetails = append(busyDetails, busyTime{
					Start:      start,
					End:        end,
					Summary:    event.Summary,
					Organizer:  organizer,
					CalendarId: calendarId,
				})
			}
		}
	}

	// Sort busy times by start time
	for i := 0; i < len(busyDetails); i++ {
		for j := i + 1; j < len(busyDetails); j++ {
			if busyDetails[i].Start.After(busyDetails[j].Start) {
				busyDetails[i], busyDetails[j] = busyDetails[j], busyDetails[i]
			}
		}
	}

	// Format results
	result := map[string]interface{}{
		"period": map[string]string{
			"start": startDate.Format("2006-01-02 15:04"),
			"end":   endDate.Format("2006-01-02 15:04"),
		},
		"calendars_checked": calendarsToCheck,
		"busy_times":        make([]map[string]interface{}, 0),
		"total_busy_times":  len(busyDetails),
	}

	// Add busy time details
	for _, busy := range busyDetails {
		busyInfo := map[string]interface{}{
			"start":     busy.Start.Format("2006-01-02 15:04"),
			"end":       busy.End.Format("2006-01-02 15:04"),
			"calendar":  busy.CalendarId,
			"summary":   busy.Summary,
			"organizer": busy.Organizer,
			"day":       busy.Start.Format("Monday"),
		}

		// Calculate duration
		duration := busy.End.Sub(busy.Start)
		busyInfo["duration_minutes"] = int(duration.Minutes())
		
		result["busy_times"] = append(result["busy_times"].([]map[string]interface{}), busyInfo)
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}