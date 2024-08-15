package modules

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

// função para criar HMAC, necessário para autenticação no FieldClimate
func createHMAC(method, request, publicKey, privateKey string) string {
	timestamp := time.Now().UTC().Format(time.RFC1123)
	msg := []byte(method + request + timestamp + publicKey)
	key := []byte(privateKey)

	hash := hmac.New(sha256.New, key)
	hash.Write(msg)
	signature := hex.EncodeToString(hash.Sum(nil))

	hmacStr := fmt.Sprintf("hmac %s:%s", publicKey, signature)
	return hmacStr
}

func APIConnect(stationid string) {
	PUBLIC_KEY, PRIVATE_KEY := ReadKeys()
	apiURI := "https://api.fieldclimate.com/v2"
	request := "/data/" + stationid + "/raw/last/1" // mude a rota aqui
	method := "POST"                                // mude o método aqui
	url := apiURI + request
	hmacStr := createHMAC(method, request, PUBLIC_KEY, PRIVATE_KEY)

	// iniciando request para a api
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// adicionando cabeçalhos no request
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Date", time.Now().UTC().Format(time.RFC1123))
	req.Header.Add("Authorization", hmacStr)

	// criando cliente http para enviar o request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// verificar status (deve ser "200 OK")
	if resp.Status != "200 OK" {
		fmt.Println("Error")
		log.Fatal(resp.Status)
	} else {
		fmt.Println("Autenticação realizada com sucesso")
		fmt.Println("Status code: " + resp.Status)
	}

	// encerrar resp
	defer resp.Body.Close()

	// ler corpo da resposta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// salva resposta em um arquivo JSON com o nome da estação
	file, _ := PrettyString(string(body))
	os.WriteFile("json/"+stationid+".json", []byte(file), 0644)
	fmt.Println("Dados da estação salvos em: json/" + stationid + ".json")

	//fmt.Println(PrettyString(string(body)))
}

// função para formatar a resposta JSON da API
func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "   "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
