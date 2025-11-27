package domain

import "time"

type EventDetail struct {
	Name              string    `json:"name"`
	Date              time.Time `json:"date"`
	TutorName         string    `json:"tutorName"`
	ParticipantCount  int       `json:"participantCount"`
	DocumentationURLs []string  `json:"documentationURLs,omitempty"`
}

type CommunityActivityData struct {
	CommunityName     string        `json:"communityName"`
	StartDate         time.Time     `json:"startDate"`
	EndDate           time.Time     `json:"endDate"`
	NewMemberCount    int64         `json:"newMemberCount"`
	ActiveMemberCount int64         `json:"activeMemberCount"`
	EventsHeldCount   int           `json:"eventsHeldCount"`
	EventDetails      []EventDetail `json:"eventDetails"`
}

type ReportJobPayload struct {
	ReportID   string                 `json:"reportID"`
	ReportType string                 `json:"reportType"`
	Filters    map[string]interface{} `json:"filters"`
}

type DemographicStat struct {
	ID    string `bson:"_id" json:"id"`
	Count int    `bson:"count" json:"count"`
}

type ParticipantDemographicsData struct {
	CommunityName     string            `json:"communityName"`
	TotalParticipants int64             `json:"totalParticipants"`
	ByStatus          []DemographicStat `json:"byStatus"`
	ByAge             []DemographicStat `json:"byAge"`
	ByLocation        []DemographicStat `json:"byLocation"`
}

type MilestoneStat struct {
	ID    string `bson:"_id" json:"id"`
	Count int    `bson:"count" json:"count"`
}

type ProgramImpactData struct {
	CommunityName string          `json:"communityName"`
	StartDate     time.Time       `json:"startDate"`
	EndDate       time.Time       `json:"endDate"`
	Stats         []MilestoneStat `json:"stats"`
}

type FinancialStat struct {
	ID    string  `bson:"_id" json:"id"`
	Total float64 `bson:"total" json:"total"`
}

type TopDonation struct {
	Source string    `bson:"source" json:"source"`
	Amount float64   `bson:"amount" json:"amount"`
	Date   time.Time `bson:"date" json:"date"`
}

type FinancialReportData struct {
	StartDate          time.Time       `json:"startDate"`
	EndDate            time.Time       `json:"endDate"`
	TotalIncome        float64         `json:"totalIncome"`
	TotalInKindValue   float64         `json:"totalInKindValue"`
	TotalExpenses      float64         `json:"totalExpenses"`
	NetIncome          float64         `json:"netIncome"`
	ExpensesByCategory []FinancialStat `json:"expensesByCategory"`
	IncomeBySource     []FinancialStat `json:"incomeBySource"`
	TopDonations       []TopDonation   `json:"topDonations"`
}

type User struct {
	ID          string    `json:"id" bson:"_id"`
	Communities []string  `json:"communities" bson:"communities"`
	Roles       []string  `json:"roles" bson:"roles"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

type Tutor struct {
	Type   string `json:"type" bson:"type"` // Internal | External
	UserID string `json:"userID,omitempty" bson:"userID,omitempty"`
	Name   string `json:"name" bson:"name"`
}

type Event struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Community string    `json:"communityName" bson:"communityName"`
	Date      time.Time `json:"date" bson:"date"`
	Tutor     Tutor     `json:"tutor" bson:"tutor"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Attendee struct {
	Type   string `json:"type" bson:"type"` // Member | Guest
	UserID string `json:"userID,omitempty" bson:"userID,omitempty"`
	Name   string `json:"name,omitempty" bson:"name,omitempty"`
}

type Attendance struct {
	ID       string   `json:"id" bson:"_id"`
	EventID  string   `json:"eventID" bson:"eventID"`
	Attendee Attendee `json:"attendee" bson:"attendee"`
	// No direct date; derive via linked Event
}

type Donation struct {
	ID           string `json:"id" bson:"_id"`
	DonationType string `json:"donationType" bson:"donationType"` // Cash | InKind
	Source       string `json:"source" bson:"source"`
	CashDetails  *struct {
		Amount float64 `json:"amount" bson:"amount"`
	} `json:"cashDetails,omitempty" bson:"cashDetails,omitempty"`
	InKindDetails *struct {
		EstimatedValue float64 `json:"estimatedValue" bson:"estimatedValue"`
		Description    string  `json:"description" bson:"description"`
	} `json:"inKindDetails,omitempty" bson:"inKindDetails,omitempty"`
	Date time.Time `json:"date" bson:"date"`
}
