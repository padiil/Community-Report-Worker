package repository

import (
	"context"
	"fmt"
	"org-worker/internal/domain"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReportRepository struct {
	db *mongo.Database
}

func NewReportRepository(db *mongo.Database) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetReportByID(ctx context.Context, id string) (domain.ReportDoc, error) {
	var doc domain.ReportDoc
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return doc, err
	}
	coll := r.db.Collection("reports")
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&doc)
	return doc, err
}

func (r *ReportRepository) GetCommunityActivityData(ctx context.Context, filters map[string]interface{}) (domain.CommunityActivityData, error) {
	var data domain.CommunityActivityData
	var err error

	communityName, ok := filters["community_name"].(string)
	if !ok {
		return data, fmt.Errorf("filter 'community_name' hilang atau bukan string")
	}
	startDateStr, _ := filters["start_date"].(string)
	endDateStr, _ := filters["end_date"].(string)
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return data, fmt.Errorf("format 'start_date' salah: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return data, fmt.Errorf("format 'end_date' salah: %w", err)
	}

	data.CommunityName = communityName
	data.StartDate = startDate
	data.EndDate = endDate

	usersCollection := r.db.Collection("users")
	eventsCollection := r.db.Collection("events")
	attendancesCollection := r.db.Collection("attendances")

	newMemberFilter := bson.M{
		"communities": communityName,
		"createdAt":   bson.M{"$gte": startDate, "$lte": endDate},
	}
	if communityName == "all" {
		newMemberFilter = bson.M{"createdAt": bson.M{"$gte": startDate, "$lte": endDate}}
	}
	data.NewMemberCount, _ = usersCollection.CountDocuments(ctx, newMemberFilter)

	eventFilter := bson.M{
		"community": communityName,
		"date":      bson.M{"$gte": primitive.NewDateTimeFromTime(startDate), "$lte": primitive.NewDateTimeFromTime(endDate)},
	}
	if communityName == "all" {
		delete(eventFilter, "community")
	}
	cursor, err := eventsCollection.Find(ctx, eventFilter)
	if err != nil {
		return data, err
	}
	defer cursor.Close(ctx)

	type mongoTutor struct {
		Type   string `bson:"type"`
		UserID string `bson:"userID,omitempty"`
		Name   string `bson:"name"`
	}
	type mongoEvent struct {
		ID             primitive.ObjectID `bson:"_id"`
		Name           string             `bson:"name"`
		Community      string             `bson:"community"`
		Date           primitive.DateTime `bson:"date"`
		Tutor          mongoTutor         `bson:"tutor"`
		ImageJobIDs    []string           `bson:"imageJobIds,omitempty"`
		Documentations []string           `bson:"documentations,omitempty"`
	}
	type mongoUser struct {
		ID   string `bson:"_id"`
		Name string `bson:"name"`
	}
	var events []mongoEvent
	if err = cursor.All(ctx, &events); err != nil {
		return data, err
	}
	data.EventsHeldCount = len(events)
	data.EventDetails = make([]domain.EventDetail, 0, data.EventsHeldCount)

	imageJobIDSet := make(map[primitive.ObjectID]struct{})
	for _, e := range events {
		for _, idStr := range e.ImageJobIDs {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			if oid, convErr := primitive.ObjectIDFromHex(idStr); convErr == nil {
				imageJobIDSet[oid] = struct{}{}
			}
		}
	}
	imageJobOIDs := make([]primitive.ObjectID, 0, len(imageJobIDSet))
	for oid := range imageJobIDSet {
		imageJobOIDs = append(imageJobOIDs, oid)
	}
	resolvedImageURLs := r.resolveImageJobURLs(ctx, imageJobOIDs)

	eventIDs := make([]primitive.ObjectID, 0, len(events))
	for _, e := range events {
		eventIDs = append(eventIDs, e.ID)
	}
	activeFilter := bson.M{
		"eventID":       bson.M{"$in": eventIDs},
		"attendee.type": "Member",
	}
	if len(eventIDs) == 0 {
		data.ActiveMemberCount = 0
	} else {
		distinctUserIDs, _ := attendancesCollection.Distinct(ctx, "attendee.userID", activeFilter)
		var activeCount int64
		for _, raw := range distinctUserIDs {
			switch v := raw.(type) {
			case nil:
				continue
			case string:
				if v == "" {
					continue
				}
			case primitive.ObjectID:
				if v == primitive.NilObjectID {
					continue
				}
			}
			activeCount++
		}
		data.ActiveMemberCount = activeCount
	}

	tutorNameCache := make(map[string]string)
	for _, event := range events {
		count, _ := attendancesCollection.CountDocuments(ctx, bson.M{"eventID": event.ID})
		tutorName := strings.TrimSpace(event.Tutor.Name)
		userRef := strings.TrimSpace(event.Tutor.UserID)
		if tutorName == "" && userRef != "" {
			if cached, ok := tutorNameCache[userRef]; ok {
				tutorName = cached
			} else {
				var tutorDoc mongoUser
				objID, objErr := primitive.ObjectIDFromHex(userRef)
				var findErr error
				if objErr == nil {
					findErr = usersCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&tutorDoc)
				}
				if findErr != nil {
					findErr = usersCollection.FindOne(ctx, bson.M{"_id": userRef}).Decode(&tutorDoc)
				}
				if findErr == nil {
					tutorName = strings.TrimSpace(tutorDoc.Name)
				}
				tutorNameCache[userRef] = tutorName
			}
		}
		if tutorName == "" {
			tutorName = "N/A"
		}
		docs := make([]string, 0, len(event.ImageJobIDs)+len(event.Documentations))
		if len(event.ImageJobIDs) > 0 {
			for _, idStr := range event.ImageJobIDs {
				normalized := strings.ToLower(strings.TrimSpace(idStr))
				if normalized == "" {
					continue
				}
				if url, ok := resolvedImageURLs[normalized]; ok && url != "" {
					docs = append(docs, url)
				}
			}
		}
		if len(event.Documentations) > 0 {
			for _, url := range event.Documentations {
				trimmed := strings.TrimSpace(url)
				if trimmed == "" {
					continue
				}
				docs = append(docs, trimmed)
			}
		}
		data.EventDetails = append(data.EventDetails, domain.EventDetail{
			Name:              event.Name,
			Date:              event.Date.Time(),
			TutorName:         tutorName,
			ParticipantCount:  int(count),
			DocumentationURLs: docs,
		})
	}
	return data, nil
}

