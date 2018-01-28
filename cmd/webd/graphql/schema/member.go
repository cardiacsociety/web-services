package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mappcpd/web-services/cmd/webd/graphql/data"
	"github.com/mappcpd/web-services/internal/platform/jwt"
	"github.com/mappcpd/web-services/internal/utility"
)

// MemberUser is exported field attached to the root query. It is a top-level 'viewer' query field that ensures data
// is restricted to the member (user) identified by the token.
var MemberUser = &graphql.Field{
	Description: "The memberUser field acts as a 'viewer' and requires a valid JSON Web Token (JWT)[https://jwt.io]. " +
		"Access to data in sub-fields is restricted to that belonging to the member identified by the token.",
	Type: memberType,
	Args: graphql.FieldConfigArgument{
		"token": &graphql.ArgumentConfig{
			Type:        &graphql.NonNull{OfType: graphql.String},
			Description: "Valid JSON web token",
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		token, ok := p.Args["token"].(string)
		if ok {
			// Validate the token, and extract the member id
			at, err := jwt.Check(token)
			if err != nil {
				return nil, err
			}
			//fmt.Println(at.Claims)
			id := at.Claims.ID
			// At this point we have a valid token from which we've extracted an id.
			// As a final step we can verify that the id is a valid user in the system,
			// for example, that it is active. Although this is a bit redundant for each request?
			return data.GetMember(id)
		}
		return nil, nil
	},
}

// member represents the main member node , and provides a path to the child nodes.
var memberType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "member",
	Description: "Member query object that provides general profile information as well as additional sub-query fields.",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.String,
			Description: "The member's unique id number",
		},
		"active": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Boolean flag indicating if the member is currently active in the system",
		},
		"title": &graphql.Field{
			Type:        graphql.String,
			Description: "The member's membership title",
		},
		"firstName": &graphql.Field{
			Type:        graphql.String,
			Description: "The member's first name",
		},
		"middleNames": &graphql.Field{
			Type:        graphql.String,
			Description: "One or more middle names",
		},
		"lastName": &graphql.Field{
			Type:        graphql.String,
			Description: "The member's surname / family name",
		},
		"postNominal": &graphql.Field{
			Type:        graphql.String,
			Description: "Option string of preferred post nominals, eg 'Ph.D', 'OAM' etc",
		},
		"dateOfBirth": &graphql.Field{
			Type:        graphql.String,
			Description: "The member's date of birth, as a string value",
		},
		"email": &graphql.Field{
			Type:        graphql.String,
			Description: "The member's primary email address",
		},
		"mobile": &graphql.Field{
			Type:        graphql.String,
			Description: "The member's mobile phone number",
		},
		"locations": &graphql.Field{
			Type:        graphql.NewList(memberLocationType),
			Description: "One or more contact locations",
		},
		"qualifications": &graphql.Field{
			Type:        graphql.NewList(qualificationType),
			Description: "The member's qualifications",
		},
		"positions": &graphql.Field{
			Type:        graphql.NewList(positionType),
			Description: "The member's positions or appointments to committees, councils etc",
		},

		// sub queries
		"activity":    memberActivity,
		"activities":  memberActivities,
		"evaluation":  memberCurrentEvaluation,
		"evaluations": memberEvaluations,
	},
})

// memberActivity query field fetches a single memberActivity that belongs to a member
var memberActivity = &graphql.Field{
	Description: "Fetches a single member memberActivity by memberActivity id.",
	Type:        memberActivityType,
	Args: graphql.FieldConfigArgument{
		"activityId": &graphql.ArgumentConfig{
			Type:        &graphql.NonNull{OfType: graphql.Int},
			Description: "ID of the member memberActivity",
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {

		// Always extract the member id from the token, available thus:
		token := p.Info.VariableValues["token"]
		at, err := jwt.Check(token.(string))
		if err != nil {
			return nil, err
		}
		memberID := at.Claims.ID

		activityID, ok := p.Args["activityId"].(int)
		if ok {
			return data.GetMemberActivity(memberID, int(activityID))
		}

		return nil, nil
	},
}

// memberActivities field fetches multiple memberActivities belonging to a member
var memberActivities = &graphql.Field{
	Description: "Fetches a list of member memberActivities",
	Type:        graphql.NewList(memberActivityType),
	Args: graphql.FieldConfigArgument{
		"last": &graphql.ArgumentConfig{
			Type:        graphql.Int,
			Description: "Fetch only the last (most recent) n records.",
		},
		"from": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "Fetch activities from this date - format 'YYYY-MM-DD'",
		},
		"to": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "Fetch activities up to and including this date - format 'YYYY-MM-DD'",
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {

		// Extract member id from the token, available thus:
		token := p.Info.VariableValues["token"]
		at, err := jwt.Check(token.(string))
		if err != nil {
			return nil, err
		}
		memberID := at.Claims.ID

		// Filter arguments
		f := make(map[string]interface{})
		last, ok := p.Args["last"].(int)
		if ok {
			f["last"] = last
		}
		from, ok := p.Args["from"].(string)
		if ok {
			f["from"], err = utility.DateStringToTime(from)
			if err != nil {
				return nil, err
			}
		}
		to, ok := p.Args["to"].(string)
		if ok {
			f["to"], err = utility.DateStringToTime(to)
			if err != nil {
				return nil, err
			}
		}

		return data.GetMemberActivities(memberID, f)
	},
}

// memberCurrentEvaluation field fetches the current evaluation period data
var memberCurrentEvaluation = &graphql.Field{
	Description: "Fetches activity data for the current evaluation period",
	Type:        memberEvaluationType,
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {

		// Extract member id from the token, available thus:
		token := p.Info.VariableValues["token"]
		at, err := jwt.Check(token.(string))
		if err != nil {
			return nil, err
		}
		memberID := at.Claims.ID

		return data.GetCurrentEvaluation(memberID)
	},
}

// memberEvaluations field fetches a list of activity evaluations for the member
var memberEvaluations = &graphql.Field{
	Description: "Fetches a history of member activity evaluation periods",
	Type:        graphql.NewList(memberEvaluationType),
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {

		// Extract member id from the token, available thus:
		token := p.Info.VariableValues["token"]
		at, err := jwt.Check(token.(string))
		if err != nil {
			return nil, err
		}
		memberID := at.Claims.ID

		return data.GetMemberEvaluations(memberID)
	},
}