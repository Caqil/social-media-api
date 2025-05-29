// internal/services/report_service.go
package services

import (
	"context"
	"errors"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReportService struct {
	collection     *mongo.Collection
	userCollection *mongo.Collection
	postCollection *mongo.Collection
	db             *mongo.Database
}

func NewReportService() *ReportService {
	return &ReportService{
		collection:     config.DB.Collection("reports"),
		userCollection: config.DB.Collection("users"),
		postCollection: config.DB.Collection("posts"),
		db:             config.DB,
	}
}

// CreateReport creates a new report
func (rs *ReportService) CreateReport(reporterID primitive.ObjectID, req models.CreateReportRequest) (*models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert target ID
	targetID, err := primitive.ObjectIDFromHex(req.TargetID)
	if err != nil {
		return nil, errors.New("invalid target ID")
	}

	// Check if target exists
	if err := rs.validateTargetExists(req.TargetType, targetID); err != nil {
		return nil, err
	}

	// Check if user already reported this target
	existingCount, err := rs.collection.CountDocuments(ctx, bson.M{
		"reporter_id": reporterID,
		"target_type": req.TargetType,
		"target_id":   targetID,
		"status":      bson.M{"$nin": []string{"resolved", "rejected"}},
	})
	if err != nil {
		return nil, err
	}

	if existingCount > 0 {
		return nil, errors.New("you have already reported this content")
	}

	// Check if this target has been reported before
	reportedBefore, err := rs.collection.CountDocuments(ctx, bson.M{
		"target_type": req.TargetType,
		"target_id":   targetID,
	})
	if err != nil {
		return nil, err
	}

	// Create report
	report := &models.Report{
		ReporterID:     reporterID,
		TargetType:     req.TargetType,
		TargetID:       targetID,
		Reason:         req.Reason,
		Description:    req.Description,
		Category:       req.Category,
		Screenshots:    req.Screenshots,
		Evidence:       req.Evidence,
		ReportedBefore: reportedBefore > 0,
	}

	report.BeforeCreate()

	// Set priority based on reason
	report.Priority = rs.determinePriority(req.Reason, reportedBefore > 0)

	result, err := rs.collection.InsertOne(ctx, report)
	if err != nil {
		return nil, err
	}

	report.ID = result.InsertedID.(primitive.ObjectID)

	// Update target's report count
	go rs.updateTargetReportCount(req.TargetType, targetID, true)

	// Auto-assign high priority reports
	if report.Priority == "high" || report.Priority == "urgent" {
		go rs.autoAssignReport(report.ID)
	}

	// Populate reporter information
	rs.populateReportRelations(report)

	return report, nil
}

// GetReports retrieves reports with filtering and pagination
func (rs *ReportService) GetReports(filter models.ReportFilter, limit, skip int) ([]models.Report, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build filter
	mongoFilter := bson.M{}

	if filter.Status != "" {
		mongoFilter["status"] = filter.Status
	}
	if filter.TargetType != "" {
		mongoFilter["target_type"] = filter.TargetType
	}
	if filter.Reason != "" {
		mongoFilter["reason"] = filter.Reason
	}
	if filter.Priority != "" {
		mongoFilter["priority"] = filter.Priority
	}
	if filter.ReporterID != "" {
		if reporterID, err := primitive.ObjectIDFromHex(filter.ReporterID); err == nil {
			mongoFilter["reporter_id"] = reporterID
		}
	}
	if filter.AssignedTo != "" {
		if assignedTo, err := primitive.ObjectIDFromHex(filter.AssignedTo); err == nil {
			mongoFilter["assigned_to"] = assignedTo
		}
	}
	if filter.AutoDetected != nil {
		mongoFilter["auto_detected"] = *filter.AutoDetected
	}
	if filter.RequiresFollowUp != nil {
		mongoFilter["requires_follow_up"] = *filter.RequiresFollowUp
	}

	// Date range filtering
	if !filter.CreatedAfter.IsZero() || !filter.CreatedBefore.IsZero() {
		dateFilter := bson.M{}
		if !filter.CreatedAfter.IsZero() {
			dateFilter["$gte"] = filter.CreatedAfter
		}
		if !filter.CreatedBefore.IsZero() {
			dateFilter["$lte"] = filter.CreatedBefore
		}
		mongoFilter["created_at"] = dateFilter
	}

	// Get total count
	totalCount, err := rs.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, err
	}

	// Set sort order
	sortOrder := bson.M{"created_at": -1}
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "priority":
			sortOrder = bson.M{"priority": -1, "created_at": -1}
		case "status":
			sortOrder = bson.M{"status": 1, "created_at": -1}
		case "oldest":
			sortOrder = bson.M{"created_at": 1}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(sortOrder)

	cursor, err := rs.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reports []models.Report
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, 0, err
	}

	// Populate relations for all reports
	for i := range reports {
		rs.populateReportRelations(&reports[i])
	}

	return reports, totalCount, nil
}

