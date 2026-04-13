package sqlite

import (
	"context"
	"testing"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoom(conferenceID int64, roomNumber string) *models.Room {
	return &models.Room{
		ConferenceID:  conferenceID,
		RoomNumber:    roomNumber,
		RoomType:      "double",
		PricePerNight: 150.00,
		Capacity:      2,
	}
}

func TestRoomRepository_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	roomRepo := NewRoomRepository(db)

	conf, err := confRepo.Create(context.Background(), newTestConference("room-test"))
	require.NoError(t, err)

	room := newTestRoom(conf.ID, "101")
	created, err := roomRepo.Create(context.Background(), room)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, conf.ID, created.ConferenceID)
	assert.Equal(t, "101", created.RoomNumber)
	assert.Equal(t, "double", created.RoomType)
	assert.Equal(t, 150.00, created.PricePerNight)
	assert.Equal(t, 2, created.Capacity)
	assert.False(t, created.CreatedAt.IsZero())
}

func TestRoomRepository_Create_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	roomRepo := NewRoomRepository(db)

	conf, err := confRepo.Create(context.Background(), newTestConference("dup-room"))
	require.NoError(t, err)

	_, err = roomRepo.Create(context.Background(), newTestRoom(conf.ID, "101"))
	require.NoError(t, err)

	_, err = roomRepo.Create(context.Background(), newTestRoom(conf.ID, "101"))
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrRoomExists)
}

func TestRoomRepository_GetByID_Success(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	roomRepo := NewRoomRepository(db)

	conf, err := confRepo.Create(context.Background(), newTestConference("get-room"))
	require.NoError(t, err)

	created, err := roomRepo.Create(context.Background(), newTestRoom(conf.ID, "201"))
	require.NoError(t, err)

	fetched, err := roomRepo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "201", fetched.RoomNumber)
	assert.Equal(t, "double", fetched.RoomType)
}

func TestRoomRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	roomRepo := NewRoomRepository(db)

	_, err := roomRepo.GetByID(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrRoomNotFound)
}

func TestRoomRepository_ListByConference_ReturnsAll(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	roomRepo := NewRoomRepository(db)

	conf, err := confRepo.Create(context.Background(), newTestConference("list-rooms"))
	require.NoError(t, err)

	_, err = roomRepo.Create(context.Background(), newTestRoom(conf.ID, "102"))
	require.NoError(t, err)
	_, err = roomRepo.Create(context.Background(), newTestRoom(conf.ID, "101"))
	require.NoError(t, err)

	rooms, err := roomRepo.ListByConference(context.Background(), conf.ID)
	require.NoError(t, err)
	require.Len(t, rooms, 2)
	// Ordered by room_number
	assert.Equal(t, "101", rooms[0].RoomNumber)
	assert.Equal(t, "102", rooms[1].RoomNumber)
}

func TestRoomRepository_ListByConference_Empty(t *testing.T) {
	db := setupTestDB(t)
	roomRepo := NewRoomRepository(db)

	rooms, err := roomRepo.ListByConference(context.Background(), 999)
	require.NoError(t, err)
	require.NotNil(t, rooms)
	assert.Empty(t, rooms)
}

func TestRoomRepository_ListByConference_OnlySpecifiedConference(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	roomRepo := NewRoomRepository(db)

	conf1, err := confRepo.Create(context.Background(), newTestConference("conf-a"))
	require.NoError(t, err)
	conf2, err := confRepo.Create(context.Background(), newTestConference("conf-b"))
	require.NoError(t, err)

	_, err = roomRepo.Create(context.Background(), newTestRoom(conf1.ID, "101"))
	require.NoError(t, err)
	_, err = roomRepo.Create(context.Background(), newTestRoom(conf2.ID, "201"))
	require.NoError(t, err)

	rooms, err := roomRepo.ListByConference(context.Background(), conf1.ID)
	require.NoError(t, err)
	require.Len(t, rooms, 1)
	assert.Equal(t, "101", rooms[0].RoomNumber)
}
