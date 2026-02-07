# Quickstart: smartLattes

## Prerequisites

- Go 1.23+ installed (apenas para desenvolvimento local)
- Acesso ao MongoDB em `192.168.1.10:27017`

## Local Development

```bash
# Clone and enter the project
cd smartLattes

# Create .env file with your credentials (this file is in .gitignore)
cp .env.example .env
# Edit .env with your real MongoDB credentials

# Download dependencies
go mod tidy

# Run locally (loads .env automatically)
go run ./cmd/smartlattes

# Application starts at http://localhost:8080
```

> **IMPORTANTE**: O arquivo `.env` contem credenciais e esta no `.gitignore`. Nunca commitar credenciais no repositorio.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_URI` | *(obrigatoria)* | String de conexao MongoDB (ex: `mongodb://user:password@hostname:27017/?authSource=dbname`) |
| `MONGODB_DATABASE` | `smartLattes` | Nome do banco de dados |
| `PORT` | `8080` | Porta do servidor HTTP |
| `MAX_UPLOAD_SIZE` | `10485760` | Tamanho maximo de upload em bytes (10MB) |

## Docker Build & Push

```bash
# Build the image
docker build -t ghcr.io/edalcin/smartlattes:latest .

# Push to GitHub Container Registry
docker push ghcr.io/edalcin/smartlattes:latest
```

## Deploy no Unraid

### Passo 1: Adicionar o container via interface web

1. Acesse a interface web do Unraid (ex: `http://<ip-do-unraid>`)
2. Va em **Docker** no menu superior
3. Clique em **Add Container** (canto inferior esquerdo)

### Passo 2: Configurar o container

Preencha os campos conforme abaixo:

| Campo | Valor |
|-------|-------|
| **Name** | `smartLattes` |
| **Repository** | `ghcr.io/edalcin/smartlattes:latest` |
| **Network Type** | `bridge` |
| **WebUI** | `http://[IP]:[PORT:8080]` |

### Passo 3: Mapear a porta

Clique em **Add another Path, Port, Variable, Label or Device** e selecione **Port**:

| Campo | Valor |
|-------|-------|
| **Config Type** | Port |
| **Name** | `Web UI Port` |
| **Container Port** | `8080` |
| **Host Port** | `8080` |
| **Connection Type** | TCP |

> Se a porta 8080 ja estiver em uso no Unraid, altere apenas o **Host Port** (ex: `8081`). O Container Port deve permanecer `8080`.

### Passo 4: Adicionar variaveis de ambiente

Clique em **Add another Path, Port, Variable, Label or Device** e selecione **Variable** para cada variavel:

#### Variavel 1: MONGODB_URI (obrigatoria)

| Campo | Valor |
|-------|-------|
| **Config Type** | Variable |
| **Name** | `MongoDB Connection String` |
| **Key** | `MONGODB_URI` |
| **Value** | `mongodb://user:password@hostname:27017/?authSource=dbname` |

#### Variavel 2: MONGODB_DATABASE (opcional)

| Campo | Valor |
|-------|-------|
| **Config Type** | Variable |
| **Name** | `MongoDB Database Name` |
| **Key** | `MONGODB_DATABASE` |
| **Value** | `smartLattes` |

### Passo 5: Aplicar

1. Clique em **Apply** no final da pagina
2. O Unraid vai baixar a imagem do `ghcr.io/edalcin/smartlattes:latest` e iniciar o container
3. Apos iniciar, clique no icone do container e selecione **WebUI** para abrir o smartLattes no navegador

### Passo 6: Verificar

- Acesse `http://<ip-do-unraid>:8080` no navegador
- A homepage do smartLattes deve aparecer com o menu de navegacao
- Clique em "Upload CV" e faca upload do arquivo XML de teste

### Atualizando o container

Quando uma nova versao da imagem for publicada no ghcr.io:

1. Va em **Docker** na interface do Unraid
2. Clique em **Check for Updates** ou clique no icone do container `smartLattes`
3. Selecione **Force Update** para baixar a imagem mais recente
4. O container sera reiniciado automaticamente com a nova versao

### Troubleshooting no Unraid

- **Container nao inicia**: Clique no icone do container > **Logs** para ver mensagens de erro
- **Erro de conexao MongoDB**: Verifique se o MongoDB em `192.168.1.10:27017` esta acessivel a partir do Unraid. Teste com: `docker exec smartLattes wget -qO- http://192.168.1.10:27017` no terminal do Unraid
- **Imagem nao encontrada**: Verifique se a imagem `ghcr.io/edalcin/smartlattes` esta publica ou se o Unraid tem credenciais configuradas para o GitHub Container Registry

## Testing

```bash
# Run all tests
go test ./...

# Test with the example XML file
curl -F "file=@docs/8334174268306003.xml" http://localhost:8080/api/upload
```

## Project Structure

```
smartLattes/
├── cmd/
│   └── smartlattes/
│       └── main.go              # Entry point, server setup
├── internal/
│   ├── handler/
│   │   ├── upload.go            # Upload endpoint handler
│   │   ├── pages.go             # HTML page handlers
│   │   └── health.go            # Health check handler
│   ├── parser/
│   │   └── lattes.go            # XML parsing + JSON conversion
│   ├── store/
│   │   └── mongo.go             # MongoDB operations (upsert)
│   └── static/
│       ├── index.html           # Homepage with navigation
│       ├── upload.html          # Upload form page
│       ├── explorer.html        # Coming soon placeholder
│       ├── css/
│       │   └── style.css        # Application styles
│       └── js/
│           └── upload.js        # Upload form behavior
├── docs/
│   └── 8334174268306003.xml     # Example Lattes XML
├── specs/                       # Feature specifications
├── Dockerfile                   # Multi-stage build
├── go.mod
└── go.sum
```