// GetReportByID retrieves a specific report
func (rs *ReportService) GetReportByID(reportID primitive.ObjectID, currentUserID primitive.ObjectID) (*models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var report models.Report
	err := rs.collection.FindOne(ctx, bson.M{"_id": reportID}).Decode(&report)
	if err != nil {
		return nil, err
	}

	// Get current user's role
	var user models.User
	err = rs.userCollection.FindOne(ctx, bson.M{"_id": currentUserID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !report.CanViewReport(currentUserID, user.Role) {
		return nil, errors.New("access denied")
	}

	// Populate relations
	rs.populateReportRelations(&report)

	return &report, nil
}

// UpdateReport updates a report (moderator only)
func (rs *ReportService) UpdateReport(reportID, currentUserID primitive.ObjectID, req models.UpdateReportRequest) (*models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get current user's role
	var user models.User
	err := rs.userCollection.FindOne(ctx, bson.M{"_id": currentUserID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Get existing report
	var report models.Report
	err = rs.collection.FindOne(ctx, bson.M{"_id": reportID}).Decode(&report)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !report.CanEditReport(currentUserID, user.Role) {
		return nil, errors.New("access denied")
	}

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	// Update fields if provided
	if req.Status != nil {
		update["$set"].(bson.M)["status"] = *req.Status
	}
	if req.Priority != nil {
		update["$set"].(bson.M)["priority"] = *req.Priority
	}
	if req.AssignedTo != nil {
		if assignedTo, err := primitive.ObjectIDFromHex(*req.AssignedTo); err == nil {
			update["$set"].(bson.M)["assigned_to"] = assignedTo
			update["$set"].(bson.M)["status"] = models.ReportReviewing
		}
	}
	if req.Resolution != nil {
		update["$set"].(bson.M)["resolution"] = *req.Resolution
	}
	if req.ResolutionNote != nil {
		update["$set"].(bson.M)["resolution_note"] = *req.ResolutionNote
	}
	if req.ActionsTaken != nil {
		update["$set"].(bson.M)["actions_taken"] = req.ActionsTaken
	}
	if req.FollowUpDate != nil {
		update["$set"].(bson.M)["follow_up_date"] = *req.FollowUpDate
		update["$set"].(bson.M)["requires_follow_up"] = true
	}
	if req.FollowUpNote != nil {
		update["$set"].(bson.M)["follow_up_note"] = *req.FollowUpNote
	}

	_, err = rs.collection.UpdateOne(ctx, bson.M{"_id": reportID}, update)
	if err != nil {
		return nil, err
	}

	return rs.GetReportByID(reportID, currentUserID)
}

// AssignReport assigns a report to a moderator
func (rs *ReportService) AssignReport(reportID, assignedBy, assignedTo primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if report exists
	var report models.Report
	err := rs.collection.FindOne(ctx, bson.M{"_id": reportID}).Decode(&report)
	if err != nil {
		return err
	}

	// Check if assignee exists and is a moderator
	var assignee models.User
	err = rs.userCollection.FindOne(ctx, bson.M{"_id": assignedTo}).Decode(&assignee)
	if err != nil {
		return errors.New("assignee not found")
	}

	if assignee.Role != models.RoleModerator && assignee.Role != models.RoleAdmin && assignee.Role != models.RoleSuperAdmin {
		return errors.New("can only assign to moderators or admins")
	}

	// Update report
	update := bson.M{
		"$set": bson.M{
			"assigned_to": assignedTo,
			"status":      models.ReportReviewing,
			"updated_at":  time.Now(),
		},
	}

	_, err = rs.collection.UpdateOne(ctx, bson.M{"_id": reportID}, update)
	return err
}

// ResolveReport resolves a report
func (rs *ReportService) ResolveReport(reportID, resolvedBy primitive.ObjectID, resolution, note string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get report
	var report models.Report
	err := rs.collection.FindOne(ctx, bson.M{"_id": reportID}).Decode(&report)
	if err != nil {
		return err
	}

	// Get resolver's role
	var user models.User
	err = rs.userCollection.FindOne(ctx, bson.M{"_id": resolvedBy}).Decode(&user)
	if err != nil {
		return err
	}

	// Check permissions
	if !report.CanResolveReport(resolvedBy, user.Role) {
		return errors.New("access denied")
	}

	// Resolve report
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":          models.ReportResolved,
			"resolution":      resolution,
			"resolution_note": note,
			"resolved_by":     resolvedBy,
			"resolved_at":     now,
			"updated_at":      now,
		},
	}

	_, err = rs.collection.UpdateOne(ctx, bson.M{"_id": reportID}, update)
	if err != nil {
		return err
	}

	// Create action record
	go rs.addReportAction(reportID, "resolve", "Report resolved: "+resolution, resolvedBy, map[string]interface{}{
		"resolution": resolution,
		"note":       note,
	})

	// Notify reporter
	go rs.notifyReporter(report.ReporterID, reportID, "resolved")

	return nil
}

// RejectReport rejects a report
func (rs *ReportService) RejectReport(reportID, rejectedBy primitive.ObjectID, note string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get report
	var report models.Report
	err := rs.collection.FindOne(ctx, bson.M{"_id": reportID}).Decode(&report)
	if err != nil {
		return err
	}

	// Get rejector's role
	var user models.User
	err = rs.userCollection.FindOne(ctx, bson.M{"_id": rejectedBy}).Decode(&user)
	if err != nil {
		return err
	}

	// Check permissions
	if !report.CanResolveReport(rejectedBy, user.Role) {
		return errors.New("access denied")
	}

	// Reject report
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":          models.ReportRejected,
			"resolution_note": note,
			"resolved_by":     rejectedBy,
			"resolved_at":     now,
			"updated_at":      now,
		},
	}

	_, err = rs.collection.UpdateOne(ctx, bson.M{"_id": reportID}, update)
	if err != nil {
		return err
	}

	// Create action record
	go rs.addReportAction(reportID, "reject", "Report rejected: "+note, rejectedBy, map[string]interface{}{
		"note": note,
	})

	// Notify reporter
	go rs.notifyReporter(report.ReporterID, reportID, "rejected")

	return nil
}

