// models/report.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Report represents a report made by users about content or other users
type Report struct {
	BaseModel `bson:",inline"`

	// Reporter Information
	ReporterID primitive.ObjectID `json:"reporter_id" bson:"reporter_id" validate:"required"`
	Reporter   UserResponse       `json:"reporter,omitempty" bson:"-"` // Populated when querying

	// Reported Content/User
	TargetType string             `json:"target_type" bson:"target_type" validate:"required"` // user, post, comment, story, group, event, message
	TargetID   primitive.ObjectID `json:"target_id" bson:"target_id" validate:"required"`

	// Report Details
	Reason      ReportReason `json:"reason" bson:"reason" validate:"required"`
	Description string       `json:"description,omitempty" bson:"description,omitempty" validate:"max=1000"`
	Category    string       `json:"category,omitempty" bson:"category,omitempty"`

	// Evidence
	Screenshots []MediaInfo            `json:"screenshots,omitempty" bson:"screenshots,omitempty"`
	Evidence    map[string]interface{} `json:"evidence,omitempty" bson:"evidence,omitempty"`

	// Report Status and Processing
	Status            ReportStatus        `json:"status" bson:"status"`
	Priority          string              `json:"priority" bson:"priority"` // low, medium, high, urgent
	AssignedTo        *primitive.ObjectID `json:"assigned_to,omitempty" bson:"assigned_to,omitempty"`
	AssignedModerator UserResponse        `json:"assigned_moderator,omitempty" bson:"-"` // Populated when querying

	// Resolution
	Resolution         string              `json:"resolution,omitempty" bson:"resolution,omitempty"`
	ResolutionNote     string              `json:"resolution_note,omitempty" bson:"resolution_note,omitempty" validate:"max=2000"`
	ResolvedAt         *time.Time          `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
	ResolvedBy         *primitive.ObjectID `json:"resolved_by,omitempty" bson:"resolved_by,omitempty"`
	ResolvingModerator UserResponse        `json:"resolving_moderator,omitempty" bson:"-"` // Populated when querying

	// Actions Taken
	ActionsTaken   []ReportAction `json:"actions_taken,omitempty" bson:"actions_taken,omitempty"`
	Warning        bool           `json:"warning" bson:"warning"`
	ContentRemoved bool           `json:"content_removed" bson:"content_removed"`
	UserSuspended  bool           `json:"user_suspended" bson:"user_suspended"`
	AccountBanned  bool           `json:"account_banned" bson:"account_banned"`

	// Follow-up
	RequiresFollowUp bool       `json:"requires_follow_up" bson:"requires_follow_up"`
	FollowUpDate     *time.Time `json:"follow_up_date,omitempty" bson:"follow_up_date,omitempty"`
	FollowUpNote     string     `json:"follow_up_note,omitempty" bson:"follow_up_note,omitempty"`

	// Internal tracking
	SimilarReports []primitive.ObjectID `json:"similar_reports,omitempty" bson:"similar_reports,omitempty"`
	ReportedBefore bool                 `json:"reported_before" bson:"reported_before"`
	AutoDetected   bool                 `json:"auto_detected" bson:"auto_detected"`

	// Reporter Feedback
	ReporterNotified bool   `json:"reporter_notified" bson:"reporter_notified"`
	ReporterFeedback string `json:"reporter_feedback,omitempty" bson:"reporter_feedback,omitempty"`
	FeedbackRating   int    `json:"feedback_rating,omitempty" bson:"feedback_rating,omitempty"` // 1-5 stars

	// Additional Context
	Source    string                 `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api, auto
	IPAddress string                 `json:"-" bson:"ip_address,omitempty"`
	UserAgent string                 `json:"-" bson:"user_agent,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// ReportAction represents an action taken in response to a report
type ReportAction struct {
	ID          primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Type        string                 `json:"type" bson:"type"` // warning, content_removal, suspension, ban, dismiss
	Description string                 `json:"description" bson:"description"`
	TakenBy     primitive.ObjectID     `json:"taken_by" bson:"taken_by"`
	TakenAt     time.Time              `json:"taken_at" bson:"taken_at"`
	Details     map[string]interface{} `json:"details,omitempty" bson:"details,omitempty"`
}

// ReportResponse represents report data returned in API responses
type ReportResponse struct {
	ID                 string         `json:"id"`
	ReporterID         string         `json:"reporter_id"`
	Reporter           UserResponse   `json:"reporter,omitempty"`
	TargetType         string         `json:"target_type"`
	TargetID           string         `json:"target_id"`
	Reason             ReportReason   `json:"reason"`
	Description        string         `json:"description,omitempty"`
	Category           string         `json:"category,omitempty"`
	Screenshots        []MediaInfo    `json:"screenshots,omitempty"`
	Status             ReportStatus   `json:"status"`
	Priority           string         `json:"priority"`
	AssignedTo         string         `json:"assigned_to,omitempty"`
	AssignedModerator  UserResponse   `json:"assigned_moderator,omitempty"`
	Resolution         string         `json:"resolution,omitempty"`
	ResolutionNote     string         `json:"resolution_note,omitempty"`
	ResolvedAt         *time.Time     `json:"resolved_at,omitempty"`
	ResolvedBy         string         `json:"resolved_by,omitempty"`
	ResolvingModerator UserResponse   `json:"resolving_moderator,omitempty"`
	ActionsTaken       []ReportAction `json:"actions_taken,omitempty"`
	Warning            bool           `json:"warning"`
	ContentRemoved     bool           `json:"content_removed"`
	UserSuspended      bool           `json:"user_suspended"`
	AccountBanned      bool           `json:"account_banned"`
	RequiresFollowUp   bool           `json:"requires_follow_up"`
	FollowUpDate       *time.Time     `json:"follow_up_date,omitempty"`
	ReportedBefore     bool           `json:"reported_before"`
	AutoDetected       bool           `json:"auto_detected"`
	ReporterNotified   bool           `json:"reporter_notified"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	TimeAgo            string         `json:"time_ago,omitempty"`
}

