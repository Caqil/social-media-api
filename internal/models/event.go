// models/event.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents an event that can be created by users or groups
type Event struct {
	BaseModel `bson:",inline"`

	// Basic Information
	Title       string `json:"title" bson:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" bson:"description" validate:"max=5000"`
	Slug        string `json:"slug" bson:"slug" validate:"required,min=3,max=200"` // URL-friendly title

	// Visual Identity
	CoverImage string      `json:"cover_image" bson:"cover_image"`
	Images     []MediaInfo `json:"images,omitempty" bson:"images,omitempty"`

	// Event Details
	Category string   `json:"category" bson:"category" validate:"required,max=50"`
	Tags     []string `json:"tags,omitempty" bson:"tags,omitempty"`
	Type     string   `json:"type" bson:"type"` // online, offline, hybrid

	// Timing
	StartTime      time.Time `json:"start_time" bson:"start_time" validate:"required"`
	EndTime        time.Time `json:"end_time" bson:"end_time" validate:"required"`
	Timezone       string    `json:"timezone" bson:"timezone"`
	IsAllDay       bool      `json:"is_all_day" bson:"is_all_day"`
	IsRecurring    bool      `json:"is_recurring" bson:"is_recurring"`
	RecurrenceRule string    `json:"recurrence_rule,omitempty" bson:"recurrence_rule,omitempty"` // RRULE format

	// Location
	Location       *Location `json:"location,omitempty" bson:"location,omitempty"`
	OnlineEventURL string    `json:"online_event_url,omitempty" bson:"online_event_url,omitempty"`
	VenueDetails   string    `json:"venue_details,omitempty" bson:"venue_details,omitempty"`

	// Organization
	CreatedBy  primitive.ObjectID   `json:"created_by" bson:"created_by" validate:"required"`
	Creator    UserResponse         `json:"creator,omitempty" bson:"-"` // Populated when querying
	GroupID    *primitive.ObjectID  `json:"group_id,omitempty" bson:"group_id,omitempty"`
	Group      *GroupResponse       `json:"group,omitempty" bson:"-"` // Populated when querying
	Organizers []primitive.ObjectID `json:"organizers,omitempty" bson:"organizers,omitempty"`
	CoHosts    []primitive.ObjectID `json:"co_hosts,omitempty" bson:"co_hosts,omitempty"`

	// Event Settings
	Status            EventStatus  `json:"status" bson:"status"`
	Privacy           PrivacyLevel `json:"privacy" bson:"privacy"`
	MaxAttendees      int64        `json:"max_attendees,omitempty" bson:"max_attendees,omitempty"`
	RequireApproval   bool         `json:"require_approval" bson:"require_approval"`
	AllowGuestInvites bool         `json:"allow_guest_invites" bson:"allow_guest_invites"`
	AllowComments     bool         `json:"allow_comments" bson:"allow_comments"`
	AllowPhotos       bool         `json:"allow_photos" bson:"allow_photos"`

	// RSVP Statistics
	AttendeesCount  int64 `json:"attendees_count" bson:"attendees_count"`
	InterestedCount int64 `json:"interested_count" bson:"interested_count"`
	GoingCount      int64 `json:"going_count" bson:"going_count"`
	MaybeCount      int64 `json:"maybe_count" bson:"maybe_count"`
	NotGoingCount   int64 `json:"not_going_count" bson:"not_going_count"`
	InvitedCount    int64 `json:"invited_count" bson:"invited_count"`

	// Engagement
	ViewsCount    int64 `json:"views_count" bson:"views_count"`
	SharesCount   int64 `json:"shares_count" bson:"shares_count"`
	CommentsCount int64 `json:"comments_count" bson:"comments_count"`

	// Additional Details
	ExternalURL    string                 `json:"external_url,omitempty" bson:"external_url,omitempty"`
	TicketURL      string                 `json:"ticket_url,omitempty" bson:"ticket_url,omitempty"`
	Price          *EventPrice            `json:"price,omitempty" bson:"price,omitempty"`
	AgeRestriction string                 `json:"age_restriction,omitempty" bson:"age_restriction,omitempty"`
	DressCode      string                 `json:"dress_code,omitempty" bson:"dress_code,omitempty"`
	ContactInfo    *EventContact          `json:"contact_info,omitempty" bson:"contact_info,omitempty"`
	CustomFields   map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`

	// Reminders
	RemindersSent []EventReminder `json:"reminders_sent,omitempty" bson:"reminders_sent,omitempty"`

	// Content Moderation
	IsReported     bool   `json:"is_reported" bson:"is_reported"`
	ReportsCount   int64  `json:"reports_count" bson:"reports_count"`
	IsHidden       bool   `json:"is_hidden" bson:"is_hidden"`
	ModerationNote string `json:"moderation_note,omitempty" bson:"moderation_note,omitempty"`
}

// EventPrice represents pricing information for an event
type EventPrice struct {
	IsFree      bool    `json:"is_free" bson:"is_free"`
	Currency    string  `json:"currency,omitempty" bson:"currency,omitempty"`
	Amount      float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	Description string  `json:"description,omitempty" bson:"description,omitempty"`
}

// EventContact represents contact information for an event
type EventContact struct {
	Email   string `json:"email,omitempty" bson:"email,omitempty"`
	Phone   string `json:"phone,omitempty" bson:"phone,omitempty"`
	Website string `json:"website,omitempty" bson:"website,omitempty"`
}

// EventReminder represents a reminder sent for an event
type EventReminder struct {
	Type            string    `json:"type" bson:"type"` // email, push, sms
	SentAt          time.Time `json:"sent_at" bson:"sent_at"`
	TimeBeforeEvent int64     `json:"time_before_event" bson:"time_before_event"` // Minutes before event
}

// EventRSVP represents a user's RSVP to an event
type EventRSVP struct {
	BaseModel `bson:",inline"`

	EventID primitive.ObjectID `json:"event_id" bson:"event_id" validate:"required"`
	UserID  primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`

	// RSVP Details
	Status     RSVPStatus `json:"status" bson:"status" validate:"required"`
	Response   string     `json:"response,omitempty" bson:"response,omitempty" validate:"max=500"`
	GuestCount int        `json:"guest_count" bson:"guest_count"`

	// Event and User info (populated when querying)
	Event EventResponse `json:"event,omitempty" bson:"-"`
	User  UserResponse  `json:"user,omitempty" bson:"-"`

	// RSVP Tracking
	InvitedBy   *primitive.ObjectID `json:"invited_by,omitempty" bson:"invited_by,omitempty"`
	InvitedAt   *time.Time          `json:"invited_at,omitempty" bson:"invited_at,omitempty"`
	RespondedAt time.Time           `json:"responded_at" bson:"responded_at"`

	// Attendance Tracking
	CheckedIn   bool       `json:"checked_in" bson:"checked_in"`
	CheckedInAt *time.Time `json:"checked_in_at,omitempty" bson:"checked_in_at,omitempty"`
	NoShow      bool       `json:"no_show" bson:"no_show"`

	// Notifications
	RemindersSent []EventReminder `json:"reminders_sent,omitempty" bson:"reminders_sent,omitempty"`

	// Additional Data
	Source     string                 `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api
	CustomData map[string]interface{} `json:"custom_data,omitempty" bson:"custom_data,omitempty"`
}

// EventInvite represents an invitation to an event
type EventInvite struct {
	BaseModel `bson:",inline"`

	EventID   primitive.ObjectID `json:"event_id" bson:"event_id" validate:"required"`
	InviterID primitive.ObjectID `json:"inviter_id" bson:"inviter_id" validate:"required"`
	InviteeID primitive.ObjectID `json:"invitee_id" bson:"invitee_id" validate:"required"`

	// Invite Details
	Message   string    `json:"message,omitempty" bson:"message,omitempty" validate:"max=500"`
	Status    string    `json:"status" bson:"status"` // pending, accepted, declined, expired
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`

	// Response Tracking
	RespondedAt *time.Time `json:"responded_at,omitempty" bson:"responded_at,omitempty"`

	// Populated when querying
	Event   EventResponse `json:"event,omitempty" bson:"-"`
	Inviter UserResponse  `json:"inviter,omitempty" bson:"-"`
	Invitee UserResponse  `json:"invitee,omitempty" bson:"-"`
}

