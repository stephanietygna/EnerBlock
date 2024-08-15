# Documentação sobre clientes

Requisitos: go versão 1.18

Para utilizar os clientes é necessário primeiro levantar a rede e instalar o seu respectivo chaincode/smart contract.

# Configuração inicial

- Através da pasta raiz, obtenha o arquivo com as informações das organizações com os seguintes comandos:

```bash
kubectl hlf inspect -c=demo --output resources/network.yaml -o INMETROMSP -o OrdererMSP

kubectl hlf ca register --name=inmetro-ca --user=admin --secret=adminpw --type=admin \
 --enroll-id enroll --enroll-secret=enrollpw --mspid INMETROMSP  

 kubectl hlf ca enroll --name=inmetro-ca --user=admin --secret=adminpw --mspid INMETROMSP \
        --ca-name ca  --output resources/peer-inmetro.yaml

kubectl hlf utils adduser --userPath=resources/peer-inmetro.yaml --config=resources/network.yaml --username=admin --mspid=INMETROMSP
``` 

(execute o comando na pasta raíz do projeto). Será criado o arquivo network.yaml na pasta resources.

- Copie o conteúdo do arquivo network.yaml

- Abra o terminal e acesse a pasta do cliente (fabcar)

- Abra o arquivo "connection-org.yaml" e substitua o seu conteúdo pelo conteúdo do arquivo encontrado em "../resources/network.yaml".

- Feito isso, dentro do arquivo "connection-org.yaml", navegue até a seção "Organizations" e, na subseção INMETROMSP, INSIRA o seguinte campo

```
    certificateAuthorities:
      - inmetro-ca.default
```

- Agora, na seção "client" campo "organization", deixe o seguinte valor

```
    client:
        organization: INMETROMSP
```

Com isso você está pronto para executar os clientes

# Tutorial de como usar o cliente FieldClimate

= Acesse a pasta do cliente e execute o cliente com o comando:

```
    go run main.go
```

Você também pode modificar o cliente, verificando as funções no chaincode e as inserindo no cliente com o invokeCC ou queryCC.

- Funções que inserem dados (putState) devem ser invocadas com o invokeCC
- Funções que buscam dados (getState) devem ser invocadas com o queryCC