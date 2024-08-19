## Sobre o projeto:

Este repositório apresenta uma plataforma de monetização que utiliza tecnologia blockchain para desenvolver um ecossistema seguro e transparente, destinado à coleta, armazenamento e comercialização de dados de telemetria veicular. A monetização é definida a partir de regras implementadas por contratos inteligentes, que avaliam não apenas o esforço do condutor em compartilhar dados, mas também como o veículo é utilizado e seu consumo de combustível. A solução está sendo implementada com o uso de ferramentas seguras como o Hyperledger Fabric e Kubernetes, promovendo transparência, integridade e incentivando práticas de condução sustentáveis, viabilizando assim a monetização confiável de dados veiculares com foco em eficiência energética.

[Telemetria Veicular](telemetria-c.pdf)

# Tutorial

Caso a plataforma não esteja instalada na máquina, existem alguns requisitos mínimos necessários para utilizar esta rede. É fundamental garantir que o sistema tenha uma conexão de internet estável, espaço de armazenamento adequado, e software atualizado compatível com a tecnologia blockchain.

- Linux (testado com Ubuntu 22.04)
- [Kubectl](https://kubernetes.io/pt-br/docs/tasks/tools/install-kubectl-linux/)
- [Krew](https://krew.sigs.k8s.io/)
- [KinD](https://kind.sigs.k8s.io/) ou [K3d](https://k3d.io/v5.6.0/)
- [Istio](https://istio.io/latest/)
- [Helm](https://helm.sh/)
- [JQ](https://jqlang.github.io/jq/download/)
- [Docker](https://docs.docker.com/get-docker/)

Para instalar todos os requisitos automaticamente, use o seguinte script

```bash
chmod 777 install.sh
./install.sh
```

## Script para levantar a rede automáticamente

Para iniciar a rede (ambiente de teste) automaticamente pelo terminal, utilize o script ./network.sh:

```bash
  ./network.sh up
```

Ele possui os seguintes comandos:

```bash
    echo "'up' - Inicia o cluster Kubernetes"
    echo "'chaincode <nome do chaincode>' - Realiza o deploy do chaincode"
    echo "'operator' - Inicia o Operator UI e o Operator API"
    echo "'upgrade' <nome do chaincode> <versao> <sequencia>- Faz o upgrade do chaincode"
    echo "'down' - Destrói o cluster Kubernetes e todos os recursos criados"
    echo "'help' - Mostra alguns comandos que podem ajudar a diagnosticar problemas na rede"
```

Alternativamente, siga o tutorial abaixo para levantar a rede manualmente

## 1. Criar Cluster Kubernetes

Para iniciar o deploy da rede Fabric, é necessário criar um cluster Kubernetes. Neste exemplo, utilizaremos o KinD.
As portas 80 e 443 são comumente usadas para serviços web. A porta 80 é utilizada para HTTP (Hypertext Transfer Protocol) e a porta 443 para HTTPS (HTTP Secure). Garantir que essas portas estejam disponíveis é crucial para o funcionamento correto de muitos serviços e aplicações que dependem de comunicação web, incluindo a instalação e operação de um cluster Kubernetes com KinD (Kubernetes in Docker).

1. **Verificar se as portas estão em uso:**

   Abra o terminal e execute o comando abaixo:

   ```bash
   sudo lsof -i -P -n | grep LISTEN
   ```

   Isso listará todas as portas em uso. Verifique se as portas 80 e 443 estão listadas. Se não estiverem, elas estão disponíveis.

2. **Verificar com `netstat`:**

   Outro comando que pode ser utilizado é o `netstat`:

   ```bash
   sudo netstat -tuln | grep -E "(:80|:443)"
   ```

Se não houver saída, significa que as portas 80 e 443 estão disponíveis.

Se descobrir que as portas 80 ou 443 estão em uso, você pode encerrar o processo que as está utilizando. Use o comando apropriado para o seu sistema operacional para identificar e encerrar o processo.

No Linux/MacOS
Identifique o PID (Process ID) do processo usando a porta:

```bash
sudo lsof -i :80
sudo lsof -i :443
```

Depois encerre o processo:

```bash
sudo kill -9 <PID>
```

Certificando-se de que essas portas estão disponíveis, você pode prosseguir com a criação do cluster Kubernetes usando o KinD.

### Usando KinD

```bash
cat << EOF > resources/kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.28.0
  extraPortMappings:
  - containerPort: 30949
    hostPort: 80
  - containerPort: 30950
    hostPort: 443
EOF

kind create cluster --config=./resources/kind-config.yaml

export STORAGE_CLASS=standard
export DATABASE=couchdb
```

## 2. Instalação do Istio

Instale o Istio no cluster Kubernetes

```bash

kubectl create namespace istio-system

istioctl operator init

kubectl apply -f - <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: istio-gateway
  namespace: istio-system
spec:
  addonComponents:
    grafana:
      enabled: false
    kiali:
      enabled: false
    prometheus:
      enabled: false
    tracing:
      enabled: false
  components:
    ingressGateways:
      - enabled: true
        k8s:
          hpaSpec:
            minReplicas: 1
          resources:
            limits:
              cpu: 500m
              memory: 512Mi
            requests:
              cpu: 100m
              memory: 128Mi
          service:
            ports:
              - name: http
                port: 80
                targetPort: 8080
                nodePort: 30949
              - name: https
                port: 443
                targetPort: 8443
                nodePort: 30950
            type: NodePort
        name: istio-ingressgateway
    pilot:
      enabled: true
      k8s:
        hpaSpec:
          minReplicas: 1
        resources:
          limits:
            cpu: 300m
            memory: 512Mi
          requests:
            cpu: 100m
            memory: 128Mi
  meshConfig:
    accessLogFile: /dev/stdout
    enableTracing: false
    outboundTrafficPolicy:
      mode: ALLOW_ANY
  profile: default

EOF
```

### Configurar DNS Interno

```bash
CLUSTER_IP=$(kubectl -n istio-system get svc istio-ingressgateway -o json | jq -r .spec.clusterIP)
kubectl apply -f - <<EOF
kind: ConfigMap
apiVersion: v1
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        rewrite name regex (.*)\.localho\.st host.ingress.internal
        hosts {
          ${CLUSTER_IP} host.ingress.internal
          fallthrough
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf {
           max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
    }
EOF
```

## 3. Instalar o HLF Operator

Nesta etapa instalaremos o operador Kubernetes para o Fabric. Isso irá instalar:

- CRD (Custom Resource Definitions) to deploy Certification Fabric Peers, Orderers and Authorities
- Deploy the program to deploy the nodes in Kubernetes

```bash
# instalar localmente (quando a rede bloqueia, mas é uma versão antiga)
helm install hlf-operator resources/hlf-operator-1.9.2.tgz
```

### Instalar o plugin Kubectl

Em caso de erro, verifique se o [Krew](https://krew.sigs.k8s.io/docs/user-guide/setup/install/) está instalado.

A seguir, instale o Kubectl com o seguinte comando:

```bash
kubectl krew install hlf
```

## 4. Deploy de organizações

### Environment Variables for AMD (Default)

```bash
export ORDERER_IMAGE=hyperledger/fabric-orderer
export ORDERER_VERSION=2.5.5

export CA_IMAGE=hyperledger/fabric-ca
export CA_VERSION=1.5.7
```

<h2>Execução do Hyperledger Fabric</h2>

Você pode usar Fabric CAs para gerar material criptográfico, onde as CAs assinam os certificados e chaves gerados para criar uma raiz de confiança válida para cada organização. O script usa Docker Compose para iniciar três CAs: uma para cada organização de mesmo nível e uma para a organização do orderer. As configurações dos servidores Fabric CA estão no diretório "organizations/fabric-ca". No mesmo diretório, o script "registerEnroll.sh" utiliza o cliente Fabric CA para criar as identidades, certificados e pastas MSP necessárias para configurar a rede de teste em "organizations/ordererOrganizations".

### Criação do CA para INMETRO

```bash
kubectl hlf ca create  --image=$CA_IMAGE --version=$CA_VERSION --storage-class=$STORAGE_CLASS --capacity=1Gi --name=inmetro-ca \
    --enroll-id=enroll --enroll-pw=enrollpw --hosts=inmetro-ca.localho.st --istio-port=443

kubectl wait --timeout=180s --for=condition=Running fabriccas.hlf.kungfusoftware.es --all
```

Verifique se o CA foi implementado e funciona:

```bash
curl -k https://inmetro-ca.localho.st:443/cainfo
```

Registre um usuário peer no CA da Organização INMETRO

```bash
# register user in CA for peers
kubectl hlf ca register --name=inmetro-ca --user=peer --secret=peerpw --type=peer \
 --enroll-id enroll --enroll-secret=enrollpw --mspid INMETROMSP

```

### Deploy de peers para a organização INMETRO (escolha um apenas)

OBS: O peer a seguir não suporta [CCAS](ccas)

```bash
export PEER_IMAGE=quay.io/kfsoftware/fabric-peer
export PEER_VERSION=2.4.1-v0.0.3
export MSP_ORG=INMETROMSP
export PEER_SECRET=peerpw

kubectl hlf peer create --statedb=$DATABASE --image=$PEER_IMAGE --version=$PEER_VERSION --storage-class=$STORAGE_CLASS --enroll-id=peer --mspid=$MSP_ORG \
--enroll-pw=$PEER_SECRET --capacity=5Gi --name=inmetro-peer0 --ca-name=inmetro-ca.default --k8s-builder=true --hosts=peer0-inmetro.localho.st --istio-port=443

kubectl wait --timeout=180s --for=condition=Running fabricpeers.hlf.kungfusoftware.es --all

# leva alguns minutos

```

Verifique se os peers foram implementados e funcionam:

```bash
openssl s_client -connect peer0-inmetro.localho.st:443

```

### Deploy de uma organização `Orderer`

para fazer o deploy de uma organização orderer, temos que:

1. Criar um certification authority (CA)
2. Registrar usuário `orderer` com senha `ordererpw`
3. Criar orderer

### Criar o CA

```bash
kubectl hlf ca create  --image=$CA_IMAGE --version=$CA_VERSION --storage-class=$STORAGE_CLASS --capacity=1Gi --name=ord-ca \
    --enroll-id=enroll --enroll-pw=enrollpw --hosts=ord-ca.localho.st --istio-port=443

kubectl wait --timeout=180s --for=condition=Running fabriccas.hlf.kungfusoftware.es --all
```

Verifique se a certificação foi implementada e funciona:

```bash
curl -vik https://ord-ca.localho.st:443/cainfo
```

### Registre o usuário `orderer`

```bash
kubectl hlf ca register --name=ord-ca --user=orderer --secret=ordererpw \
    --type=orderer --enroll-id enroll --enroll-secret=enrollpw --mspid=OrdererMSP --ca-url="https://ord-ca.localho.st:443"

```

### Deploy de três orderers

```bash
  kubectl hlf ordnode create --image=$ORDERER_IMAGE --version=$ORDERER_VERSION \
      --storage-class=$STORAGE_CLASS --enroll-id=orderer --mspid=OrdererMSP \
      --enroll-pw=ordererpw --capacity=2Gi --name=ord-node0 --ca-name=ord-ca.default \
      --hosts=orderer0-ord.localho.st --istio-port=443 --admin-hosts=admin-orderer0-ord.localho.st

kubectl wait --timeout=180s --for=condition=Running fabricorderernodes.hlf.kungfusoftware.es --all
```

Verifique se os orderers funcionam:

```bash
kubectl get pods
```

```bash
openssl s_client -connect orderer0-ord.localho.st:443
```

### Criar canal

Para criar o canal nós precisamos criar o "wallet secret", que irá conter as identidades usadas pelo bevel operator para gerenciar o canal

## Registrar e matricular identidade OrdererMSP

```bash
  ## register OrdererMSP Identity
  kubectl hlf ca register --name=ord-ca --user=admin --secret=adminpw \
      --type=admin --enroll-id enroll --enroll-secret=enrollpw --mspid=OrdererMSP

  kubectl hlf identity create --name orderer-admin-sign --namespace default \
      --ca-name ord-ca --ca-namespace default \
      --ca ca --mspid OrdererMSP --enroll-id admin --enroll-secret adminpw  # sign identity

  kubectl hlf identity create --name orderer-admin-tls --namespace default \
      --ca-name ord-ca --ca-namespace default \
      --ca tlsca --mspid OrdererMSP --enroll-id admin --enroll-secret adminpw l # tls identity
```

## Registrar e matricular identidade INMETROMSP

```bash
  ## register INMETROMSP Identity
  kubectl hlf ca register --name=inmetro-ca --namespace=default --user=admin --secret=adminpw \
      --type=admin --enroll-id enroll --enroll-secret=enrollpw --mspid=INMETROMSP

  # enroll
  kubectl hlf identity create --name inmetro-admin --namespace default \
      --ca-name inmetro-ca --ca-namespace default \
      --ca ca --mspid INMETROMSP --enroll-id admin --enroll-secret adminpw
```

## Criando canal principal

```bash
export PEER_ORG_SIGN_CERT=$(kubectl get fabriccas inmetro-ca -o=jsonpath='{.status.ca_cert}')
export PEER_ORG_TLS_CERT=$(kubectl get fabriccas inmetro-ca -o=jsonpath='{.status.tlsca_cert}')
export IDENT_8=$(printf "%8s" "")
export ORDERER_TLS_CERT=$(kubectl get fabriccas ord-ca -o=jsonpath='{.status.tlsca_cert}' | sed -e "s/^/${IDENT_8}/" )
export ORDERER0_TLS_CERT=$(kubectl get fabricorderernodes ord-node0 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/" )
export ORDERER1_TLS_CERT=$(kubectl get fabricorderernodes ord-node1 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/" )
export ORDERER2_TLS_CERT=$(kubectl get fabricorderernodes ord-node2 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/" )

kubectl apply -f - <<EOF
apiVersion: hlf.kungfusoftware.es/v1alpha1
kind: FabricMainChannel
metadata:
  name: demo
spec:
  name: demo
  adminOrdererOrganizations:
    - mspID: OrdererMSP
  adminPeerOrganizations:
    - mspID: INMETROMSP
  channelConfig:
    application:
      acls: null
      capabilities:
        - V2_0
      policies: null
    capabilities:
      - V2_0
    orderer:
      batchSize:
        absoluteMaxBytes: 1048576
        maxMessageCount: 10
        preferredMaxBytes: 524288
      batchTimeout: 2s
      capabilities:
        - V2_0
      etcdRaft:
        options:
          electionTick: 10
          heartbeatTick: 1
          maxInflightBlocks: 5
          snapshotIntervalSize: 16777216
          tickInterval: 500ms
      ordererType: etcdraft
      policies: null
      state: STATE_NORMAL
    policies: null
  externalOrdererOrganizations: []
  peerOrganizations:
    - mspID: INMETROMSP
      caName: "inmetro-ca"
      caNamespace: "default"
    - mspID: PUCMSP
      caName: "puc-ca"
      caNamespace: "default"
  identities:
    OrdererMSP:
      secretKey: user.yaml
      secretName: orderer-admin-tls
      secretNamespace: default
    OrdererMSP-sign:
      secretKey: user.yaml
      secretName: orderer-admin-sign
      secretNamespace: default
    INMETROMSP:
      secretKey: user.yaml
      secretName: inmetro-admin
      secretNamespace: default
  externalPeerOrganizations: []
  ordererOrganizations:
    - caName: "ord-ca"
      caNamespace: "default"
      externalOrderersToJoin:
        - host: ord-node0
          port: 7053
        - host: ord-node1
          port: 7053
        - host: ord-node2
          port: 7053
      mspID: OrdererMSP
      ordererEndpoints:
        - ord-node0:7050
        - ord-node1:7050
        - ord-node2:7050
      orderersToJoin: []
  orderers:
    - host: ord-node0
      port: 7050
      tlsCert: |-
${ORDERER0_TLS_CERT}
    - host: ord-node1
      port: 7050
      tlsCert: |-
${ORDERER1_TLS_CERT}
    - host: ord-node2
      port: 7050
      tlsCert: |-
${ORDERER2_TLS_CERT}
EOF


```

## Inserir peers do INMETRO no canal

```bash

export IDENT_8=$(printf "%8s" "")
export ORDERER0_TLS_CERT=$(kubectl get fabricorderernodes ord-node0 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/" )

kubectl apply -f - <<EOF
apiVersion: hlf.kungfusoftware.es/v1alpha1
kind: FabricFollowerChannel
metadata:
  name: demo-inmetromsp
spec:
  anchorPeers:
    - host: inmetro-peer0.default
      port: 7051
  hlfIdentity:
    secretKey: user.yaml
    secretName: inmetro-admin
    secretNamespace: default
  mspId: INMETROMSP
  name: demo
  externalPeersToJoin: []
  orderers:
    - certificate: |
${ORDERER0_TLS_CERT}
      url: grpcs://ord-node0.default:7050
  peersToJoin:
    - name: inmetro-peer0
      namespace: default
EOF
```

## Instalação do chaincode

Caso queira testar o Chaincode As a Service, veja o tutorial em [ccas](ccas). Caso contrário, prossiga com a instalação do chaincode local a seguir.

## Instalação de chaincode Local

### Preparar string / arquivo de conexão para um peer

Para preparar a string de conexão, precisamos:

1. Obter a string de conexão sem usuários para a organização Org1MSP e OrdererMSP
2. Registre um usuário no CA para assinatura (registro)
3. Obter os certificados usando o usuário criado no passo 2 (enroll)
4. Anexar o usuário à string de conexão

(Repetir 2, 3 e 4 para Org2)

---

1. Obter a string de conexão sem usuários para a organização Org1MSP e OrdererMSP

```bash
kubectl hlf inspect -c=demo --output resources/network.yaml -o INMETROMSP -o PUCMSP -o OrdererMSP
```

### Registrar usuário para INMETRO

2. Registre um usuário no CA para assinatura (registro)

```bash
kubectl hlf ca register --name=inmetro-ca --user=admin --secret=adminpw --type=admin \
 --enroll-id enroll --enroll-secret=enrollpw --mspid INMETROMSP
```

3. Obter os certificados usando o usuário criado no passo 2 (enroll)

```bash
kubectl hlf ca enroll --name=inmetro-ca --user=admin --secret=adminpw --mspid INMETROMSP \
        --ca-name ca  --output resources/peer-inmetro.yaml
```

4. Anexar o usuário à string de conexão

```bash
kubectl hlf utils adduser --userPath=resources/peer-inmetro.yaml --config=resources/network.yaml --username=admin --mspid=INMETROMSP
```

### Instalação do chaincode

Com o arquivo de conexão preparado, vamos instalar o chaincode no peer que possua o atributo k8s-builder, como explicado no passo de deploy de peers

```bash
kubectl hlf chaincode install --path=./chaincodes/VehicleContract \
    --config=resources/network.yaml --language=golang --label=VehicleContract --user=admin --peer=inmetro-peer0.default

kubectl hlf chaincode install --path=./chaincodes/VehicleContract \
    --config=resources/network.yaml --language=golang --label=VehicleContract --user=admin --peer=puc-peer0.default

# this can take 3-4 minutes
```

Verificação de chaincodes instalados

```bash
kubectl hlf chaincode queryinstalled --config=resources/network.yaml --user=admin --peer=inmetro-peer0.default

kubectl hlf chaincode queryinstalled --config=resources/network.yaml --user=admin --peer=puc-peer0.default
```

Aprovar chaincode

```bash
  export PACKAGE_ID=$(kubectl hlf chaincode calculatepackageid --path=./examples/chaincodes/$CHAINCODE_LABEL --language=go --label=$CHAINCODE_LABEL)
  echo "PACKAGE_ID=$PACKAGE_ID"

#Organização INMETRO
kubectl hlf chaincode approveformyorg --config=resources/network.yaml --user=admin --peer=inmetro-peer0.default \
    --package-id=$PACKAGE_ID \
    --version "1.0" --sequence 1 --name=VehicleContract \
    --policy="AND('INMETROMSP.member', 'PUCMSP.member')" --channel=demo

# Organização PUC

kubectl hlf chaincode approveformyorg --config=resources/network.yaml --user=admin --peer=puc-peer0.default \
    --package-id=$PACKAGE_ID \
    --version "1.0" --sequence 1 --name=VehicleContract \
    --policy="AND('INMETROMSP.member', 'PUCMSP.member')" --channel=demo

# Fazer o commit do chaincode

kubectl hlf chaincode commit --config=resources/network.yaml --mspid=INMETROMSP --user=admin \
    --version "1.0" --sequence 1 --name=$CHAINCODE_LABEL \
    --policy="AND('INMETROMSP.member')" --channel=demo
```

### Testar chaincode

```bash
kubectl hlf chaincode invoke --config=resources/network.yaml \
    --user=admin --peer=puc-peer0.default \
    --chaincode=VehicleContract --channel=demo \
    --fcn=Createuser -a '["teste"]' '["teste"]' '["teste"]' '["teste"]'
```

## Fazendo upgrade de chaincode

Para atualizar o chaincode, repita o mesmo proceso de instalação, alterando os valores dos parâmetros `--version` e `---sequence`.
Alternativamente, utilize variáveis de ambiente, como no exemplo a seguir:

```bash
export CC_VERSION="1.1"
export CC_SEQUENCE=2

kubectl hlf chaincode install --path=./chaincode/$CHAINCODE_LABEL \
    --config=resources/network.yaml --language=golang --label=$CHAINCODE_LABEL --user=admin --peer=inmetro-peer0.default

export PACKAGE_ID=$(kubectl hlf chaincode calculatepackageid --path=chaincode/$CHAINCODE_LABEL --language=golang --label=$CHAINCODE_LABEL)
echo "PACKAGE_ID=$PACKAGE_ID"

kubectl hlf chaincode approveformyorg --config=resources/network.yaml --user=admin --peer=inmetro-peer0.default \
  --package-id=$PACKAGE_ID \
  --version $CC_VERSION --sequence $CC_SEQUENCE --name=$CHAINCODE_LABEL \
  --policy="AND('INMETROMSP.member')" --channel=demo

kubectl hlf chaincode commit --config=resources/network.yaml --mspid=INMETROMSP --user=admin \
    --version $CC_VERSION --sequence $CC_SEQUENCE --name=$CHAINCODE_LABEL \
    --policy="AND('INMETROMSP.member')" --channel=demo

```

## Usando clientes:

[Usando Clientes](cliente)

# Operator UI

O HLF Operator UI fornece uma interface gráfica para uma experiência de usuário mais conveniente. O Operator UI torna mais fácil o processo de criar, clonar, supervisionar, editar e deletar os nós de Peers, Orderers e CAs.

Ele consiste de dois componentes

# Levantando Operator UI

O HLF Operator UI fornece uma interface gráfica para uma experiência de usuário mais conveniente. O Operator UI torna mais fácil o processo de criar, clonar, supervisionar, editar e deletar os nós de Peers, Orderers e CAs.

Ele consiste de dois componentes

#### Operator API

Fornece acesso aos dados para serem exibidos pelo Operator UI

- Canais
- Peers
- Orderer Nodes
- Certificate Authorities

#### Operator UI

Interface gráfica que permite:

- Criar peers
- Criar CAs
- Criar orderers
- Renovar certificados

## Levantando o Operator UI

Primeiro deve-se levantar o Operator API

```bash
export API_URL=api-operator.localho.st # URL de acesso

kubectl hlf operatorapi create --name=operator-api --namespace=default --hosts=$API_URL --ingress-class-name=istio
```

Agora, para levantar o Operator UI

```bash
export HOST=operator-ui.localho.st
export API_URL="http://api-operator.localho.st/graphql"

kubectl hlf operatorui create --name=operator-ui --namespace=default --hosts=$HOST --ingress-class-name=istio --api-url=$API_URL
```

Verifique se eles estão funcionando com o comando a seguir

```
kubectl get pods
```

Seus containeres devem estar com o estado "Running"

## Acessando o Operator UI

No navegador, insira a URL:

operator-ui.localho.st

## Finalizando

A essa altura, você deve ter:

- Um serviço de ordenação com 3 orderers e CA
- Organização INMETRO com 1 peer e CA
- Organização PUC com 1 peer e CA
- Um canal chamado "demo"
- Um chaincode instalado nos peers do INMETRO e PUC, aprovado e "commitado"

## Derrubando o ambiente

```bash
kubectl delete fabricorderernodes.hlf.kungfusoftware.es --all-namespaces --all
kubectl delete fabricpeers.hlf.kungfusoftware.es --all-namespaces --all
kubectl delete fabriccas.hlf.kungfusoftware.es --all-namespaces --all
kubectl delete fabricchaincode.hlf.kungfusoftware.es --all-namespaces --all
kubectl delete fabricmainchannels --all-namespaces --all
kubectl delete fabricfollowerchannels --all-namespaces --all

kind delete cluster
```