// Response Models

// EventResponse represents the event data returned in API responses
type EventResponse struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Slug              string         `json:"slug"`
	CoverImage        string         `json:"cover_image"`
	Images            []MediaInfo    `json:"images,omitempty"`
	Category          string         `json:"category"`
	Tags              []string       `json:"tags,omitempty"`
	Type              string         `json:"type"`
	StartTime         time.Time      `json:"start_time"`
	EndTime           time.Time      `json:"end_time"`
	Timezone          string         `json:"timezone"`
	IsAllDay          bool           `json:"is_all_day"`
	IsRecurring       bool           `json:"is_recurring"`
	RecurrenceRule    string         `json:"recurrence_rule,omitempty"`
	Location          *Location      `json:"location,omitempty"`
	OnlineEventURL    string         `json:"online_event_url,omitempty"`
	VenueDetails      string         `json:"venue_details,omitempty"`
	CreatedBy         string         `json:"created_by"`
	Creator           UserResponse   `json:"creator"`
	GroupID           string         `json:"group_id,omitempty"`
	Group             *GroupResponse `json:"group,omitempty"`
	Status            EventStatus    `json:"status"`
	Privacy           PrivacyLevel   `json:"privacy"`
	MaxAttendees      int64          `json:"max_attendees,omitempty"`
	RequireApproval   bool           `json:"require_approval"`
	AllowGuestInvites bool           `json:"allow_guest_invites"`
	AllowComments     bool           `json:"allow_comments"`
	AllowPhotos       bool           `json:"allow_photos"`
	AttendeesCount    int64          `json:"attendees_count"`
	InterestedCount   int64          `json:"interested_count"`
	GoingCount        int64          `json:"going_count"`
	MaybeCount        int64          `json:"maybe_count"`
	NotGoingCount     int64          `json:"not_going_count"`
	InvitedCount      int64          `json:"invited_count"`
	ViewsCount        int64          `json:"views_count"`
	SharesCount       int64          `json:"shares_count"`
	CommentsCount     int64          `json:"comments_count"`
	ExternalURL       string         `json:"external_url,omitempty"`
	TicketURL         string         `json:"ticket_url,omitempty"`
	Price             *EventPrice    `json:"price,omitempty"`
	AgeRestriction    string         `json:"age_restriction,omitempty"`
	DressCode         string         `json:"dress_code,omitempty"`
	ContactInfo       *EventContact  `json:"contact_info,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`

	// User-specific context
	UserRSVP       RSVPStatus `json:"user_rsvp,omitempty"`
	CanEdit        bool       `json:"can_edit,omitempty"`
	CanInvite      bool       `json:"can_invite,omitempty"`
	CanModerate    bool       `json:"can_moderate,omitempty"`
	IsOrganizer    bool       `json:"is_organizer,omitempty"`
	IsCoHost       bool       `json:"is_co_host,omitempty"`
	TimeUntilStart string     `json:"time_until_start,omitempty"`
	TimeAgo        string     `json:"time_ago,omitempty"`
}

