package database_wiiu

import (
	"database/sql"

	"github.com/PretendoNetwork/friends/database"
	"github.com/PretendoNetwork/nex-go/v2/types"
	friends_wiiu_types "github.com/PretendoNetwork/nex-protocols-go/v2/friends-wiiu/types"
)

// GetUserPrincipalPreference returns the user preferences
func GetUserPrincipalPreference(pid uint32) (friends_wiiu_types.PrincipalPreference, error) {
	preference := friends_wiiu_types.NewPrincipalPreference()

	var showOnlinePresence bool
	var showCurrentTitle bool
	var blockFriendRequests bool

	row, err := database.Manager.QueryRow(`SELECT show_online, show_current_game, block_friend_requests FROM wiiu.user_data WHERE pid=$1`, pid)
	if err != nil {
		return preference, err
	}

	err = row.Scan(&showOnlinePresence, &showCurrentTitle, &blockFriendRequests)
	if err != nil {
		if err == sql.ErrNoRows {
			return preference, database.ErrPIDNotFound
		} else {
			return preference, err
		}
	}

	preference.ShowOnlinePresence = types.NewBool(showOnlinePresence)
	preference.ShowCurrentTitle = types.NewBool(showCurrentTitle)
	preference.BlockFriendRequests = types.NewBool(blockFriendRequests)

	return preference, nil
}
