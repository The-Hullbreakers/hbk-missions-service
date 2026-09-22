package dynamodb

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func getString(item map[string]types.AttributeValue, key string) string {
	v, ok := item[key].(*types.AttributeValueMemberS)
	if !ok {
		return ""
	}
	return v.Value
}

func getInt(item map[string]types.AttributeValue, key string) int {
	v, ok := item[key].(*types.AttributeValueMemberN)
	if !ok {
		return 0
	}
	n, _ := strconv.Atoi(v.Value)
	return n
}

func getStringSlice(item map[string]types.AttributeValue, key string) []string {
	v, ok := item[key].(*types.AttributeValueMemberL)
	if !ok {
		return nil
	}

	var values []string
	for _, attr := range v.Value {
		if s, ok := attr.(*types.AttributeValueMemberS); ok {
			values = append(values, s.Value)
		}
	}
	return values
}

func stringSliceToAttribute(values []string) types.AttributeValue {
	members := make([]types.AttributeValue, 0, len(values))
	for _, v := range values {
		members = append(members, &types.AttributeValueMemberS{Value: v})
	}
	return &types.AttributeValueMemberL{Value: members}
}

func boolToAttribute(value bool) types.AttributeValue {
	return &types.AttributeValueMemberBOOL{Value: value}
}

func getBool(item map[string]types.AttributeValue, key string) bool {
	v, ok := item[key].(*types.AttributeValueMemberBOOL)
	if !ok {
		return false
	}
	return v.Value
}

func timeToAttribute(value time.Time) types.AttributeValue {
	return &types.AttributeValueMemberS{Value: value.UTC().Format(time.RFC3339Nano)}
}

func getTime(item map[string]types.AttributeValue, key string) time.Time {
	raw := getString(item, key)
	if raw == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}