// GetReportStats retrieves report statistics
func (rs *ReportService) GetReportStats() (*models.ReportStatsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregate statistics
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$status",
				"count": bson.M{"$sum": 1},
				"avg_resolution_time": bson.M{
					"$avg": bson.M{
						"$cond": []interface{}{
							bson.M{"$ne": []interface{}{"$resolved_at", nil}},
							bson.M{
								"$divide": []interface{}{
									bson.M{"$subtract": []interface{}{"$resolved_at", "$created_at"}},
									1000 * 60 * 60, // Convert to hours
								},
							},
							nil,
						},
					},
				},
			},
		},
	}

	cursor, err := rs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := &models.ReportStatsResponse{}

	for cursor.Next(ctx) {
		var result struct {
			ID                string  `bson:"_id"`
			Count             int64   `bson:"count"`
			AvgResolutionTime float64 `bson:"avg_resolution_time"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		switch result.ID {
		case "pending":
			stats.PendingReports = result.Count
		case "reviewing":
			stats.ReviewingReports = result.Count
		case "resolved":
			stats.ResolvedReports = result.Count
			if result.AvgResolutionTime > 0 {
				stats.AverageResolutionTime = result.AvgResolutionTime
			}
		case "rejected":
			stats.RejectedReports = result.Count
		}

		stats.TotalReports += result.Count
	}

	// Get high priority count
	highPriorityCount, _ := rs.collection.CountDocuments(ctx, bson.M{
		"priority": bson.M{"$in": []string{"high", "urgent"}},
		"status":   bson.M{"$nin": []string{"resolved", "rejected"}},
	})
	stats.HighPriorityReports = highPriorityCount

	// Get auto-detected count
	autoDetectedCount, _ := rs.collection.CountDocuments(ctx, bson.M{
		"auto_detected": true,
	})
	stats.AutoDetectedReports = autoDetectedCount

	return stats, nil
}

// GetReportSummary gets report summary by reason
func (rs *ReportService) GetReportSummary() ([]models.ReportSummaryResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get total count for percentage calculation
	totalReports, err := rs.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$reason",
				"count": bson.M{"$sum": 1},
				"resolved_count": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$status", "resolved"}},
							1,
							0,
						},
					},
				},
				"pending_count": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$status", "pending"}},
							1,
							0,
						},
					},
				},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
	}

	cursor, err := rs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var summaries []models.ReportSummaryResponse
	for cursor.Next(ctx) {
		var result struct {
			ID            models.ReportReason `bson:"_id"`
			Count         int64               `bson:"count"`
			ResolvedCount int64               `bson:"resolved_count"`
			PendingCount  int64               `bson:"pending_count"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		percentage := float64(0)
		if totalReports > 0 {
			percentage = (float64(result.Count) / float64(totalReports)) * 100
		}

		summaries = append(summaries, models.ReportSummaryResponse{
			Reason:        result.ID,
			Count:         result.Count,
			ResolvedCount: result.ResolvedCount,
			PendingCount:  result.PendingCount,
			Percentage:    percentage,
		})
	}

	return summaries, nil
}

