package types

type SearchComputers struct {
	MaxItems       int              `json:"maxItems"`
	SearchCriteria []SearchCriteria `json:"searchCriteria"`
	SortByObjectID bool             `json:"sortByObjectID"`
}

type SearchCriteria struct {
	FieldName   string `json:"fieldName"`
	StringTest  string `json:"stringTest"`
	StringValue string `json:"stringValue"`
}
