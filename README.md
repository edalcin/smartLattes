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

O projeto segue o modelo de contextos do [C4 Model](https://c4model.com/), organizando-se em quatro contextos funcionais:

### Contexto de Aquisição

Responsável pela ingestão dos dados. O pesquisador acessa a interface web, faz upload do arquivo XML de seu currículo Lattes, e o sistema converte o XML para uma estrutura JSON que é armazenada no banco de dados MongoDB.

### Contexto de Transformação

Responsável pela geração de resumos analíticos de pesquisadores utilizando Inteligência Artificial. O sistema integra-se com três provedores de IA (OpenAI, Anthropic e Google Gemini) via chamadas REST diretas, sem SDKs. O usuário fornece sua própria chave de API (nunca armazenada), seleciona o modelo desejado, e o sistema gera um resumo estruturado contendo:

- **Perfil e Principais Características** — trajetória acadêmica e pontos fortes
- **Áreas de Especialidade** — lista hierárquica de áreas de atuação
- **Potencial de Contribuição Científica** — relevância e impacto potencial
- **Coautores Mais Frequentes** — tabela com os 15 coautores mais frequentes
- **Produção Quantificada por Área** — tabela com produção por área de especialidade

Os resumos gerados são armazenados no MongoDB e podem ser baixados em formato Markdown (.md).

### Contexto de Análise

Responsável pela identificação de redes de pesquisa e oportunidades de colaboração entre pesquisadores. A partir dos currículos armazenados, o sistema utiliza IA para comparar o perfil de um pesquisador com os demais da base, gerando um relatório estruturado contendo:

- **Áreas de Sinergia** — identificação de convergências temáticas entre pesquisadores
- **Potenciais Colaborações** — sugestões concretas de projetos, artigos e iniciativas conjuntas
- **Complementaridade de Competências** — mapeamento de habilidades complementares
- **Oportunidades de Pesquisa** — lacunas e fronteiras de conhecimento identificadas

A análise é oferecida como fluxo opcional após a geração do resumo (nas páginas "Enviar CV" e "Gerar Resumo") e também como página dedicada ("Analisar Relações"). Os relatórios são armazenados na coleção `relacoes` do MongoDB e podem ser baixados em formato Markdown.

O sistema aplica uma estratégia progressiva de truncamento para respeitar os limites de tokens dos modelos de IA, priorizando os dados do pesquisador atual e reduzindo progressivamente os dados dos demais currículos.

### Contexto de Apresentação (futuro)

Responsável pela visualização e exploração dos dados armazenados. Este contexto permitirá a navegação interativa pelos perfis, a visualização gráfica de redes de colaboração, e a consulta a produtos de pesquisa recomendados. Está previsto para implementação futura.

## Stack Tecnológico

| Componente | Tecnologia | Justificativa |
|------------|------------|---------------|
| **Backend** | Go 1.23 | Linguagem compilada que produz binários estáticos, resultando em imagens Docker extremamente leves (~25-35 MB). Biblioteca padrão robusta para HTTP e XML. |
| **Frontend** | HTML/CSS/JS (vanilla) | Embutido no binário via `go:embed`. Sem dependências externas, sem etapa de build frontend, sem framework. |
| **Banco de Dados** | MongoDB | Modelo de documentos flexível, ideal para armazenar a estrutura hierárquica e variável dos currículos Lattes sem necessidade de schema rígido. |
| **IA** | OpenAI, Anthropic, Gemini | Integração via REST API direto (`net/http`), sem SDKs. Chaves de API fornecidas pelo usuário a cada uso. |
| **Containerização** | Docker (Alpine) | Imagem multi-stage com base Alpine (~7 MB), publicada em `ghcr.io/edalcin/smartlattes`. |
| **Deploy** | Unraid | Container gerenciado via interface web do Unraid, sem necessidade de orquestração. |

## Como Funciona

### Upload de CV

1. O pesquisador exporta seu currículo da Plataforma Lattes em formato XML
2. Acessa a interface web do smartLattes e faz upload do arquivo
3. O sistema decodifica o arquivo (ISO-8859-1, padrão do Lattes)
4. Valida a estrutura XML (elemento raiz `CURRICULO-VITAE`, atributo `NUMERO-IDENTIFICADOR`)
5. Converte recursivamente toda a árvore XML para uma estrutura JSON genérica (chaves em minúsculas)
6. Armazena o documento no MongoDB usando o `NUMERO-IDENTIFICADOR` como chave única (upsert)
7. Exibe ao pesquisador um resumo com nome, ID Lattes, data de atualização e contagens de produção

### Geração de Resumo por IA

1. O usuário busca um currículo por nome ou ID Lattes (via página "Gerar Resumo" ou após upload)
2. Seleciona o provedor de IA (OpenAI, Anthropic ou Google Gemini)
3. Fornece sua chave de API (transiente, nunca armazenada)
4. Clica em "Carregar Modelos" para listar os modelos disponíveis
5. Seleciona o modelo e clica em "Gerar Resumo"
6. O sistema envia os dados do CV ao provedor de IA com um prompt estruturado
7. O resumo gerado é exibido na tela e pode ser baixado em formato Markdown
8. Ao confirmar, o resumo é salvo na coleção `resumos` do MongoDB

### Análise de Relações entre Pesquisadores

1. Após gerar um resumo (ou via página dedicada "Analisar Relações"), o usuário pode iniciar a análise
2. O sistema recupera os dados de todos os currículos armazenados na base
3. Aplica truncamento progressivo para caber nos limites de tokens do modelo (preservando integralmente o CV atual)
4. Envia os dados ao provedor de IA com um prompt especializado em identificação de redes
5. O relatório gerado é exibido na tela e pode ser baixado em formato Markdown
6. Ao confirmar, o relatório é salvo na coleção `relacoes` do MongoDB
7. Caso haja apenas um pesquisador na base, o sistema informa que não há outros perfis para comparação (HTTP 409)

## Estrutura do Projeto

```
smartLattes/
├── cmd/smartlattes/
│   ├── main.go                  # Ponto de entrada, rotas, graceful shutdown
│   ├── resumoPrompt.md          # Prompt de IA para geração de resumos
│   └── analisePrompt.md         # Prompt de IA para análise de relações
├── internal/
│   ├── handler/                 # Handlers HTTP (upload, search, models, summary, analysis, download, health)
│   ├── parser/                  # Parser XML → JSON (genérico, recursivo)
│   ├── store/                   # Cliente MongoDB (curriculos + resumos + relacoes)
│   ├── ai/                      # Provedores de IA (OpenAI, Anthropic, Gemini) + truncamento
│   ├── export/                  # Exportação de documentos (Markdown)
│   └── static/                  # Arquivos estáticos (HTML, CSS, JS)
├── docs/                        # Arquivo XML de exemplo para testes
├── specs/                       # Especificações e artefatos de design
├── Dockerfile                   # Build multi-stage
├── go.mod                       # Dependências Go
├── .env.example                 # Template de variáveis de ambiente
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