// CreateReportRequest represents the request to create a report
type CreateReportRequest struct {
	TargetType  string                 `json:"target_type" validate:"required,oneof=user post comment story group event message"`
	TargetID    string                 `json:"target_id" validate:"required"`
	Reason      ReportReason           `json:"reason" validate:"required"`
	Description string                 `json:"description,omitempty" validate:"max=1000"`
	Category    string                 `json:"category,omitempty"`
	Screenshots []MediaInfo            `json:"screenshots,omitempty"`
	Evidence    map[string]interface{} `json:"evidence,omitempty"`
}

// UpdateReportRequest represents the request to update a report (by moderators)
type UpdateReportRequest struct {
	Status         *ReportStatus  `json:"status,omitempty"`
	Priority       *string        `json:"priority,omitempty" validate:"omitempty,oneof=low medium high urgent"`
	AssignedTo     *string        `json:"assigned_to,omitempty"`
	Resolution     *string        `json:"resolution,omitempty"`
	ResolutionNote *string        `json:"resolution_note,omitempty" validate:"omitempty,max=2000"`
	ActionsTaken   []ReportAction `json:"actions_taken,omitempty"`
	FollowUpDate   *time.Time     `json:"follow_up_date,omitempty"`
	FollowUpNote   *string        `json:"follow_up_note,omitempty"`
}

// ReportStatsResponse represents report statistics
type ReportStatsResponse struct {
	TotalReports          int64   `json:"total_reports"`
	PendingReports        int64   `json:"pending_reports"`
	ReviewingReports      int64   `json:"reviewing_reports"`
	ResolvedReports       int64   `json:"resolved_reports"`
	RejectedReports       int64   `json:"rejected_reports"`
	HighPriorityReports   int64   `json:"high_priority_reports"`
	AutoDetectedReports   int64   `json:"auto_detected_reports"`
	AverageResolutionTime float64 `json:"average_resolution_time"` // in hours
}