type facetResult struct {
	Total []struct {
		Count int64 `bson:"count"`
	} `bson:"total"`
	ByStatus   []domain.DemographicStat `bson:"byStatus"`
	ByAge      []domain.DemographicStat `bson:"byAge"`
	ByLocation []domain.DemographicStat `bson:"byLocation"`
}

func (r *ReportRepository) GetParticipantDemographicsData(ctx context.Context, filters map[string]interface{}) (domain.ParticipantDemographicsData, error) {
	var data domain.ParticipantDemographicsData
	var err error

	communityName, ok := filters["community_name"].(string)
	if !ok {
		return data, fmt.Errorf("filter 'community_name' hilang atau bukan string")
	}
	data.CommunityName = communityName

	usersCollection := r.db.Collection("users")
	matchStage := bson.M{}
	if communityName != "all" {
		matchStage = bson.M{"communities": communityName}
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchStage}},
		bson.D{{Key: "$facet", Value: bson.M{
			"total": bson.A{
				bson.M{"$count": "count"},
			},
			"byStatus": bson.A{
				bson.M{"$group": bson.M{"_id": "$statusPekerjaan", "count": bson.M{"$sum": 1}}},
				bson.M{"$sort": bson.M{"count": -1}},
			},
			"byAge": bson.A{
				bson.M{"$group": bson.M{"_id": "$kategoriUsia", "count": bson.M{"$sum": 1}}},
				bson.M{"$sort": bson.M{"count": -1}},
			},
			"byLocation": bson.A{
				bson.M{"$group": bson.M{"_id": "$domisili", "count": bson.M{"$sum": 1}}},
				bson.M{"$sort": bson.M{"count": -1}},
				bson.M{"$limit": 10},
			},
		}}},
	}

	cursor, err := usersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return data, err
	}
	defer cursor.Close(ctx)

	var results []facetResult
	if err = cursor.All(ctx, &results); err != nil {
		return data, err
	}
	if len(results) == 0 {
		return data, fmt.Errorf("tidak ada data demografi ditemukan")
	}
	result := results[0]
	if len(result.Total) > 0 {
		data.TotalParticipants = result.Total[0].Count
	}
	data.ByStatus = result.ByStatus
	data.ByAge = result.ByAge
	data.ByLocation = result.ByLocation

	return data, nil
}