// EventRSVPResponse represents RSVP data
type EventRSVPResponse struct {
	ID          string       `json:"id"`
	EventID     string       `json:"event_id"`
	UserID      string       `json:"user_id"`
	User        UserResponse `json:"user"`
	Status      RSVPStatus   `json:"status"`
	Response    string       `json:"response,omitempty"`
	GuestCount  int          `json:"guest_count"`
	RespondedAt time.Time    `json:"responded_at"`
	CheckedIn   bool         `json:"checked_in"`
	CheckedInAt *time.Time   `json:"checked_in_at,omitempty"`
	NoShow      bool         `json:"no_show"`
	TimeAgo     string       `json:"time_ago,omitempty"`
}

// EventInviteResponse represents event invite data
type EventInviteResponse struct {
	ID        string        `json:"id"`
	EventID   string        `json:"event_id"`
	Event     EventResponse `json:"event"`
	InviterID string        `json:"inviter_id"`
	Inviter   UserResponse  `json:"inviter"`
	InviteeID string        `json:"invitee_id"`
	Invitee   UserResponse  `json:"invitee,omitempty"`
	Message   string        `json:"message,omitempty"`
	Status    string        `json:"status"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	TimeAgo   string        `json:"time_ago,omitempty"`
}

// Request Models

// CreateEventRequest represents the request to create an event
type CreateEventRequest struct {
	Title             string                 `json:"title" validate:"required,min=3,max=200"`
	Description       string                 `json:"description" validate:"max=5000"`
	Category          string                 `json:"category" validate:"required,max=50"`
	Tags              []string               `json:"tags,omitempty"`
	Type              string                 `json:"type" validate:"required,oneof=online offline hybrid"`
	StartTime         time.Time              `json:"start_time" validate:"required"`
	EndTime           time.Time              `json:"end_time" validate:"required"`
	Timezone          string                 `json:"timezone" validate:"required"`
	IsAllDay          bool                   `json:"is_all_day"`
	IsRecurring       bool                   `json:"is_recurring"`
	RecurrenceRule    string                 `json:"recurrence_rule,omitempty"`
	Location          *Location              `json:"location,omitempty"`
	OnlineEventURL    string                 `json:"online_event_url,omitempty"`
	VenueDetails      string                 `json:"venue_details,omitempty"`
	GroupID           string                 `json:"group_id,omitempty"`
	Privacy           PrivacyLevel           `json:"privacy" validate:"required,oneof=public friends private"`
	MaxAttendees      int64                  `json:"max_attendees,omitempty" validate:"omitempty,min=1"`
	RequireApproval   bool                   `json:"require_approval"`
	AllowGuestInvites bool                   `json:"allow_guest_invites"`
	AllowComments     bool                   `json:"allow_comments"`
	AllowPhotos       bool                   `json:"allow_photos"`
	ExternalURL       string                 `json:"external_url,omitempty"`
	TicketURL         string                 `json:"ticket_url,omitempty"`
	Price             *EventPrice            `json:"price,omitempty"`
	AgeRestriction    string                 `json:"age_restriction,omitempty"`
	DressCode         string                 `json:"dress_code,omitempty"`
	ContactInfo       *EventContact          `json:"contact_info,omitempty"`
	CustomFields      map[string]interface{} `json:"custom_fields,omitempty"`
}

// UpdateEventRequest represents the request to update an event
type UpdateEventRequest struct {
	Title             *string                `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Description       *string                `json:"description,omitempty" validate:"omitempty,max=5000"`
	Category          *string                `json:"category,omitempty" validate:"omitempty,max=50"`
	Tags              []string               `json:"tags,omitempty"`
	Type              *string                `json:"type,omitempty" validate:"omitempty,oneof=online offline hybrid"`
	StartTime         *time.Time             `json:"start_time,omitempty"`
	EndTime           *time.Time             `json:"end_time,omitempty"`
	Timezone          *string                `json:"timezone,omitempty"`
	IsAllDay          *bool                  `json:"is_all_day,omitempty"`
	Location          *Location              `json:"location,omitempty"`
	OnlineEventURL    *string                `json:"online_event_url,omitempty"`
	VenueDetails      *string                `json:"venue_details,omitempty"`
	Privacy           *PrivacyLevel          `json:"privacy,omitempty" validate:"omitempty,oneof=public friends private"`
	MaxAttendees      *int64                 `json:"max_attendees,omitempty" validate:"omitempty,min=1"`
	RequireApproval   *bool                  `json:"require_approval,omitempty"`
	AllowGuestInvites *bool                  `json:"allow_guest_invites,omitempty"`
	AllowComments     *bool                  `json:"allow_comments,omitempty"`
	AllowPhotos       *bool                  `json:"allow_photos,omitempty"`
	ExternalURL       *string                `json:"external_url,omitempty"`
	TicketURL         *string                `json:"ticket_url,omitempty"`
	Price             *EventPrice            `json:"price,omitempty"`
	AgeRestriction    *string                `json:"age_restriction,omitempty"`
	DressCode         *string                `json:"dress_code,omitempty"`
	ContactInfo       *EventContact          `json:"contact_info,omitempty"`
	CustomFields      map[string]interface{} `json:"custom_fields,omitempty"`
}