// GetUserReports gets reports made by a specific user
func (rs *ReportService) GetUserReports(userID primitive.ObjectID, limit, skip int) ([]models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"reporter_id": userID}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := rs.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.Report
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, err
	}

	// Populate relations
	for i := range reports {
		rs.populateReportRelations(&reports[i])
	}

	return reports, nil
}

// GetReportsByTarget gets reports for a specific target
func (rs *ReportService) GetReportsByTarget(targetType string, targetID primitive.ObjectID, limit, skip int) ([]models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"target_type": targetType,
		"target_id":   targetID,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := rs.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.Report
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, err
	}

	// Populate relations
	for i := range reports {
		rs.populateReportRelations(&reports[i])
	}

	return reports, nil
}

// Helper methods

func (rs *ReportService) validateTargetExists(targetType string, targetID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var collection *mongo.Collection
	switch targetType {
	case "post":
		collection = rs.db.Collection("posts")
	case "comment":
		collection = rs.db.Collection("comments")
	case "user":
		collection = rs.userCollection
	case "story":
		collection = rs.db.Collection("stories")
	case "group":
		collection = rs.db.Collection("groups")
	case "event":
		collection = rs.db.Collection("events")
	case "message":
		collection = rs.db.Collection("messages")
	default:
		return errors.New("invalid target type")
	}

	count, err := collection.CountDocuments(ctx, bson.M{"_id": targetID})
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("target not found")
	}

	return nil
}

