package data

import (
	"errors"
	"fmt"
	"time"

	"github.com/mappcpd/web-services/internal/members"
)

// MemberActivity is a simpler representation of the member activity than the nested one in the current REST api.
type MemberActivity struct {
	ID          int64     `json:"id"`
	Date        time.Time `json:"date"`
	Credit      float64   `json:"credit"`
	CategoryID  int64     `json:"categoryId"`
	Category    string    `json:"category"`
	TypeID      int64     `json:"typeId"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
}

// GetMemberActivities fetches activities for a member
func GetMemberActivities(memberID int64) ([]MemberActivity, error) {
	var xa []MemberActivity

	// This returns a nested struct which we can simplify
	xma, err := members.MemberActivitiesByMemberID(memberID)
	for _, v := range xma {
		a := MemberActivity{
			ID:          v.ID,
			Date:        v.DateISO,
			Credit:      v.Credit,
			CategoryID:  v.Category.ID,
			Category:    v.Category.Name,
			TypeID:      v.Activity.ID,
			Type:        v.Activity.Name,
			Description: v.Description,
		}
		xa = append(xa, a)
	}

	return xa, err
}

// Unpack an object into a value of type MemberActivity
func (ma *MemberActivity) Unpack(obj map[string]interface{}) {
	if val, ok := obj["id"].(int64); ok {
		ma.ID = val
	}
	if val, ok := obj["date"].(string); ok {
		// todo: should handle date string error?
		d, _ := time.Parse(time.RFC3339, val)
		ma.Date = d
	}
	if val, ok := obj["credit"].(float64); ok {
		ma.Credit = val
	}
	if val, ok := obj["categoryId"].(int64); ok {
		ma.CategoryID = val
	}
	if val, ok := obj["typeId"].(int64); ok {
		ma.TypeID = val
	}
	if val, ok := obj["description"].(string); ok {
		ma.Description = val
	}
}

// GetMemberActivity fetches a single activities by id.
// It verifies that the activity is owned by the member by memberID.
func GetMemberActivity(memberID, activityID int64) (MemberActivity, error) {

	var a MemberActivity

	// This returns a nested struct which we can simplify
	ma, err := members.MemberActivityByID(activityID)
	if err != nil {
		return a, err
	}

	// Verify owner match
	if ma.MemberID != memberID {
		msg := fmt.Sprintf("MemberActivity with id %v does not belong to member with id %v", activityID, memberID)
		return a, errors.New(msg)
	}

	a.ID = ma.ID
	a.Date = ma.DateISO
	a.Credit = ma.Credit
	a.CategoryID = ma.Category.ID
	a.Category = ma.Category.Name
	a.TypeID = ma.Activity.ID
	a.Type = ma.Activity.Name
	a.Description = ma.Description

	return a, nil
}

// AddMemberActivity adds a member activity
func AddMemberActivity(memberID int64, memberActivity MemberActivity) (MemberActivity, error) {

	// Create the required type for the insert
	// todo: add evidence and categoryId
	ma := members.MemberActivityRow{
		MemberID:    memberID,
		ActivityID:  memberActivity.TypeID,
		Date:        memberActivity.Date.String(),
		Quantity:    memberActivity.Credit,
		Description: memberActivity.Description,
	}

	// A return value for the new record
	var mar MemberActivity

	// This just returns the new record id, so re-fetch the member activity record
	// so that all the fields are populated for the response.
	newID, err := members.AddMemberActivity(ma)
	if err != nil {
		return mar, err
	}

	return GetMemberActivity(memberID, newID)

}

// UpdateMemberActivity adds a member activity
func UpdateMemberActivity(ma MemberActivity) {

}