func (r *ReportRepository) GetProgramImpactData(ctx context.Context, filters map[string]interface{}) (domain.ProgramImpactData, error) {
	var data domain.ProgramImpactData
	var err error

	communityName, ok := filters["community_name"].(string)
	if !ok {
		return data, fmt.Errorf("filter 'community_name' hilang atau bukan string")
	}
	startDateStr, _ := filters["start_date"].(string)
	endDateStr, _ := filters["end_date"].(string)
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return data, fmt.Errorf("format 'start_date' salah: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return data, fmt.Errorf("format 'end_date' salah: %w", err)
	}

	data.CommunityName = communityName
	data.StartDate = startDate
	data.EndDate = endDate

	milestonesCollection := r.db.Collection("milestones")

	matchStage := bson.M{
		"date": bson.M{"$gte": primitive.NewDateTimeFromTime(startDate), "$lte": primitive.NewDateTimeFromTime(endDate)},
	}
	if communityName != "all" {
		usersCollection := r.db.Collection("users")
		userCursor, err := usersCollection.Find(ctx, bson.M{"communities": communityName})
		if err != nil {
			return data, fmt.Errorf("gagal mengambil user untuk komunitas: %w", err)
		}
		defer userCursor.Close(ctx)
		type userRow struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		var userRows []userRow
		if err = userCursor.All(ctx, &userRows); err != nil {
			return data, fmt.Errorf("gagal decode user rows: %w", err)
		}
		userIDs := make([]primitive.ObjectID, 0, len(userRows))
		for _, u := range userRows {
			userIDs = append(userIDs, u.ID)
		}
		if len(userIDs) == 0 {
			data.Stats = []domain.MilestoneStat{}
			return data, nil
		}
		matchStage["userID"] = bson.M{"$in": userIDs}
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchStage}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$type",
			"count": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"count": -1}}},
	}

	cursor, err := milestonesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return data, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &data.Stats); err != nil {
		return data, err
	}

	highlights, err := r.fetchImpactHighlights(ctx, filters, communityName, startDate, endDate)
	if err != nil {
		return data, err
	}
	data.Highlights = highlights

	return data, nil
}