// RSVPToEventRequest represents the request to RSVP to an event
type RSVPToEventRequest struct {
	Status     RSVPStatus `json:"status" validate:"required,oneof=going maybe not_going"`
	Response   string     `json:"response,omitempty" validate:"max=500"`
	GuestCount int        `json:"guest_count,omitempty" validate:"min=0,max=10"`
}

// InviteToEventRequest represents the request to invite users to an event
type InviteToEventRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1,max=50"`
	Message string   `json:"message,omitempty" validate:"max=500"`
}

// Methods for Event model

// BeforeCreate sets default values before creating event
func (e *Event) BeforeCreate() {
	e.BaseModel.BeforeCreate()

	// Set default values
	e.Status = EventPublished
	e.AttendeesCount = 0
	e.InterestedCount = 0
	e.GoingCount = 0
	e.MaybeCount = 0
	e.NotGoingCount = 0
	e.InvitedCount = 0
	e.ViewsCount = 0
	e.SharesCount = 0
	e.CommentsCount = 0
	e.IsReported = false
	e.ReportsCount = 0
	e.IsHidden = false

	// Set default settings
	e.RequireApproval = false
	e.AllowGuestInvites = true
	e.AllowComments = true
	e.AllowPhotos = true

	// Generate slug from title if not provided
	if e.Slug == "" {
		e.Slug = generateSlug(e.Title)
	}

	// Set default timezone if not provided
	if e.Timezone == "" {
		e.Timezone = "UTC"
	}
}

