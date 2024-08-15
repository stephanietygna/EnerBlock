package modules

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	log "github.com/sirupsen/logrus"
)

func InvokeCCgw(configFilePath, channelName, userName, mspID, chaincodeName, fcn string, params []string) {

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

	// aqui ele chama a função com os parametros!
	//resp, err := contract.SubmitTransaction(fcn, userName, "a", "b", "1", "ewdscwds")
	resp, err := contract.SubmitTransaction(fcn, params...)

	if err != nil {
		fmt.Println("Um erro ocorreu, verifique o arquivo log para mais informações")
		log.Error("Failed submit transaction: %s", err)
	}
	log.Info("-------------------")
	log.Info("Response: ")
	log.Info(resp)
	log.Info("-------------------")
	fmt.Println(resp)

}

func InvokeCC(configFilePath, channelName, userName, mspID, chaincodeName, fcn string) {

	userName = "admin"

	configBackend := config.FromFile(configFilePath)
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		log.Error(err)
	}

	chContext := sdk.ChannelContext(
		channelName,
		fabsdk.WithUser(userName),
		fabsdk.WithOrg(mspID),
	)

	ch, err := channel.New(chContext)
	if err != nil {
		log.Error(err)
	}

	var args [][]byte

	inputArgs := []string{userName, "23", "234", "2324", "234"}
	for _, arg := range inputArgs {
		args = append(args, []byte(arg))
	}
	response, err := ch.Execute(
		channel.Request{
			ChaincodeID:     chaincodeName,
			Fcn:             fcn,
			Args:            args,
			TransientMap:    nil,
			InvocationChain: nil,
			IsInit:          false,
		},
	)

	if err != nil {
		log.Error(err)
	}

	log.Infof("txid=%s", response.TransactionID)
}
