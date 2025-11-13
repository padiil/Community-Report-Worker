package repository

import (
	"context"
	"fmt"
	"org-worker/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReportRepository struct {
	db *mongo.Database
}

func NewReportRepository(db *mongo.Database) *ReportRepository {
	return &ReportRepository{db: db}
}

// Ambil dokumen report berdasarkan ID (untuk pola dokumen pelacak)
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

// Ambil data aktivitas komunitas (menggunakan skema baru)
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

	// New members: createdAt in range AND communities contains communityName
	newMemberFilter := bson.M{
		"communities": communityName,
		"createdAt":   bson.M{"$gte": startDate, "$lte": endDate},
	}
	if communityName == "all" {
		newMemberFilter = bson.M{"createdAt": bson.M{"$gte": startDate, "$lte": endDate}}
	}
	data.NewMemberCount, _ = usersCollection.CountDocuments(ctx, newMemberFilter)

	// Fetch events first (communityName & date range)
	eventFilter := bson.M{
		"communityName": communityName,
		"date":          bson.M{"$gte": primitive.NewDateTimeFromTime(startDate), "$lte": primitive.NewDateTimeFromTime(endDate)},
	}
	if communityName == "all" {
		delete(eventFilter, "communityName")
	}
	cursor, err := eventsCollection.Find(ctx, eventFilter)
	if err != nil {
		return data, err
	}
	defer cursor.Close(ctx)

	type mongoTutor struct {
		Type   string             `bson:"type"`
		UserID primitive.ObjectID `bson:"userID,omitempty"`
		Name   string             `bson:"name"`
	}
	type mongoEvent struct {
		ID        primitive.ObjectID `bson:"_id"`
		Name      string             `bson:"name"`
		Community string             `bson:"communityName"`
		Date      primitive.DateTime `bson:"date"`
		Tutor     mongoTutor         `bson:"tutor"`
	}
	var events []mongoEvent
	if err = cursor.All(ctx, &events); err != nil {
		return data, err
	}
	data.EventsHeldCount = len(events)
	data.EventDetails = make([]domain.EventDetail, 0, data.EventsHeldCount)

	// Collect event IDs for active member calculation
	eventIDs := make([]primitive.ObjectID, 0, len(events))
	for _, e := range events {
		eventIDs = append(eventIDs, e.ID)
	}
	// Active members: distinct attendee.userID where attendee.type == 'Member' across these events
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

	// Per-event participant count (Members + Guests)
	for _, event := range events {
		count, _ := attendancesCollection.CountDocuments(ctx, bson.M{"eventID": event.ID})
		tutorName := event.Tutor.Name
		data.EventDetails = append(data.EventDetails, domain.EventDetail{
			Name:             event.Name,
			Date:             event.Date.Time(),
			TutorName:        tutorName,
			ParticipantCount: int(count),
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

	// Build match stage without community (removed from milestones). If a specific community is requested, first gather userIDs belonging to that community.
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

	return data, nil
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
	donationPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchStage}},
		bson.D{{Key: "$facet", Value: bson.M{
			"totalCash": bson.A{
				bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$donationType", "Cash"}}, "$cashDetails.amount", 0}}}}},
			},
			"bySource": bson.A{
				bson.M{"$group": bson.M{"_id": "$source", "total": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$donationType", "Cash"}}, "$cashDetails.amount", 0}}}}},
				bson.M{"$sort": bson.M{"total": -1}},
			},
			"top5Cash": bson.A{
				bson.M{"$match": bson.M{"donationType": "Cash"}},
				bson.M{"$sort": bson.M{"cashDetails.amount": -1}},
				bson.M{"$limit": 5},
				bson.M{"$project": bson.M{"source": 1, "amount": "$cashDetails.amount", "date": 1}},
			},
			"inKindTotal": bson.A{
				bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$donationType", "InKind"}}, "$inKindDetails.estimatedValue", 0}}}}},
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
