## Sobre o projeto:
Este repositório apresenta uma plataforma de monetização que utiliza tecnologia blockchain para desenvolver um ecossistema seguro e transparente, destinado à coleta, armazenamento e comercialização de dados de telemetria veicular. A monetização é definida a partir de regras implementadas por contratos inteligentes, que avaliam não apenas o esforço do condutor em compartilhar dados, mas também como o veículo é utilizado e seu consumo de combustível. A solução está sendo implementada com o uso de ferramentas seguras como o Hyperledger Fabric e Kubernetes, promovendo transparência, integridade e incentivando práticas de condução sustentáveis, viabilizando assim a monetização confiável de dados veiculares com foco em eficiência energética.

[Telemetria Veicular](https://github.com/stephanietygna/EnerBlock/blob/main/telemetria-c%20(2).pdf)

# Tutorial

Caso a plataforma não esteja instalada na máquina, existem alguns requisitos mínimos necessários para utilizar esta rede. É fundamental garantir que o sistema tenha uma conexão de internet estável, espaço de armazenamento adequado, e software atualizado compatível com a tecnologia blockchain.

- Linux (testado com Ubuntu 22.04)
- [Kubectl](https://kubernetes.io/pt-br/docs/tasks/tools/install-kubectl-linux/)
- [Krew](https://krew.sigs.k8s.io/)
- [KinD](https://kind.sigs.k8s.io/) ou [K3d](https://k3d.io/v5.6.0/)
- [Istio](https://istio.io/latest/ ) 
- [Helm](https://helm.sh/)
- [JQ](https://jqlang.github.io/jq/download/)
- [Docker](https://docs.docker.com/get-docker/)

Para instalar todos os requisitos automaticamente, use o seguinte script

```bash
chmod 777 install.sh
./install.sh
```
