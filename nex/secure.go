package nex

import (
	"fmt"
	"os"
	"time"

	database_3ds "github.com/CloudnetworkTeam/friends/database/3ds"
	database_wiiu "github.com/CloudnetworkTeam/friends/database/wiiu"
	"github.com/CloudnetworkTeam/friends/globals"
	notifications_3ds "github.com/CloudnetworkTeam/friends/notifications/3ds"
	notifications_wiiu "github.com/CloudnetworkTeam/friends/notifications/wiiu"
	"github.com/CloudnetworkTeam/friends/types"
	nex "github.com/PretendoNetwork/nex-go"
	_ "github.com/PretendoNetwork/nex-protocols-go"
)

func StartSecureServer() {
	globals.SecureServer = nex.NewServer()
	globals.SecureServer.SetFragmentSize(900)
	globals.SecureServer.SetPRUDPVersion(0)
	globals.SecureServer.SetKerberosKeySize(16)
	globals.SecureServer.SetKerberosPassword(globals.KerberosPassword)
	globals.SecureServer.SetPingTimeout(20) // Maybe too long?
	globals.SecureServer.SetAccessKey("ridfebb9")
	globals.SecureServer.SetDefaultNEXVersion(&nex.NEXVersion{
		Major: 1,
		Minor: 1,
		Patch: 0,
	})

	globals.SecureServer.On("Data", func(packet *nex.PacketV0) {
		request := packet.RMCRequest()

		fmt.Println("==Friends - Secure==")
		fmt.Printf("Protocol ID: %#v\n", request.ProtocolID())
		fmt.Printf("Method ID: %#v\n", request.MethodID())
		fmt.Println("====================")
	})

	globals.SecureServer.On("Kick", func(packet *nex.PacketV0) {
		pid := packet.Sender().PID()

		if globals.ConnectedUsers[pid] == nil {
			return
		}

		platform := globals.ConnectedUsers[pid].Platform
		lastOnline := nex.NewDateTime(0)
		lastOnline.FromTimestamp(time.Now())

		if platform == types.WUP {
			err := database_wiiu.UpdateUserLastOnlineTime(pid, lastOnline)
			if err != nil {
				globals.Logger.Critical(err.Error())
			}

			notifications_wiiu.SendUserWentOfflineGlobally(packet.Sender())
		} else if platform == types.CTR {
			err := database_3ds.UpdateUserLastOnlineTime(pid, lastOnline)
			if err != nil {
				globals.Logger.Critical(err.Error())
			}

			notifications_3ds.SendUserWentOfflineGlobally(packet.Sender())
		}

		delete(globals.ConnectedUsers, pid)
		fmt.Println("Leaving (Kick)")
	})

	globals.SecureServer.On("Disconnect", func(packet *nex.PacketV0) {
		fmt.Println("Leaving (Disconnect)")
	})

	globals.SecureServer.On("Connect", connect)

	registerCommonSecureServerProtocols()
	registerSecureServerProtocols()

	globals.SecureServer.Listen(fmt.Sprintf(":%s", os.Getenv("PN_FRIENDS_SECURE_SERVER_PORT")))
}
