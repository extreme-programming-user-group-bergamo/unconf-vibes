package rooms

import "github.com/katurdays/unconf/internal/client"

// RoomsLoadedMsg is emitted after room data is loaded successfully.
type RoomsLoadedMsg struct {
	Rooms []client.RoomResponse
}

// RoomsLoadErrMsg is emitted when room data loading fails.
type RoomsLoadErrMsg struct {
	Err error
}