// ReportSummaryResponse represents a summary of reports by type/reason
type ReportSummaryResponse struct {
	Reason        ReportReason `json:"reason"`
	Count         int64        `json:"count"`
	ResolvedCount int64        `json:"resolved_count"`
	PendingCount  int64        `json:"pending_count"`
	Percentage    float64      `json:"percentage"`
}

// Methods for Report model

// BeforeCreate sets default values before creating report
func (r *Report) BeforeCreate() {
	r.BaseModel.BeforeCreate()

	// Set default values
	r.Status = ReportPending
	r.Priority = "medium"
	r.Warning = false
	r.ContentRemoved = false
	r.UserSuspended = false
	r.AccountBanned = false
	r.RequiresFollowUp = false
	r.ReportedBefore = false
	r.AutoDetected = false
	r.ReporterNotified = false

	// Set source if not provided
	if r.Source == "" {
		r.Source = "web"
	}
}

// ToReportResponse converts Report to ReportResponse
func (r *Report) ToReportResponse() ReportResponse {
	response := ReportResponse{
		ID:               r.ID.Hex(),
		ReporterID:       r.ReporterID.Hex(),
		TargetType:       r.TargetType,
		TargetID:         r.TargetID.Hex(),
		Reason:           r.Reason,
		Description:      r.Description,
		Category:         r.Category,
		Screenshots:      r.Screenshots,
		Status:           r.Status,
		Priority:         r.Priority,
		Resolution:       r.Resolution,
		ResolutionNote:   r.ResolutionNote,
		ResolvedAt:       r.ResolvedAt,
		ActionsTaken:     r.ActionsTaken,
		Warning:          r.Warning,
		ContentRemoved:   r.ContentRemoved,
		UserSuspended:    r.UserSuspended,
		AccountBanned:    r.AccountBanned,
		RequiresFollowUp: r.RequiresFollowUp,
		FollowUpDate:     r.FollowUpDate,
		ReportedBefore:   r.ReportedBefore,
		AutoDetected:     r.AutoDetected,
		ReporterNotified: r.ReporterNotified,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}

	if r.AssignedTo != nil {
		response.AssignedTo = r.AssignedTo.Hex()
	}

	if r.ResolvedBy != nil {
		response.ResolvedBy = r.ResolvedBy.Hex()
	}

	return response
}

// AssignToModerator assigns the report to a moderator
func (r *Report) AssignToModerator(moderatorID primitive.ObjectID) {
	r.AssignedTo = &moderatorID
	r.Status = ReportReviewing
	r.BeforeUpdate()
}

// Resolve resolves the report
func (r *Report) Resolve(resolvedBy primitive.ObjectID, resolution string, note string) {
	r.Status = ReportResolved
	r.Resolution = resolution
	r.ResolutionNote = note
	r.ResolvedBy = &resolvedBy
	now := time.Now()
	r.ResolvedAt = &now
	r.BeforeUpdate()
}

// Reject rejects the report
func (r *Report) Reject(rejectedBy primitive.ObjectID, note string) {
	r.Status = ReportRejected
	r.ResolutionNote = note
	r.ResolvedBy = &rejectedBy
	now := time.Now()
	r.ResolvedAt = &now
	r.BeforeUpdate()
}

// AddAction adds an action taken for this report
func (r *Report) AddAction(actionType, description string, takenBy primitive.ObjectID, details map[string]interface{}) {
	action := ReportAction{
		ID:          primitive.NewObjectID(),
		Type:        actionType,
		Description: description,
		TakenBy:     takenBy,
		TakenAt:     time.Now(),
		Details:     details,
	}

	r.ActionsTaken = append(r.ActionsTaken, action)

	// Update relevant flags based on action type
	switch actionType {
	case "warning":
		r.Warning = true
	case "content_removal":
		r.ContentRemoved = true
	case "suspension":
		r.UserSuspended = true
	case "ban":
		r.AccountBanned = true
	}

	r.BeforeUpdate()
}

