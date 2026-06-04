package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"villum/ent"
	"villum/ent/aiendpoint"
	"villum/models"
)

func entAIEndpointToModel(e *ent.AIEndpoint) models.AIEndpoint {
	m := models.AIEndpoint{
		ID:              e.ID,
		Name:            e.Name,
		Type:            string(e.Type),
		BaseURL:         e.BaseURL,
		EncryptedAPIKey: e.EncryptedAPIKey,
		Model:           e.Model,
		Tags:            e.Tags,
		Enabled:         e.Enabled,
		Temperature:     e.Temperature,
		MaxTokens:       e.MaxTokens,
		ImageSize:       e.ImageSize,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
	return m
}

func GetAIEndpoints(ctx context.Context) ([]models.AIEndpoint, error) {
	ents, err := Client.AIEndpoint.Query().Order(ent.Asc(aiendpoint.FieldName)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list ai endpoints: %w", err)
	}
	out := make([]models.AIEndpoint, len(ents))
	for i, e := range ents {
		out[i] = entAIEndpointToModel(e)
	}
	return out, nil
}

func GetAIEndpoint(ctx context.Context, id int64) (*models.AIEndpoint, error) {
	e, err := Client.AIEndpoint.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ai endpoint %d: %w", id, err)
	}
	m := entAIEndpointToModel(e)
	return &m, nil
}

func CreateAIEndpoint(ctx context.Context, m *models.AIEndpoint) (*models.AIEndpoint, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	create := Client.AIEndpoint.Create().
		SetName(m.Name).
		SetType(aiendpoint.Type(m.Type)).
		SetBaseURL(m.BaseURL).
		SetEncryptedAPIKey(m.EncryptedAPIKey).
		SetModel(m.Model).
		SetEnabled(m.Enabled).
		SetCreatedAt(now).
		SetUpdatedAt(now)

	if len(m.Tags) > 0 {
		create.SetTags(m.Tags)
	}
	if m.Temperature != nil {
		create.SetTemperature(*m.Temperature)
	}
	if m.MaxTokens != nil {
		create.SetMaxTokens(*m.MaxTokens)
	}
	if m.ImageSize != nil {
		create.SetImageSize(*m.ImageSize)
	}

	e, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ai endpoint: %w", err)
	}
	m2 := entAIEndpointToModel(e)
	return &m2, nil
}

func UpdateAIEndpoint(ctx context.Context, id int64, m *models.AIEndpoint) (*models.AIEndpoint, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	update := Client.AIEndpoint.UpdateOneID(id).
		SetName(m.Name).
		SetType(aiendpoint.Type(m.Type)).
		SetBaseURL(m.BaseURL).
		SetModel(m.Model).
		SetEnabled(m.Enabled).
		SetUpdatedAt(now)

	if m.Tags != nil {
		update.SetTags(m.Tags)
	}

	// Only update API key if a new one was provided
	if m.EncryptedAPIKey != "" {
		update.SetEncryptedAPIKey(m.EncryptedAPIKey)
	}

	if m.Temperature != nil {
		update.SetTemperature(*m.Temperature)
	} else {
		update.ClearTemperature()
	}
	if m.MaxTokens != nil {
		update.SetMaxTokens(*m.MaxTokens)
	} else {
		update.ClearMaxTokens()
	}
	if m.ImageSize != nil {
		update.SetImageSize(*m.ImageSize)
	} else {
		update.ClearImageSize()
	}

	e, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update ai endpoint %d: %w", id, err)
	}
	m2 := entAIEndpointToModel(e)
	return &m2, nil
}

func DeleteAIEndpoint(ctx context.Context, id int64) error {
	err := Client.AIEndpoint.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete ai endpoint %d: %w", id, err)
	}
	return nil
}

func GetEnabledAIEndpointsByType(ctx context.Context, endpointType string) ([]models.AIEndpoint, error) {
	ents, err := Client.AIEndpoint.Query().
		Where(aiendpoint.Enabled(true), aiendpoint.TypeEQ(aiendpoint.Type(endpointType))).
		Order(ent.Asc(aiendpoint.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled %s ai endpoints: %w", endpointType, err)
	}
	out := make([]models.AIEndpoint, len(ents))
	for i, e := range ents {
		m := entAIEndpointToModel(e)
		m.EncryptedAPIKey = "" // don't expose key to user
		out[i] = m
	}
	return out, nil
}

// AIEndpointInfo is a lightweight view of an AI endpoint safe for DM exposure.
type AIEndpointInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Model string `json:"model"`
	Type  string `json:"type"`
}

// ListEnabledAIEndpoints returns only id, name, model, type for enabled endpoints
// of the given type — safe for non-admin users (no API key or base URL exposed).
func ListEnabledAIEndpoints(ctx context.Context, endpointType string) ([]AIEndpointInfo, error) {
	ents, err := Client.AIEndpoint.Query().
		Where(aiendpoint.Enabled(true), aiendpoint.TypeEQ(aiendpoint.Type(endpointType))).
		Order(ent.Asc(aiendpoint.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled %s ai endpoints: %w", endpointType, err)
	}
	out := make([]AIEndpointInfo, len(ents))
	for i, e := range ents {
		out[i] = AIEndpointInfo{ID: e.ID, Name: e.Name, Model: e.Model, Type: string(e.Type)}
	}
	return out, nil
}

func CheckAIEndpointNameUnique(ctx context.Context, name string, excludeID int64) (bool, error) {
	q := Client.AIEndpoint.Query().Where(aiendpoint.Name(name))
	if excludeID != 0 {
		q = q.Where(aiendpoint.IDNEQ(excludeID))
	}
	count, err := q.Count(ctx)
	if err != nil {
		return false, fmt.Errorf("check name uniqueness: %w", err)
	}
	return count == 0, nil
}

func InitAIEndpointsTable(ctx context.Context) error {
	log.Printf("AI endpoints table managed by ent migration")
	return nil
}