// ToEventResponse converts Event model to EventResponse
func (e *Event) ToEventResponse() EventResponse {
	response := EventResponse{
		ID:                e.ID.Hex(),
		Title:             e.Title,
		Description:       e.Description,
		Slug:              e.Slug,
		CoverImage:        e.CoverImage,
		Images:            e.Images,
		Category:          e.Category,
		Tags:              e.Tags,
		Type:              e.Type,
		StartTime:         e.StartTime,
		EndTime:           e.EndTime,
		Timezone:          e.Timezone,
		IsAllDay:          e.IsAllDay,
		IsRecurring:       e.IsRecurring,
		RecurrenceRule:    e.RecurrenceRule,
		Location:          e.Location,
		OnlineEventURL:    e.OnlineEventURL,
		VenueDetails:      e.VenueDetails,
		CreatedBy:         e.CreatedBy.Hex(),
		Status:            e.Status,
		Privacy:           e.Privacy,
		MaxAttendees:      e.MaxAttendees,
		RequireApproval:   e.RequireApproval,
		AllowGuestInvites: e.AllowGuestInvites,
		AllowComments:     e.AllowComments,
		AllowPhotos:       e.AllowPhotos,
		AttendeesCount:    e.AttendeesCount,
		InterestedCount:   e.InterestedCount,
		GoingCount:        e.GoingCount,
		MaybeCount:        e.MaybeCount,
		NotGoingCount:     e.NotGoingCount,
		InvitedCount:      e.InvitedCount,
		ViewsCount:        e.ViewsCount,
		SharesCount:       e.SharesCount,
		CommentsCount:     e.CommentsCount,
		ExternalURL:       e.ExternalURL,
		TicketURL:         e.TicketURL,
		Price:             e.Price,
		AgeRestriction:    e.AgeRestriction,
		DressCode:         e.DressCode,
		ContactInfo:       e.ContactInfo,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}

	if e.GroupID != nil {
		response.GroupID = e.GroupID.Hex()
	}

	return response
}

// IncrementGoingCount increments the going count
func (e *Event) IncrementGoingCount() {
	e.GoingCount++
	e.UpdateAttendeesCount()
	e.BeforeUpdate()
}

// DecrementGoingCount decrements the going count
func (e *Event) DecrementGoingCount() {
	if e.GoingCount > 0 {
		e.GoingCount--
	}
	e.UpdateAttendeesCount()
	e.BeforeUpdate()
}

// IncrementMaybeCount increments the maybe count
func (e *Event) IncrementMaybeCount() {
	e.MaybeCount++
	e.BeforeUpdate()
}

// DecrementMaybeCount decrements the maybe count
func (e *Event) DecrementMaybeCount() {
	if e.MaybeCount > 0 {
		e.MaybeCount--
	}
	e.BeforeUpdate()
}

// UpdateAttendeesCount updates the total attendees count
func (e *Event) UpdateAttendeesCount() {
	e.AttendeesCount = e.GoingCount + e.MaybeCount
}

// CanViewEvent checks if a user can view this event
func (e *Event) CanViewEvent(currentUserID primitive.ObjectID, isInvited bool, isGroupMember bool) bool {
	// Check if event is published and not hidden
	if e.Status != EventPublished || e.IsHidden {
		return e.CreatedBy == currentUserID // Only creator can view unpublished/hidden events
	}

	// Check privacy settings
	switch e.Privacy {
	case PrivacyPublic:
		return true
	case PrivacyFriends:
		return isInvited || isGroupMember
	case PrivacyPrivate:
		return isInvited
	default:
		return false
	}
}

// CanEditEvent checks if a user can edit this event
func (e *Event) CanEditEvent(currentUserID primitive.ObjectID) bool {
	// Creator can edit
	if e.CreatedBy == currentUserID {
		return true
	}

	// Check if user is organizer
	for _, organizerID := range e.Organizers {
		if organizerID == currentUserID {
			return true
		}
	}

	// Check if user is co-host
	for _, coHostID := range e.CoHosts {
		if coHostID == currentUserID {
			return true
		}
	}

	return false
}

// CanInviteToEvent checks if a user can invite others to this event
func (e *Event) CanInviteToEvent(currentUserID primitive.ObjectID) bool {
	// Check if guest invites are allowed
	if !e.AllowGuestInvites {
		return e.CanEditEvent(currentUserID)
	}

	// Any attendee can invite if guest invites are allowed
	return true
}

