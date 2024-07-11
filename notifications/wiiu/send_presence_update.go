package notifications_wiiu

import (
	"fmt"

	"github.com/CloudnetworkTeam/friends/database"
	database_wiiu "github.com/CloudnetworkTeam/friends/database/wiiu"
	"github.com/CloudnetworkTeam/friends/globals"
	nex "github.com/PretendoNetwork/nex-go"
	friends_wiiu_types "github.com/PretendoNetwork/nex-protocols-go/friends-wiiu/types"
	nintendo_notifications "github.com/PretendoNetwork/nex-protocols-go/nintendo-notifications"
	nintendo_notifications_types "github.com/PretendoNetwork/nex-protocols-go/nintendo-notifications/types"
)

func SendPresenceUpdate(presence *friends_wiiu_types.NintendoPresenceV2) {
	eventObject := nintendo_notifications_types.NewNintendoNotificationEvent()
	eventObject.Type = 24
	eventObject.SenderPID = presence.PID
	eventObject.DataHolder = nex.NewDataHolder()
	eventObject.DataHolder.SetTypeName("NintendoPresenceV2")
	eventObject.DataHolder.SetObjectData(presence)

	stream := nex.NewStreamOut(globals.SecureServer)
	eventObjectBytes := eventObject.Bytes(stream)

	rmcRequest := nex.NewRMCRequest()
	rmcRequest.SetProtocolID(nintendo_notifications.ProtocolID)
	rmcRequest.SetCallID(3810693103)
	rmcRequest.SetMethodID(nintendo_notifications.MethodProcessNintendoNotificationEvent2)
	rmcRequest.SetParameters(eventObjectBytes)

	rmcRequestBytes := rmcRequest.Bytes()

	friendList, err := database_wiiu.GetUserFriendList(presence.PID)
	if err != nil && err != database.ErrEmptyList {
		globals.Logger.Critical(err.Error())
	}

	for i := 0; i < len(friendList); i++ {
		if friendList[i] == nil || friendList[i].NNAInfo == nil || friendList[i].NNAInfo.PrincipalBasicInfo == nil {
			// TODO: Fix this
			pid := presence.PID
			var friendPID uint32 = 0

			if friendList[i] != nil && friendList[i].Presence != nil {
				// TODO: Better track the bad users PID
				friendPID = friendList[i].Presence.PID
			}

			globals.Logger.Error(fmt.Sprintf("User %d has friend %d with bad presence data", pid, friendPID))

			if friendList[i] == nil {
				globals.Logger.Error(fmt.Sprintf("%d friendList[i] nil", friendPID))
			} else if friendList[i].NNAInfo == nil {
				globals.Logger.Error(fmt.Sprintf("%d friendList[i].NNAInfo is nil", friendPID))
			} else if friendList[i].NNAInfo.PrincipalBasicInfo == nil {
				globals.Logger.Error(fmt.Sprintf("%d friendList[i].NNAInfo.PrincipalBasicInfo is nil", friendPID))
			}

			continue
		}

		friendPID := friendList[i].NNAInfo.PrincipalBasicInfo.PID
		connectedUser := globals.ConnectedUsers[friendPID]

		if connectedUser != nil {
			requestPacket, _ := nex.NewPacketV0(connectedUser.Client, nil)

			requestPacket.SetVersion(0)
			requestPacket.SetSource(0xA1)
			requestPacket.SetDestination(0xAF)
			requestPacket.SetType(nex.DataPacket)
			requestPacket.SetPayload(rmcRequestBytes)

			requestPacket.AddFlag(nex.FlagNeedsAck)
			requestPacket.AddFlag(nex.FlagReliable)

			globals.SecureServer.Send(requestPacket)
		}
	}
}