func (r *ReportRepository) fetchImpactHighlights(ctx context.Context, filters map[string]interface{}, communityName string, startDate, endDate time.Time) ([]domain.ImpactHighlight, error) {
	milestoneIDs := extractHighlightMilestoneIDs(filters)
	milestonesCollection := r.db.Collection("milestones")
	match := bson.M{
		"type": "project_submitted",
		"date": bson.M{"$gte": primitive.NewDateTimeFromTime(startDate), "$lte": primitive.NewDateTimeFromTime(endDate)},
	}
	if communityName != "all" {
		match["communityName"] = communityName
	}
	if len(milestoneIDs) > 0 {
		match["_id"] = bson.M{"$in": milestoneIDs}
	}
	findOpts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(3)
	if len(milestoneIDs) > 0 {
		findOpts.SetLimit(int64(len(milestoneIDs)))
	}
	cursor, err := milestonesCollection.Find(ctx, match, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type milestoneDoc struct {
		ID     primitive.ObjectID `bson:"_id"`
		UserID primitive.ObjectID `bson:"userID"`
		Detail struct {
			Title       string `bson:"title"`
			Summary     string `bson:"summary"`
			Description string `bson:"description"`
		} `bson:"detail"`
	}
	var rows []milestoneDoc
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	userIDSet := make(map[primitive.ObjectID]struct{})
	milestoneOIDList := make([]primitive.ObjectID, 0, len(rows))
	for _, row := range rows {
		milestoneOIDList = append(milestoneOIDList, row.ID)
		if row.UserID != primitive.NilObjectID {
			userIDSet[row.UserID] = struct{}{}
		}
	}

	userNames := r.fetchUserNames(ctx, userIDSet)
	assetURLs := r.fetchMilestoneAssetURLs(ctx, milestoneOIDList)

	highlights := make([]domain.ImpactHighlight, 0, len(rows))
	for _, row := range rows {
		title := strings.TrimSpace(row.Detail.Title)
		if title == "" {
			title = "Project Highlight"
		}
		summary := strings.TrimSpace(row.Detail.Summary)
		if summary == "" {
			summary = strings.TrimSpace(row.Detail.Description)
		}
		owner := userNames[row.UserID.Hex()]
		if owner == "" {
			owner = "-"
		}
		highlights = append(highlights, domain.ImpactHighlight{
			Title:             title,
			OwnerName:         owner,
			Summary:           summary,
			DocumentationURLs: assetURLs[row.ID.Hex()],
		})
	}

	return highlights, nil
}

func (r *ReportRepository) fetchUserNames(ctx context.Context, ids map[primitive.ObjectID]struct{}) map[string]string {
	result := make(map[string]string)
	if len(ids) == 0 {
		return result
	}
	userIDs := make([]primitive.ObjectID, 0, len(ids))
	for id := range ids {
		userIDs = append(userIDs, id)
	}
	usersCollection := r.db.Collection("users")
	cursor, err := usersCollection.Find(
		ctx,
		bson.M{"_id": bson.M{"$in": userIDs}},
		options.Find().SetProjection(bson.M{"name": 1}),
	)
	if err != nil {
		return result
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var user struct {
			ID   primitive.ObjectID `bson:"_id"`
			Name string             `bson:"name"`
		}
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		result[user.ID.Hex()] = strings.TrimSpace(user.Name)
	}
	return result
}

func (r *ReportRepository) fetchMilestoneAssetURLs(ctx context.Context, milestoneIDs []primitive.ObjectID) map[string][]string {
	result := make(map[string][]string)
	if len(milestoneIDs) == 0 {
		return result
	}
	assetsCollection := r.db.Collection("user_assets")
	cursor, err := assetsCollection.Find(
		ctx,
		bson.M{"milestoneID": bson.M{"$in": milestoneIDs}},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return result
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var asset struct {
			MilestoneID primitive.ObjectID `bson:"milestoneID"`
			FileURL     string             `bson:"fileURL"`
		}
		if err := cursor.Decode(&asset); err != nil {
			continue
		}
		url := strings.TrimSpace(asset.FileURL)
		if url == "" {
			continue
		}
		key := asset.MilestoneID.Hex()
		result[key] = append(result[key], url)
	}
	return result
}

func extractHighlightMilestoneIDs(filters map[string]interface{}) []primitive.ObjectID {
	if len(filters) == 0 {
		return nil
	}
	var ids []primitive.ObjectID
	keys := []string{"highlight_milestone_ids", "highlightMilestoneIds", "highlight_ids"}
	for _, key := range keys {
		raw, ok := filters[key]
		if !ok || raw == nil {
			continue
		}
		for _, val := range normalizeToStringSlice(raw) {
			if oid, err := primitive.ObjectIDFromHex(val); err == nil {
				ids = append(ids, oid)
			}
		}
		if len(ids) > 0 {
			break
		}
	}
	return ids
}

func normalizeToStringSlice(value interface{}) []string {
	var result []string
	switch v := value.(type) {
	case []string:
		for _, item := range v {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				result = append(result, trimmed)
			}
		}
	case []interface{}:
		for _, item := range v {
			if str := fmt.Sprint(item); strings.TrimSpace(str) != "" {
				result = append(result, strings.TrimSpace(str))
			}
		}
	case primitive.A:
		for _, item := range v {
			if str := fmt.Sprint(item); strings.TrimSpace(str) != "" {
				result = append(result, strings.TrimSpace(str))
			}
		}
	case string:
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	default:
		if value == nil {
			return result
		}
		if str := strings.TrimSpace(fmt.Sprint(value)); str != "" {
			result = append(result, str)
		}
	}
	return result
}

func (r *ReportRepository) resolveImageJobURLs(ctx context.Context, ids []primitive.ObjectID) map[string]string {
	result := make(map[string]string)
	if len(ids) == 0 {
		return result
	}
	imageJobsCollection := r.db.Collection("image_jobs")
	cursor, err := imageJobsCollection.Find(
		ctx,
		bson.M{"_id": bson.M{"$in": ids}, "status": "COMPLETED"},
		options.Find().SetProjection(bson.M{"outputImageURL": 1}),
	)
	if err != nil {
		return result
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var job struct {
			ID             primitive.ObjectID `bson:"_id"`
			OutputImageURL string             `bson:"outputImageURL"`
		}
		if err := cursor.Decode(&job); err != nil {
			continue
		}
		url := strings.TrimSpace(job.OutputImageURL)
		if url == "" {
			continue
		}
		result[strings.ToLower(job.ID.Hex())] = url
	}
	return result
}