// CanViewReport checks if a user can view this report
func (r *Report) CanViewReport(currentUserID primitive.ObjectID, userRole UserRole) bool {
	// Reporter can view their own report
	if r.ReporterID == currentUserID {
		return true
	}

	// Moderators and admins can view any report
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// CanEditReport checks if a user can edit this report
func (r *Report) CanEditReport(currentUserID primitive.ObjectID, userRole UserRole) bool {
	// Only moderators and admins can edit reports
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// CanResolveReport checks if a user can resolve this report
func (r *Report) CanResolveReport(currentUserID primitive.ObjectID, userRole UserRole) bool {
	// Must be assigned to the user or be admin/super admin
	if r.AssignedTo != nil && *r.AssignedTo == currentUserID {
		return true
	}

	return userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// IsHighPriority checks if the report is high priority
func (r *Report) IsHighPriority() bool {
	return r.Priority == "high" || r.Priority == "urgent"
}

// IsResolved checks if the report has been resolved
func (r *Report) IsResolved() bool {
	return r.Status == ReportResolved || r.Status == ReportRejected
}

// GetResolutionTime returns the time taken to resolve the report
func (r *Report) GetResolutionTime() *time.Duration {
	if r.ResolvedAt == nil {
		return nil
	}

	duration := r.ResolvedAt.Sub(r.CreatedAt)
	return &duration
}

// MarkAsNotified marks that the reporter has been notified
func (r *Report) MarkAsNotified() {
	r.ReporterNotified = true
	r.BeforeUpdate()
}

// SetPriority sets the priority of the report
func (r *Report) SetPriority(priority string) {
	validPriorities := []string{"low", "medium", "high", "urgent"}
	for _, valid := range validPriorities {
		if priority == valid {
			r.Priority = priority
			r.BeforeUpdate()
			return
		}
	}
}

// RequireFollowUp marks the report as requiring follow-up
func (r *Report) RequireFollowUp(followUpDate time.Time, note string) {
	r.RequiresFollowUp = true
	r.FollowUpDate = &followUpDate
	r.FollowUpNote = note
	r.BeforeUpdate()
}

// CompleteFollowUp marks the follow-up as completed
func (r *Report) CompleteFollowUp() {
	r.RequiresFollowUp = false
	r.FollowUpDate = nil
	r.FollowUpNote = ""
	r.BeforeUpdate()
}

// GetActionHistory returns a formatted history of actions taken
func (r *Report) GetActionHistory() []string {
	var history []string

	for _, action := range r.ActionsTaken {
		entry := action.Type + ": " + action.Description + " (" + action.TakenAt.Format("2006-01-02 15:04:05") + ")"
		history = append(history, entry)
	}

	return history
}

// Utility functions for Report

// GetReportReasons returns all available report reasons
func GetReportReasons() []ReportReason {
	return []ReportReason{
		ReportSpam,
		ReportHarassment,
		ReportHateSpeech,
		ReportViolence,
		ReportNudity,
		ReportFakeNews,
		ReportCopyright,
		ReportOther,
	}
}

// GetReportReasonText returns human-readable text for report reasons
func GetReportReasonText(reason ReportReason) string {
	switch reason {
	case ReportSpam:
		return "Spam"
	case ReportHarassment:
		return "Harassment or Bullying"
	case ReportHateSpeech:
		return "Hate Speech"
	case ReportViolence:
		return "Violence or Threats"
	case ReportNudity:
		return "Nudity or Sexual Content"
	case ReportFakeNews:
		return "False Information"
	case ReportCopyright:
		return "Copyright Violation"
	case ReportOther:
		return "Other"
	default:
		return "Unknown"
	}
}
