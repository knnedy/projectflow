package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/repository"
)

const defaultActivityLimit = 50

type ActivityService struct {
	queries *repository.Queries
}

func NewActivityService(db *repository.DB) *ActivityService {
	return &ActivityService{
		queries: db.Queries(),
	}
}

// LogInput is used internally by other services to log activity
type LogInput struct {
	OrgID      string
	ProjectID  string // optional — empty for org level actions
	EntityType repository.ActivityEntityType
	EntityID   string
	Action     repository.ActivityAction
	ActorID    string
	Metadata   any // optional — marshalled to JSON
}

// ActivityPage is the result of a paginated activity log query
type ActivityPage struct {
	Logs       []repository.ActivityLog
	NextCursor *time.Time
	HasMore    bool
}

func (s *ActivityService) Log(ctx context.Context, input LogInput) error {
	// parse org ID
	parsedOrgID, err := uuid.Parse(input.OrgID)
	if err != nil {
		return domain.ErrInternal
	}

	// parse entity ID
	parsedEntityID, err := uuid.Parse(input.EntityID)
	if err != nil {
		return domain.ErrInternal
	}

	// parse actor ID
	parsedActorID, err := uuid.Parse(input.ActorID)
	if err != nil {
		return domain.ErrInternal
	}

	// handle optional project ID
	var projectID pgtype.UUID
	if input.ProjectID != "" {
		parsedProjectID, err := uuid.Parse(input.ProjectID)
		if err != nil {
			return domain.ErrInternal
		}
		projectID = pgtype.UUID{Bytes: parsedProjectID, Valid: true}
	}

	// marshal metadata to JSON if provided
	var metadata []byte
	if input.Metadata != nil {
		metadata, err = json.Marshal(input.Metadata)
		if err != nil {
			return domain.ErrInternal
		}
	}

	// create activity log entry
	_, err = s.queries.CreateActivityLog(ctx, repository.CreateActivityLogParams{
		ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OrganisationID: pgtype.UUID{Bytes: parsedOrgID, Valid: true},
		ProjectID:      projectID,
		EntityType:     input.EntityType,
		EntityID:       pgtype.UUID{Bytes: parsedEntityID, Valid: true},
		Action:         input.Action,
		ActorID:        pgtype.UUID{Bytes: parsedActorID, Valid: true},
		Metadata:       metadata,
	})
	if err != nil {
		return domain.ErrDatabase
	}

	return nil
}

func (s *ActivityService) ListByOrg(ctx context.Context, orgID string, cursor *time.Time, limit int32) (ActivityPage, error) {
	// parse org ID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return ActivityPage{}, domain.ErrNotFound
	}

	// apply default limit if not provided
	if limit <= 0 {
		limit = defaultActivityLimit
	}

	// build cursor param — zero value means first page, fetches all
	var cursorParam pgtype.Timestamp
	if cursor != nil {
		cursorParam = pgtype.Timestamp{Time: *cursor, Valid: true}
	}

	// fetch limit + 1 to detect if there is a next page
	logs, err := s.queries.ListActivityLogsByOrg(ctx, repository.ListActivityLogsByOrgParams{
		OrganisationID: pgtype.UUID{Bytes: parsedOrgID, Valid: true},
		Column2:        cursorParam,
		Limit:          limit + 1,
	})
	if err != nil {
		return ActivityPage{}, domain.ErrDatabase
	}

	return buildActivityPage(logs, limit), nil
}

func (s *ActivityService) ListByProject(ctx context.Context, projectID string, cursor *time.Time, limit int32) (ActivityPage, error) {
	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return ActivityPage{}, domain.ErrNotFound
	}

	// apply default limit if not provided
	if limit <= 0 {
		limit = defaultActivityLimit
	}

	// build cursor param — zero value means first page
	var cursorParam pgtype.Timestamp
	if cursor != nil {
		cursorParam = pgtype.Timestamp{Time: *cursor, Valid: true}
	}

	// fetch limit + 1 to detect if there is a next page
	logs, err := s.queries.ListActivityLogsByProject(ctx, repository.ListActivityLogsByProjectParams{
		ProjectID: pgtype.UUID{Bytes: parsedProjectID, Valid: true},
		Column2:   cursorParam,
		Limit:     limit + 1,
	})
	if err != nil {
		return ActivityPage{}, domain.ErrDatabase
	}

	return buildActivityPage(logs, limit), nil
}

func (s *ActivityService) ListByEntity(ctx context.Context, entityID string, cursor *time.Time, limit int32) (ActivityPage, error) {
	// parse entity ID
	parsedEntityID, err := uuid.Parse(entityID)
	if err != nil {
		return ActivityPage{}, domain.ErrNotFound
	}

	// apply default limit if not provided
	if limit <= 0 {
		limit = defaultActivityLimit
	}

	// build cursor param — zero value means first page
	var cursorParam pgtype.Timestamp
	if cursor != nil {
		cursorParam = pgtype.Timestamp{Time: *cursor, Valid: true}
	}

	// fetch limit + 1 to detect if there is a next page
	logs, err := s.queries.ListActivityLogsByEntity(ctx, repository.ListActivityLogsByEntityParams{
		EntityID: pgtype.UUID{Bytes: parsedEntityID, Valid: true},
		Column2:  cursorParam,
		Limit:    limit + 1,
	})
	if err != nil {
		return ActivityPage{}, domain.ErrDatabase
	}

	return buildActivityPage(logs, limit), nil
}

// buildActivityPage trims the extra record and builds the cursor for the next page
func buildActivityPage(logs []repository.ActivityLog, limit int32) ActivityPage {
	hasMore := int32(len(logs)) > limit

	if hasMore {
		// trim the extra record we fetched to detect more pages
		logs = logs[:limit]
	}

	var nextCursor *time.Time
	if hasMore && len(logs) > 0 {
		// next cursor is the created_at of the last item in this page
		t := logs[len(logs)-1].CreatedAt.Time
		nextCursor = &t
	}

	return ActivityPage{
		Logs:       logs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}
