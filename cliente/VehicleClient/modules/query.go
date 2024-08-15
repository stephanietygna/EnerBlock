package modules

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	log "github.com/sirupsen/logrus"
)

func QueryCCgw(configFilePath, channelName, userName, mspID, chaincodeName, fcn string, args []string) {

	configBackend := config.FromFile(configFilePath)
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		log.Error(err)
	}

	wallet, err := gateway.NewFileSystemWallet(fmt.Sprintf("wallet/%s", mspID))

	gw, err := gateway.Connect(
		gateway.WithSDK(sdk),
		gateway.WithUser(userName),
		gateway.WithIdentity(wallet, userName),
	)

	if err != nil {
		log.Error("Failed to create new Gateway: %s", err)
	}
	defer gw.Close()
	nw, err := gw.GetNetwork(channelName)
	if err != nil {
		log.Error("Failed to get network: %s", err)
	}

	contract := nw.GetContract(chaincodeName)

	resp, err := contract.EvaluateTransaction(fcn, args...)

	if err != nil {
		log.Error("Failed submit transaction: %s", err)
	}
	log.Info(string(resp))
	fmt.Println(string(resp))

}

func queryCC(configFilePath, channelName, userName, mspID, chaincodeName, fcn string) {
	userName = "admin"

	configBackend := config.FromFile(configFilePath)
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		log.Error(err)
	}
	log.Println(sdk)
	chContext := sdk.ChannelContext(
		channelName,
		fabsdk.WithUser(userName),
		fabsdk.WithOrg(mspID),
	)

	ch, err := channel.New(chContext)
	if err != nil {
		log.Error(err)
	}

	response, err := ch.Query(
		channel.Request{
			ChaincodeID:     chaincodeName,
			Fcn:             fcn,
			Args:            nil,
			TransientMap:    nil,
			InvocationChain: nil,
			IsInit:          false,
		},
	)

	if err != nil {
		log.Error(err)
	}
	log.Infof("response=%s", response.Payload)
}