func (rs *ReportService) determinePriority(reason models.ReportReason, reportedBefore bool) string {
	switch reason {
	case models.ReportViolence, models.ReportHateSpeech:
		return "urgent"
	case models.ReportHarassment, models.ReportNudity:
		return "high"
	case models.ReportSpam, models.ReportFakeNews:
		if reportedBefore {
			return "high"
		}
		return "medium"
	default:
		return "medium"
	}
}

func (rs *ReportService) populateReportRelations(report *models.Report) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Populate reporter
	var reporter models.User
	if err := rs.userCollection.FindOne(ctx, bson.M{"_id": report.ReporterID}).Decode(&reporter); err == nil {
		report.Reporter = reporter.ToUserResponse()
	}

	// Populate assigned moderator
	if report.AssignedTo != nil {
		var moderator models.User
		if err := rs.userCollection.FindOne(ctx, bson.M{"_id": *report.AssignedTo}).Decode(&moderator); err == nil {
			report.AssignedModerator = moderator.ToUserResponse()
		}
	}

	// Populate resolving moderator
	if report.ResolvedBy != nil {
		var resolver models.User
		if err := rs.userCollection.FindOne(ctx, bson.M{"_id": *report.ResolvedBy}).Decode(&resolver); err == nil {
			report.ResolvingModerator = resolver.ToUserResponse()
		}
	}
}

func (rs *ReportService) updateTargetReportCount(targetType string, targetID primitive.ObjectID, increment bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	var collection *mongo.Collection
	switch targetType {
	case "post":
		collection = rs.db.Collection("posts")
	case "comment":
		collection = rs.db.Collection("comments")
	case "user":
		collection = rs.userCollection
	case "story":
		collection = rs.db.Collection("stories")
	default:
		return
	}

	collection.UpdateOne(ctx, bson.M{"_id": targetID}, bson.M{
		"$inc": bson.M{"reports_count": value},
		"$set": bson.M{
			"is_reported": increment,
			"updated_at":  time.Now(),
		},
	})
}

func (rs *ReportService) autoAssignReport(reportID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find available moderator with least assigned reports
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"role":      bson.M{"$in": []string{"moderator", "admin", "super_admin"}},
				"is_active": true,
			},
		},
		{
			"$lookup": bson.M{
				"from": "reports",
				"let":  bson.M{"userId": "$_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr":  bson.M{"$eq": []interface{}{"$assigned_to", "$$userId"}},
							"status": bson.M{"$nin": []string{"resolved", "rejected"}},
						},
					},
					{"$count": "assignedCount"},
				},
				"as": "assignedReports",
			},
		},
		{
			"$addFields": bson.M{
				"assignedCount": bson.M{
					"$ifNull": []interface{}{
						bson.M{"$arrayElemAt": []interface{}{"$assignedReports.assignedCount", 0}},
						0,
					},
				},
			},
		},
		{
			"$sort": bson.M{"assignedCount": 1},
		},
		{
			"$limit": 1,
		},
	}

	cursor, err := rs.userCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var moderator models.User
		if err := cursor.Decode(&moderator); err == nil {
			rs.AssignReport(reportID, primitive.NilObjectID, moderator.ID)
		}
	}
}

func (rs *ReportService) addReportAction(reportID primitive.ObjectID, actionType, description string, takenBy primitive.ObjectID, details map[string]interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	action := models.ReportAction{
		ID:          primitive.NewObjectID(),
		Type:        actionType,
		Description: description,
		TakenBy:     takenBy,
		TakenAt:     time.Now(),
		Details:     details,
	}

	rs.collection.UpdateOne(ctx, bson.M{"_id": reportID}, bson.M{
		"$push": bson.M{"actions_taken": action},
	})
}

func (rs *ReportService) notifyReporter(reporterID, reportID primitive.ObjectID, status string) {
	// This would integrate with notification service
	// Implementation depends on notification system
}
func (us *ReportService) GetCollection() *mongo.Collection {
	return us.collection
}