type donationFacetResult struct {
	TotalCash []struct {
		Total float64 `bson:"total"`
	} `bson:"totalCash"`
	InKindTotal []struct {
		Total float64 `bson:"total"`
	} `bson:"inKindTotal"`
	BySource []domain.FinancialStat `bson:"bySource"`
	Top5     []domain.TopDonation   `bson:"top5Cash"`
}

type expenseFacetResult struct {
	Total []struct {
		Total float64 `bson:"total"`
	} `bson:"total"`
	ByCategory []domain.FinancialStat `bson:"byCategory"`
}

func (r *ReportRepository) GetFinancialSummaryData(ctx context.Context, filters map[string]interface{}) (domain.FinancialReportData, error) {
	var data domain.FinancialReportData
	var err error

	startDateStr, _ := filters["start_date"].(string)
	endDateStr, _ := filters["end_date"].(string)
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return data, fmt.Errorf("format 'start_date' salah: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return data, fmt.Errorf("format 'end_date' salah: %w", err)
	}

	data.StartDate = startDate
	data.EndDate = endDate

	donationsCollection := r.db.Collection("donations")
	expensesCollection := r.db.Collection("expenses")
	matchStage := bson.M{
		"date": bson.M{"$gte": primitive.NewDateTimeFromTime(startDate), "$lte": primitive.NewDateTimeFromTime(endDate)},
	}

	// Donations pipeline (new schema: donationType, cashDetails.amount, inKindDetails.estimatedValue)
	cashCondition := []interface{}{bson.M{"$eq": []interface{}{"$donationType", "Cash"}}, "$cashDetails.amount", 0}
	inKindCondition := []interface{}{bson.M{"$eq": []interface{}{"$donationType", "InKind"}}, "$inKindDetails.estimatedValue", 0}
	donationPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchStage}},
		bson.D{{Key: "$facet", Value: bson.M{
			"totalCash": bson.A{
				bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": bson.M{"$cond": cashCondition}}}},
			},
			"bySource": bson.A{
				bson.M{"$group": bson.M{"_id": "$source", "total": bson.M{"$sum": bson.M{"$cond": cashCondition}}}},
				bson.M{"$sort": bson.M{"total": -1}},
			},
			"top5Cash": bson.A{
				bson.M{"$match": bson.M{"donationType": "Cash"}},
				bson.M{"$sort": bson.M{"cashDetails.amount": -1}},
				bson.M{"$limit": 5},
				bson.M{"$project": bson.M{"source": 1, "amount": "$cashDetails.amount", "date": 1}},
			},
			"inKindTotal": bson.A{
				bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": bson.M{"$cond": inKindCondition}}}},
			},
		}}},
	}
	cursor, err := donationsCollection.Aggregate(ctx, donationPipeline)
	if err != nil {
		return data, fmt.Errorf("gagal agregasi donasi: %w", err)
	}
	var donationResults []donationFacetResult
	if err = cursor.All(ctx, &donationResults); err != nil {
		return data, err
	}
	if len(donationResults) > 0 {
		if len(donationResults[0].TotalCash) > 0 {
			data.TotalIncome = donationResults[0].TotalCash[0].Total
		}
		if len(donationResults[0].InKindTotal) > 0 {
			data.TotalInKindValue = donationResults[0].InKindTotal[0].Total
		}
		data.IncomeBySource = donationResults[0].BySource
		data.TopDonations = donationResults[0].Top5
	}
	cursor.Close(ctx)

	// Expenses pipeline
	expensePipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchStage}},
		bson.D{{Key: "$facet", Value: bson.M{
			"total": bson.A{
				bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$amount"}}},
			},
			"byCategory": bson.A{
				bson.M{"$group": bson.M{"_id": "$category", "total": bson.M{"$sum": "$amount"}}},
				bson.M{"$sort": bson.M{"total": -1}},
			},
		}}},
	}
	cursor, err = expensesCollection.Aggregate(ctx, expensePipeline)
	if err != nil {
		return data, fmt.Errorf("gagal agregasi pengeluaran: %w", err)
	}
	var expenseResults []expenseFacetResult
	if err = cursor.All(ctx, &expenseResults); err != nil {
		return data, err
	}
	if len(expenseResults) > 0 {
		if len(expenseResults[0].Total) > 0 {
			data.TotalExpenses = expenseResults[0].Total[0].Total
		}
		data.ExpensesByCategory = expenseResults[0].ByCategory
	}
	cursor.Close(ctx)

	data.NetIncome = data.TotalIncome - data.TotalExpenses
	return data, nil
}
