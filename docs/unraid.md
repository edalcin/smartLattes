# Instalando o smartLattes no Unraid

Este guia descreve passo a passo como instalar o container Docker do smartLattes em um servidor Unraid utilizando a interface grafica.

## Pre-requisitos

- Servidor Unraid com o servico Docker habilitado
- Acesso a interface web do Unraid (ex: `http://192.168.1.xxx`)
- MongoDB acessivel na rede local (ex: `192.168.1.10:27017`)
- A imagem `ghcr.io/edalcin/smartlattes:latest` publicada no GitHub Container Registry

## Passo 1: Acessar a aba Docker

1. Abra o navegador e acesse a interface web do Unraid
2. No menu superior, clique em **Docker**
3. Verifique que o servico Docker esta **Enabled** (indicador verde no canto superior direito)

## Passo 2: Adicionar novo container

1. Na parte inferior da pagina Docker, clique no botao **Add Container**
2. Sera exibido o formulario de configuracao do container

## Passo 3: Configuracao basica

Preencha os campos do formulario conforme a tabela abaixo:

| Campo | Valor |
|-------|-------|
| **Name** | `smartLattes` |
| **Repository** | `ghcr.io/edalcin/smartlattes:latest` |
| **Network Type** | `bridge` |
| **WebUI** | `http://[IP]:[PORT:8080]` |

> **Nota sobre Network Type**: se o Unraid e o MongoDB estiverem na mesma rede, `bridge` funciona. Se houver problemas de conectividade com o MongoDB, mude para `host`.

## Passo 4: Mapear a porta

Clique em **Add another Path, Port, Variable, Label or Device** e configure:

1. Selecione **Port** no menu suspenso
2. Preencha os campos:

| Campo | Valor |
|-------|-------|
| **Config Type** | Port |
| **Name** | `Web Interface` |
| **Container Port** | `8080` |
| **Host Port** | `8080` |
| **Connection Type** | TCP |

> Se a porta 8080 ja estiver em uso por outro container no Unraid, altere apenas o **Host Port** para um valor livre (ex: `8081`, `9090`). O **Container Port** deve permanecer `8080`.

## Passo 5: Configurar variaveis de ambiente

Para cada variavel abaixo, clique em **Add another Path, Port, Variable, Label or Device**, selecione **Variable** e preencha:

### Variavel 1: MONGODB_URI (obrigatoria)

| Campo | Valor |
|-------|-------|
| **Config Type** | Variable |
| **Name** | `MongoDB Connection String` |
| **Key** | `MONGODB_URI` |
| **Value** | `mongodb://usuario:senha@192.168.1.10:27017/?authSource=smartLattes` |

> **IMPORTANTE**: Substitua `usuario`, `senha` e o IP pelos valores reais do seu MongoDB. Se a senha contiver caracteres especiais (`@`, `#`, etc.), eles devem estar codificados em URL (ex: `@` vira `%40`).

### Variavel 2: MONGODB_DATABASE (opcional)

| Campo | Valor |
|-------|-------|
| **Config Type** | Variable |
| **Name** | `MongoDB Database Name` |
| **Key** | `MONGODB_DATABASE` |
| **Value** | `smartLattes` |

> Se omitida, o valor padrao `smartLattes` sera utilizado.

## Passo 6: Revisar e aplicar

1. Revise todas as configuracoes no formulario
2. O resumo deve mostrar:
   - Repository: `ghcr.io/edalcin/smartlattes:latest`
   - Port: `8080` -> `8080/tcp`
   - Variable: `MONGODB_URI` = sua string de conexao
   - Variable: `MONGODB_DATABASE` = `smartLattes`
3. Clique em **Apply**
4. O Unraid ira baixar a imagem e iniciar o container

## Passo 7: Verificar a instalacao

1. Apos o container iniciar (poucos segundos), ele aparecera na lista de containers Docker
2. O icone do container deve mostrar status **Started** (verde)
3. Clique no icone do container e selecione **WebUI**
4. O navegador abrira a homepage do smartLattes
5. Verifique que o menu exibe "Enviar CV" e "Explorar Dados"
6. Clique em "Enviar CV" e teste o upload com um arquivo XML do Lattes

### Verificacao via health check

Acesse `http://<ip-do-unraid>:8080/api/health` no navegador. A resposta esperada:

```json
{"status":"healthy","mongodb":"connected"}
```

Se aparecer `"mongodb":"disconnected"`, verifique a string de conexao MONGODB_URI e a acessibilidade do MongoDB.

## Atualizando o container

Quando uma nova versao for publicada:

1. Na aba **Docker**, localize o container `smartLattes`
2. Se houver atualizacao disponivel, um icone de atualizacao aparecera
3. Clique no icone do container e selecione **Force Update**
4. O Unraid baixara a imagem mais recente e reiniciara o container
5. Nenhuma reconfiguracao e necessaria — as variaveis de ambiente sao preservadas

## Removendo o container

1. Na aba **Docker**, clique no icone do container `smartLattes`
2. Selecione **Remove**
3. Confirme a remocao
4. Para remover tambem a imagem Docker, va em **Docker** > **Docker Image** e delete `ghcr.io/edalcin/smartlattes`

## Resolucao de problemas

### Container nao inicia

1. Clique no icone do container > **Logs**
2. Procure por mensagens de erro
3. Erro comum: `MONGODB_URI e obrigatorio` — significa que a variavel de ambiente nao foi configurada

### Erro de conexao com MongoDB

1. Verifique se o MongoDB esta rodando: `http://192.168.1.10:27017` deve responder
2. Verifique se usuario e senha estao corretos na MONGODB_URI
3. Verifique se o `authSource` corresponde ao banco de dados onde o usuario foi criado
4. Se estiver usando Network Type `bridge`, tente mudar para `host`

### Pagina nao carrega

1. Verifique se o container esta rodando (status verde na aba Docker)
2. Verifique se a porta esta mapeada corretamente
3. Tente acessar diretamente: `http://<ip-do-unraid>:8080`
4. Verifique se nao ha conflito de porta com outro container

### Upload falha com erro de banco de dados

1. Acesse `http://<ip-do-unraid>:8080/api/health` para verificar o status do MongoDB
2. Se mostrar `disconnected`, o problema esta na conexao com o MongoDB
3. Verifique a MONGODB_URI e a acessibilidade da rede
