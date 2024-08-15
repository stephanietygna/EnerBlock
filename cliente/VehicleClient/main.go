package main

import (
	"fmt"
	"math/rand"
	"os"

	"time"

	"vehiclecontract/modules"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func main() {
	// inicializando o log
	file, err := os.OpenFile("logs/log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Info("Iniciando cliente...")
	fmt.Println("Em caso de erro, verifique o arquivo log")

	//configFilePath := os.Args[1]
	configFilePath := "connection-org.yaml"
	channelName := "demo"
	mspID := "INMETROMSP"
	chaincodeName := "VehicleContract"

	enrollID := "admin"
	registerEnrollUser(configFilePath, enrollID, mspID)

	/* O invoke pode ser feito com o gateway/gw (recomendado) ou sem */
	// Registro de Estação

	modules.InvokeCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "Createuser", []string{
		"123456",
		"John Doe",
		"tanq",
		"Truck",
	})

	// modules.InvokeCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "CreateWallet", []string{
	// 	stationID,
	// 	"Victor Daniel",
	// })

	// modules.QueryCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "QueryWallet", []string{"00206C61", "Victor Daniel"})
	modules.QueryCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "Userget", []string{"123456"})
}

func registerEnrollUser(configFilePath, enrollID, mspID string) {
	log.Info("Registering User : ", enrollID)
	sdk, err := fabsdk.New(config.FromFile(configFilePath))

	if err != nil {
		log.Errorf("failed to create new SDK: %s\n", err)
	}

	ctx := sdk.Context()
	caClient, err := mspclient.New(
		ctx,
		mspclient.WithCAInstance("inmetro-ca.default"),
		mspclient.WithOrg(mspID),
	)

	if err != nil {
		log.Errorf("failed to create msp client: %s\n", err)
	}

	if caClient != nil {
		log.Info("ca client created")
	}
	enrollmentSecret, err := caClient.Register(&mspclient.RegistrationRequest{
		Name:           enrollID,
		Type:           "client",
		MaxEnrollments: -1,
		Affiliation:    "",
		Attributes:     nil,
		Secret:         enrollID,
	})
	if err != nil {
		log.Error(err)
	}
	err = caClient.Enroll(
		enrollID,
		mspclient.WithSecret(enrollmentSecret),
		mspclient.WithProfile("tls"),
	)
	if err != nil {
		log.Error(errors.WithMessage(err, "failed to register identity"))
	}

	wallet, err := gateway.NewFileSystemWallet(fmt.Sprintf("wallet/%s", mspID))
	if err != nil {
		log.Error(err)
	}

	signingIdentity, err := caClient.GetSigningIdentity(enrollID)
	if err != nil {
		log.Error(err)
	}

	key, err := signingIdentity.PrivateKey().Bytes()
	if err != nil {
		log.Error(err)
	}

	identity := gateway.NewX509Identity(mspID, string(signingIdentity.EnrollmentCertificate()), string(key))

	err = wallet.Put(enrollID, identity)
	if err != nil {
		log.Error(err)
	}

}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
