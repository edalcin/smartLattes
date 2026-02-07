# smartLattes

**Estabelecimento de redes de pesquisadores baseadas na análise de seus perfis na Plataforma Lattes, com auxílio da Inteligência Artificial.**

## Motivação

A [Plataforma Lattes](https://lattes.cnpq.br/) do CNPq constitui a principal base de dados de currículos acadêmicos do Brasil, reunindo informações detalhadas sobre a trajetória profissional, produção científica, formação acadêmica e áreas de atuação de pesquisadores de todo o país. Embora esses dados estejam publicamente disponíveis na plataforma, seu formato de apresentação — voltado à consulta individual — limita a capacidade de análise transversal e a identificação de padrões que emergem apenas quando múltiplos perfis são analisados em conjunto.

O **smartLattes** nasce da premissa de que esses dados públicos, quando estruturados em uma base de dados adequada e submetidos a técnicas de Inteligência Artificial, podem revelar informações de alto valor estratégico para a comunidade científica. Entre as possibilidades que se abrem, destacam-se:

- **Identificação de redes de pesquisa**: a partir da análise das competências, áreas de atuação e produção científica dos pesquisadores, a IA pode sugerir conexões entre profissionais com interesses complementares, favorecendo a formação de grupos de pesquisa interdisciplinares.

- **Sugestão de produtos de pesquisa colaborativos**: com base em competências complementares identificadas nos perfis, a IA pode propor projetos, artigos ou iniciativas que combinem habilidades distintas para abordar problemas complexos.

- **Detecção de lacunas de pesquisa**: a análise agregada dos perfis permite identificar áreas do conhecimento com menor cobertura ou com potencial inexplorado, orientando investimentos e esforços de pesquisa para onde são mais necessários.

- **Mapeamento de competências institucionais**: instituições de pesquisa podem compreender melhor o conjunto de habilidades disponíveis em seus quadros, facilitando a alocação de recursos e a definição de estratégias de desenvolvimento.

## Princípios de Privacidade e Voluntariedade

O smartLattes opera sob dois princípios fundamentais:

1. **Voluntariedade**: os pesquisadores enviam seus currículos por iniciativa própria, mediante upload do arquivo XML exportado diretamente da Plataforma Lattes. Nenhum dado é coletado de forma automatizada ou sem o consentimento explícito do pesquisador.

2. **Dados já públicos**: o arquivo XML exportado pela Plataforma Lattes contém exclusivamente informações que o próprio pesquisador tornou públicas em seu currículo. O smartLattes não solicita, armazena ou processa dados que não estejam já disponíveis publicamente na plataforma.

## Arquitetura

O projeto segue o modelo de contextos do [C4 Model](https://c4model.com/), organizando-se em dois contextos funcionais:

### Contexto de Aquisição (atual)

Responsável pela ingestão dos dados. O pesquisador acessa a interface web, faz upload do arquivo XML de seu currículo Lattes, e o sistema converte o XML para uma estrutura JSON que é armazenada no banco de dados MongoDB. Este é o contexto atualmente implementado.

### Contexto de Apresentação (futuro)

Responsável pela visualização e análise dos dados armazenados. Este contexto permitirá a exploração dos perfis, a visualização de redes de colaboração sugeridas pela IA, e a consulta a produtos de pesquisa recomendados. Está previsto para implementação futura.

## Stack Tecnológico

| Componente | Tecnologia | Justificativa |
|------------|------------|---------------|
| **Backend** | Go 1.23 | Linguagem compilada que produz binários estáticos, resultando em imagens Docker extremamente leves (~25-35 MB). Biblioteca padrão robusta para HTTP e XML. |
| **Frontend** | HTML/CSS/JS (vanilla) | Embutido no binário via `go:embed`. Sem dependências externas, sem etapa de build frontend, sem framework. |
| **Banco de Dados** | MongoDB | Modelo de documentos flexível, ideal para armazenar a estrutura hierárquica e variável dos currículos Lattes sem necessidade de schema rígido. |
| **Containerização** | Docker (Alpine) | Imagem multi-stage com base Alpine (~7 MB), publicada em `ghcr.io/edalcin/smartlattes`. |
| **Deploy** | Unraid | Container gerenciado via interface web do Unraid, sem necessidade de orquestração. |

## Como Funciona

```
Pesquisador                    smartLattes                     MongoDB
    |                              |                              |
    |  1. Exporta XML do Lattes    |                              |
    |  2. Upload via interface web |                              |
    |----------------------------->|                              |
    |                              |  3. Decodifica ISO-8859-1    |
    |                              |  4. Valida XML Lattes        |
    |                              |  5. Converte XML → JSON      |
    |                              |  6. Upsert no MongoDB        |
    |                              |----------------------------->|
    |                              |  7. Confirmação              |
    |                              |<-----------------------------|
    |  8. Exibe resumo do CV       |                              |
    |<-----------------------------|                              |
```

1. O pesquisador exporta seu currículo da Plataforma Lattes em formato XML
2. Acessa a interface web do smartLattes e faz upload do arquivo
3. O sistema decodifica o arquivo (ISO-8859-1, padrão do Lattes)
4. Valida a estrutura XML (elemento raiz `CURRICULO-VITAE`, atributo `NUMERO-IDENTIFICADOR`)
5. Converte recursivamente toda a árvore XML para uma estrutura JSON genérica
6. Armazena o documento no MongoDB usando o `NUMERO-IDENTIFICADOR` como chave única (upsert)
7. Recebe confirmação do banco de dados
8. Exibe ao pesquisador um resumo com nome, ID Lattes, data de atualização e contagens de produção

## Estrutura do Projeto

```
smartLattes/
├── cmd/smartlattes/main.go       # Ponto de entrada da aplicação
├── internal/
│   ├── handler/                  # Handlers HTTP (upload, health, páginas)
│   ├── parser/                   # Parser XML → JSON (genérico, recursivo)
│   ├── store/                    # Cliente MongoDB (conexão, upsert, ping)
│   └── static/                   # Arquivos estáticos (HTML, CSS, JS)
├── docs/                         # Arquivo XML de exemplo para testes
├── specs/                        # Especificações e artefatos de design
├── Dockerfile                    # Build multi-stage
├── go.mod                        # Dependências Go
├── .env.example                  # Template de variáveis de ambiente
└── README.md
```

## Variáveis de Ambiente

| Variável | Obrigatória | Padrão | Descrição |
|----------|-------------|--------|-----------|
| `MONGODB_URI` | Sim | — | String de conexão MongoDB |
| `MONGODB_DATABASE` | Não | `smartLattes` | Nome do banco de dados |
| `PORT` | Não | `8080` | Porta do servidor HTTP |
| `MAX_UPLOAD_SIZE` | Não | `10485760` | Tamanho máximo de upload em bytes (10 MB) |

## Deploy

### Docker

```bash
# Build da imagem
docker build -t ghcr.io/edalcin/smartlattes:latest .

# Execução
docker run -p 8080:8080 \
  -e MONGODB_URI="mongodb://user:password@hostname:27017/?authSource=dbname" \
  ghcr.io/edalcin/smartlattes:latest
```

### Unraid

Instruções detalhadas para deploy via interface web do Unraid estão disponíveis em [`specs/quickstart.md`](specs/quickstart.md).

## Licença

Este projeto é de código aberto. Consulte o arquivo de licença para mais detalhes.
