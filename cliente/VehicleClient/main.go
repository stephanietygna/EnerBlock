package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"fieldclimate/modules"

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
	chaincodeName := "fieldclimate"

	// id da estação para se conectar
	fmt.Print("Insira o ID da estação (ex: 00206C61): ")
	var stationID string
	fmt.Scanln(&stationID)
	log.Info("Estação buscada: ", stationID)

	// leitura da data do json pré conexão com a api (para comparar com os dados lidos após a conexão)
	oldDeviceDate, _ := modules.ReadDate(stationID)

	/* conecta-se a API buscando a estação desejada e insere seus dados em um json */
	//modules.APIConnect(stationID)

	fmt.Print("Insira o ID do dispositivo da estação (ex: HC Air temperature): ")
	reader := bufio.NewReader(os.Stdin)
	stationDevice, _ := reader.ReadString('\n')
	stationDevice = strings.Replace(stationDevice, "\n", "", -1)
	log.Info("Dispositivo buscado: ", stationDevice)

	// lê os dados do dispositivo no json
	deviceName, deviceValues, deviceUnit, deviceDate := modules.JSONRead(stationID, stationDevice)
	fmt.Println("Dados lidos do json: ", deviceName, deviceValues, deviceUnit, deviceDate)

	var resposta string

	if oldDeviceDate == deviceDate {
		log.Info("Dados repetidos encontrados")
		fmt.Println("Os dados da estação " + stationID + " não foram atualizados")
		fmt.Println("Última atualização: " + deviceDate)
		fmt.Println("Deseja continuar? (s/n)")
		fmt.Scanln(&resposta)

		if resposta == "n" {
			log.Info("Encerrando...")
			log.Exit(0)
		} else if resposta == "s" {
			log.Info("Prosseguindo execução com dados repetidos")
			fmt.Println("Continuando...")
		}
	}

	/* CONVERSÃO DE DATAS EM UNIX */
	// Obtém a data e hora atual e converte em unix
	dataAtual := time.Now()
	clientExecutionUnix := dataAtual.Unix()

	// obtém a data e hora dos dados da api e converte em unix
	layout := "2006-01-02 15:04:05"

	parsedDeviceDate, err := time.Parse(layout, deviceDate)
	if err != nil {
		fmt.Println("Erro ao analisar a data:", err)
		return
	}
	deviceDateUnix := parsedDeviceDate.Unix()

	// passando valores para string
	deviceValuesString := ""
	for key, values := range deviceValues {
		deviceValuesString += key + ": " + fmt.Sprint(values) + " "
	}

	deviceDateUnixStr := strconv.FormatInt(deviceDateUnix, 10) //lastupdateunix
	_ = deviceDateUnixStr
	clientExecutionUnixStr := strconv.FormatInt(deviceDateUnix, 10)
	_ = clientExecutionUnixStr

	fmt.Println("Nome do dispositivo: ", deviceName)
	fmt.Println("Dados enviados por ele: ", deviceValuesString)
	fmt.Println("Unidade de medição: ", deviceUnit)
	fmt.Println("Horário em que os dados foram atualizados na API: ", deviceDate)
	fmt.Println("Horário de inserção dos dados na API em Unix: ", deviceDateUnix)
	fmt.Println("Horário de execução do cliente: ", dataAtual)
	fmt.Println("Horário de execução do cliente em unix: ", clientExecutionUnix)

	enrollID := randomString(10)
	registerEnrollUser(configFilePath, enrollID, mspID)

	/* O invoke pode ser feito com o gateway/gw (recomendado) ou sem */
	// Registro de Estação

	modules.InvokeCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "RegisterStation", []string{
		stationID,
		"Victor Daniel",
		clientExecutionUnixStr, // register date
		"10",                   // credits
	})

	// modules.InvokeCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "CreateWallet", []string{
	// 	stationID,
	// 	"Victor Daniel",
	// })

	// modules.QueryCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "QueryWallet", []string{"00206C61", "Victor Daniel"})
	modules.QueryCCgw(configFilePath, channelName, enrollID, mspID, chaincodeName, "QueryStation", []string{"00206C61"})

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