// IsUpcoming checks if the event is upcoming
func (e *Event) IsUpcoming() bool {
	return e.StartTime.After(time.Now())
}

// IsOngoing checks if the event is currently ongoing
func (e *Event) IsOngoing() bool {
	now := time.Now()
	return e.StartTime.Before(now) && e.EndTime.After(now)
}

// IsCompleted checks if the event has completed
func (e *Event) IsCompleted() bool {
	return e.EndTime.Before(time.Now())
}

// GetEventDuration returns the duration of the event
func (e *Event) GetEventDuration() time.Duration {
	return e.EndTime.Sub(e.StartTime)
}

// Methods for EventRSVP model

// BeforeCreate sets default values before creating RSVP
func (er *EventRSVP) BeforeCreate() {
	er.BaseModel.BeforeCreate()

	// Set default values
	er.RespondedAt = er.CreatedAt
	er.CheckedIn = false
	er.NoShow = false
	er.GuestCount = 0
}

// ToEventRSVPResponse converts EventRSVP to EventRSVPResponse
func (er *EventRSVP) ToEventRSVPResponse() EventRSVPResponse {
	return EventRSVPResponse{
		ID:          er.ID.Hex(),
		EventID:     er.EventID.Hex(),
		UserID:      er.UserID.Hex(),
		Status:      er.Status,
		Response:    er.Response,
		GuestCount:  er.GuestCount,
		RespondedAt: er.RespondedAt,
		CheckedIn:   er.CheckedIn,
		CheckedInAt: er.CheckedInAt,
		NoShow:      er.NoShow,
	}
}

// CheckIn marks the user as checked in to the event
func (er *EventRSVP) CheckIn() {
	er.CheckedIn = true
	now := time.Now()
	er.CheckedInAt = &now
	er.BeforeUpdate()
}

// MarkAsNoShow marks the user as a no-show
func (er *EventRSVP) MarkAsNoShow() {
	er.NoShow = true
	er.BeforeUpdate()
}

// Methods for EventInvite model

// BeforeCreate sets default values before creating event invite
func (ei *EventInvite) BeforeCreate() {
	ei.BaseModel.BeforeCreate()

	// Set default values
	ei.Status = "pending"
	ei.ExpiresAt = ei.CreatedAt.Add(7 * 24 * time.Hour) // Expires in 7 days
}

// ToEventInviteResponse converts EventInvite to EventInviteResponse
func (ei *EventInvite) ToEventInviteResponse() EventInviteResponse {
	return EventInviteResponse{
		ID:        ei.ID.Hex(),
		EventID:   ei.EventID.Hex(),
		InviterID: ei.InviterID.Hex(),
		InviteeID: ei.InviteeID.Hex(),
		Message:   ei.Message,
		Status:    ei.Status,
		ExpiresAt: ei.ExpiresAt,
		CreatedAt: ei.CreatedAt,
	}
}

// Accept accepts the event invitation
func (ei *EventInvite) Accept() {
	ei.Status = "accepted"
	now := time.Now()
	ei.RespondedAt = &now
	ei.BeforeUpdate()
}

// Decline declines the event invitation
func (ei *EventInvite) Decline() {
	ei.Status = "declined"
	now := time.Now()
	ei.RespondedAt = &now
	ei.BeforeUpdate()
}

// IsExpired checks if the invitation has expired
func (ei *EventInvite) IsExpired() bool {
	return time.Now().After(ei.ExpiresAt)
}

// Utility functions

// GetEventCategories returns available event categories
func GetEventCategories() []string {
	return []string{
		"business",
		"technology",
		"education",
		"health",
		"arts",
		"music",
		"sports",
		"entertainment",
		"food",
		"travel",
		"lifestyle",
		"community",
		"charity",
		"networking",
		"workshop",
		"conference",
		"meetup",
		"party",
		"festival",
		"other",
	}
}

// IsValidEventCategory checks if a category is valid
func IsValidEventCategory(category string) bool {
	validCategories := GetEventCategories()
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

// GetEventTypes returns available event types
func GetEventTypes() []string {
	return []string{"online", "offline", "hybrid"}
}

// IsValidEventType checks if an event type is valid
func IsValidEventType(eventType string) bool {
	validTypes := GetEventTypes()
	for _, valid := range validTypes {
		if eventType == valid {
			return true
		}
	}
	return false
}
