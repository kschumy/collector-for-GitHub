package querystrings

import (
	"fmt"
	"strings"
	"time"

	"github.com/collector-for-github/pkg/github-query/types"
)

// TODO: what is sting when no labels/terms?

type QueryStrings struct {
	base         string
	terms        string
	labels       string
	relativeTime string
	querySchema  string
}

type RequestProvider interface {
	GetTerms() []string
	GetLabels() []string
	GetObjectType() types.ObjectType
	GetSearchIn() types.SearchIn
	GetState() types.State
	GetOwnerLogin() string
	GetRepoName() string
	GetAccessible() types.Accessible
	GetRelativeTime() types.RelativeTime
	GetLabelAtIndex(int) (string, error)
}

// InitializeToDefault returns an instance of QueryStrings, which is defaulted to separate terms by
// "OR", and it does not set labels.
func InitializeToDefault(input *RequestProvider) (*QueryStrings, error) {
	qs := QueryStrings{}
	base, err := getBase(input)
	if err != nil {
		return nil, fmt.Errorf("error initializing QueryStrings: %s", err)
	}
	qs.base = base
	qs.querySchema = querySchemaToString((*input).GetObjectType())
	qs.relativeTime = relativeTimeDateToString((*input).GetRelativeTime())
	qs.SetTermsWithOr(*input)
	return &qs, nil
}

func (qs *QueryStrings) GetString() (string, error) {
	if !qs.isInitialized() {
		return "", fmt.Errorf("cannot get string on uninitilized QueryStrings")
	}
	// TODO: move first string out when allow for n results
	return fmt.Sprintf(
		"{\"query\":\"query{search(first: 100, query:\\\"%v\\\", type: ISSUE) %s}\"}",
		strings.Join([]string{qs.base, qs.terms, qs.labels, qs.relativeTime}, " "),
		qs.querySchema,
	), nil
}

// TODO: test this!! There might be an extra space being added here
func (qs *QueryStrings) BuildQueryStringFactory(input *RequestProvider) (func(time.Time) (string, error), error) {
	if !qs.isInitialized() {
		return nil, fmt.Errorf("cannot get with blank date on uninitilized QueryStrings")
	}
	relativeTime := (*input).GetRelativeTime()

	return func(newDate time.Time) (string, error) {
		err := relativeTime.SetTime(newDate)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			"{\"query\":\"query{search(first: 100, query:\\\"%v %s\\\", type: ISSUE) %s}\"}",
			strings.Join([]string{qs.base, qs.terms, qs.labels}, " "),
			relativeTimeDateToString(relativeTime),
			qs.querySchema), nil
	}, nil

}

func (queryStr *QueryStrings) isInitialized() bool {
	return queryStr.base != "" && queryStr.querySchema != "" && queryStr.relativeTime != ""
}
