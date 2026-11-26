package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

var ErrPlacementManifestConflict = errors.New("placement manifest progress conflicts with stored state")

func manifestFromItem(item map[string]types.AttributeValue) (application.PlacementManifest, error) {
	count, countErr := manifestNumber(item, "FactCount")
	index, indexErr := manifestNumber(item, "NextIndex")
	cutoff, timeErr := time.Parse(time.RFC3339Nano, getString(item, "Cutoff"))
	season, attempt := getString(item, "SeasonID"), getString(item, "AttemptID")
	id, hash := getString(item, "SK"), getString(item, "FactHash")
	if countErr != nil || indexErr != nil || timeErr != nil || count < 0 || index < 0 || index > count ||
		season == "" || attempt == "" || id != "MANIFEST#"+season+"#"+attempt || len(hash) != 64 {
		return application.PlacementManifest{}, fmt.Errorf("invalid stored placement manifest")
	}
	return application.PlacementManifest{
		ID: id, SeasonID: season, AttemptID: attempt, Cutoff: cutoff,
		FactCount: count, FactHash: hash, NextIndex: index,
	}, nil
}

func (s *SeasonStore) storedManifest(ctx context.Context, id string) (application.PlacementManifest, string, error) {
	item, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: outboxPartition},
			"SK": &types.AttributeValueMemberS{Value: id},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return application.PlacementManifest{}, "", fmt.Errorf("failed to read placement manifest: %w", err)
	}
	if len(item.Item) == 0 {
		return application.PlacementManifest{}, "", ErrPlacementManifestConflict
	}
	manifest, err := manifestFromItem(item.Item)
	if err != nil {
		return application.PlacementManifest{}, "", err
	}
	return manifest, getString(item.Item, "Status"), nil
}

// MarkFactPublished advances only the fact at the committed cursor.
// A repeated mark of the last acknowledged fact is a no-op.
func (s *SeasonStore) MarkFactPublished(ctx context.Context, id, eventID string) error {
	manifest, status, err := s.storedManifest(ctx, id)
	if err != nil {
		return err
	}
	if manifest.NextIndex > 0 {
		previous := manifest
		previous.NextIndex--
		remaining, readErr := s.ListRemaining(ctx, previous)
		if readErr != nil {
			return readErr
		}
		if remaining[0].EventID == eventID {
			return nil
		}
	}
	if status != outboxStatusPending {
		return ErrPlacementManifestConflict
	}
	remaining, err := s.ListRemaining(ctx, manifest)
	if err != nil {
		return err
	}
	if len(remaining) == 0 || remaining[0].EventID != eventID {
		return ErrPlacementManifestConflict
	}
	_, err = s.client.Client().UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: outboxPartition},
			"SK": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:         aws.String("SET NextIndex = :next"),
		ConditionExpression:      aws.String("#status = :pending AND NextIndex = :old AND FactCount >= :next"),
		ExpressionAttributeNames: map[string]string{"#status": "Status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pending": &types.AttributeValueMemberS{Value: outboxStatusPending},
			":old":     &types.AttributeValueMemberN{Value: fmt.Sprint(manifest.NextIndex)},
			":next":    &types.AttributeValueMemberN{Value: fmt.Sprint(manifest.NextIndex + 1)},
		},
	})
	if err == nil {
		return nil
	}
	var conflict *types.ConditionalCheckFailedException
	if errors.As(err, &conflict) {
		return ErrPlacementManifestConflict
	}
	return fmt.Errorf("failed to mark placement fact published: %w", err)
}

// MarkComplete closes a manifest only after every staged fact has been marked.
func (s *SeasonStore) MarkComplete(ctx context.Context, id string) error {
	manifest, status, err := s.storedManifest(ctx, id)
	if err != nil {
		return err
	}
	if status == outboxStatusPublished && manifest.NextIndex == manifest.FactCount {
		return nil
	}
	if status != outboxStatusPending || manifest.NextIndex != manifest.FactCount {
		return ErrPlacementManifestConflict
	}
	if _, err := s.ListRemaining(ctx, manifest); err != nil {
		return err
	}
	_, err = s.client.Client().UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: outboxPartition},
			"SK": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:         aws.String("SET #status = :published"),
		ConditionExpression:      aws.String("#status = :pending AND NextIndex = :count AND FactCount = :count"),
		ExpressionAttributeNames: map[string]string{"#status": "Status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pending":   &types.AttributeValueMemberS{Value: outboxStatusPending},
			":published": &types.AttributeValueMemberS{Value: outboxStatusPublished},
			":count":     &types.AttributeValueMemberN{Value: fmt.Sprint(manifest.FactCount)},
		},
	})
	if err == nil {
		return nil
	}
	var conflict *types.ConditionalCheckFailedException
	if errors.As(err, &conflict) {
		return ErrPlacementManifestConflict
	}
	return fmt.Errorf("failed to complete placement manifest: %w", err)
}
