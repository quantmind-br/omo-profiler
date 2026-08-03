# Caça a bugs — omo-profiler

Alvo: `/home/diogo/dev/omo-profiler` @ `b117b82` **+ árvore de trabalho** (122 arquivos
não commitados aplicados via `BUGHUNT_DIRTY=aplicar`, mais os 72 arquivos não rastreados
copiados para o clone confinado com `sbx_boot`).

A cobertura do estado commitado sozinho não faria sentido: `internal/web/`,
`internal/config/document.go` e `internal/config/jsonc.go` são arquivos **não rastreados**
que o `AGENTS.md` já descreve como arquitetura corrente. O `HEAD` está defasado; a caça
cobre o que o usuário realmente tem em disco.

Toolchain: `go1.26.5-X:nodwarf5 linux/amd64`, `golangci-lint 2.12.2`.
Sem CI no repositório — a linha de base é `Makefile` + `gofmt`/`go vet`/`golangci-lint`.

---

### BUG-01 — o módulo não compila: `frontend/dist/.gitkeep` não existe

- **Severidade:** crítica
- **Módulo:** `internal/web`
- **Invariante violado:** I1 (`AGENTS.md`: "`make build` stays Node-free via `dist/.gitkeep`"; `embed.go`: "a committed frontend/dist/.gitkeep makes this compile before the UI is built")
- **Oráculo que detectou:** O3 (build quebra)
- **Iteração:** 1 (falha já presente na linha de base)

**Reprodução mínima**

```
go build ./...
```

**Comportamento observado:** o pacote `internal/web` e tudo que o importa
(`cmd/omo-profiler`, `internal/cli`, `internal/cli/cmd`) falham em *setup*. O binário não
é construído e 4 pacotes rodam **zero** testes.

**Comportamento esperado:** compilar sem Node. O `.gitignore` já reserva a exceção
(`internal/web/frontend/dist/*` + `!internal/web/frontend/dist/.gitkeep`) — o arquivo
simplesmente nunca foi criado, então só compila em máquina que já rodou `make web-build`.

**Causa raiz:** `internal/web/embed.go:8` — `//go:embed all:frontend/dist` exige ao menos
um arquivo correspondente, e `frontend/dist/` está inteiramente ignorada pelo Git sem o
placeholder que a própria diretiva documenta.

**Correção:** criado `internal/web/frontend/dist/.gitkeep` (vazio). Nenhuma mudança de
código — o `.gitignore` e o `embed.go` já estavam corretos.

**Teste de regressão:** o próprio `go build ./...` (falha antes / passa depois). Um teste
Go não serve aqui: sem a correção o pacote não compila, então nenhum teste chega a rodar —
a falha de compilação é o estado vermelho.

**Evidência:**

```
# antes
$ go build ./...
internal/web/embed.go:8:12: pattern all:frontend/dist: no matching files found

$ go test ./...
FAIL	github.com/diogenes/omo-profiler/cmd/omo-profiler [setup failed]
FAIL	github.com/diogenes/omo-profiler/internal/cli [setup failed]
FAIL	github.com/diogenes/omo-profiler/internal/cli/cmd [setup failed]
FAIL	github.com/diogenes/omo-profiler/internal/web [setup failed]

# depois
$ mkdir -p internal/web/frontend/dist && : > internal/web/frontend/dist/.gitkeep
$ go build ./... && echo BUILD_OK
BUILD_OK
```

---

### BUG-02 — os backups nunca são rotacionados: um arquivo por escrita, para sempre

- **Severidade:** média
- **Módulo:** `internal/backup`
- **Invariante violado:** I8 (`Clean` documenta "removes old backups, keeping only the N most recent" — mas nada a chama)
- **Oráculo que detectou:** O6 (vazamento de recurso: disco)
- **Iteração:** 4

**Reprodução mínima**

```go
for range 40 { backup.CreateOmoIfPresent() }
list, _ := backup.List()   // len(list) == 40
```

**Comportamento observado:** 40 escritas deixam 40 arquivos `omo.json.bak.*` em `~/.omo/`.
`Clean`, `List` e `Restore` são exportadas e **não têm nenhum chamador** em todo o
repositório (`grep -rn` em `--include='*.go'` fora dos testes do próprio pacote): a
rotação existe como código e nunca executa.

**Comportamento esperado:** um limite. Os 8 pontos de mutação (`profile.Save`,
`Create`, `CreateFrom`, `CreateAvailable`, `Delete`, `Rename`, `UpdateOpenCodeBlock`,
`SaveOpenCodeBlock`, mais o wizard da TUI) passam todos por `CreateOmoIfPresent`, então
cada edição no wizard/web deixa mais uma cópia integral de um documento que pode conter
tokens.

**Causa raiz:** `internal/backup/backup.go:211` — `CreateOmoIfPresent` chama `Create` e
retorna. A política "faça backup antes de escrever" tem um ponto único de aplicação; a
política "não guarde infinitos backups" não tinha nenhum.

**Correção:** rotação no mesmo ponto de estrangulamento — `CreateOmoIfPresent` passa a
chamar `Clean(KeepLast)` após um snapshot bem-sucedido, com `KeepLast = 10`. O erro de
`Clean` é descartado de propósito: o snapshot, que é o que o chamador depende, já teve
êxito; reprovar a escrita porque cópias velhas não puderam ser podadas trocaria uma
capacidade real por uma questão de faxina.

**Teste de regressão:** `TestHuntBackupsAreRotated` em
`internal/backup/hunt_probe_test.go`

**Evidência:**

```
# antes
--- FAIL: TestHuntBackupsAreRotated (0.00s)
    Error:      "40" is not less than "40"
    Messages:   40 writes left 40 backups — rotation never runs

# depois
ok  	github.com/diogenes/omo-profiler/internal/backup	1.017s
```

---

### BUG-03 — `backup.Clean` com `keepLast` negativo entra em pânico

- **Severidade:** baixa (latente hoje; vira alcançável assim que a rotação do BUG-02 é ligada)
- **Módulo:** `internal/backup`
- **Invariante violado:** I8 (a API pública de rotação não deve derrubar o processo)
- **Oráculo que detectou:** O1 (crash)
- **Iteração:** 4

**Reprodução mínima**

```go
backup.Clean(-1)   // panic: runtime error: index out of range [-1]
```

**Comportamento observado:** `runtime error: index out of range [-1]` em
`backup.go:198`.

**Comportamento esperado:** `keepLast` negativo significa "não guarde nenhum" — remover
tudo, sem pânico.

**Causa raiz:** `internal/backup/backup.go:197` — o laço começa em `i := keepLast` sem
piso; a guarda anterior (`len(backups) <= keepLast`) nunca dispara para valores
negativos, então `backups[-1]` é indexado.

**Correção:** `keepLast = max(keepLast, 0)` no início de `Clean`.

**Teste de regressão:** `TestHuntCleanNonPositiveKeep` em
`internal/backup/hunt_probe_test.go`

**Evidência:**

```
# antes
--- FAIL: TestHuntCleanNonPositiveKeep (0.00s)
    Panic value:	runtime error: index out of range [-1]
    Panic stack:
      github.com/diogenes/omo-profiler/internal/backup.Clean(0xffffffffffffffff)
        internal/backup/backup.go:198 +0x25d
    Messages:   Clean(-1) panicked

# depois
ok  	github.com/diogenes/omo-profiler/internal/backup	1.017s
```

---

### BUG-04 — `backup.Restore` alarga 0600 → 0644 e escreve sem atomicidade

- **Severidade:** média
- **Módulo:** `internal/backup`
- **Invariante violado:** I2 (`config.WriteFileAtomic`: "this document can hold secrets (bot tokens), so a user who tightened it to 0600 must not have it silently widened") e I3 (`Document.Save`: "a plain truncating write that fails midway would take the whole set with it")
- **Oráculo que detectou:** I2 (invariante do projeto contradita pelo código)
- **Iteração:** 4

**Reprodução mínima**

```go
os.WriteFile(config.OmoFile(), doc, 0o600)
bak, _ := backup.Create(config.OmoFile())   // backup nasce 0600, correto
os.Remove(config.OmoFile())
backup.Restore(bak)
// modo do arquivo restaurado: 0644
```

**Comportamento observado:** `-rw-r--r--`. O documento restaurado fica legível por
qualquer usuário da máquina.

**Comportamento esperado:** `-rw-------`. Todo outro caminho de escrita do repositório
preserva/estreita o modo, e o `Create` deste mesmo arquivo faz questão de herdar 0600 com
um comentário explicando por quê.

**Causa raiz:** `internal/backup/backup.go:171` — `os.WriteFile(configPath, data, 0644)`.
Duas falhas na mesma linha: o literal `0644` (que só se aplica na criação, ou seja
exatamente no cenário de recuperação após perda do arquivo) e o fato de ser uma escrita
truncante — a recuperação podia destruir o documento que estava recuperando.

**Correção:** `config.WriteFileAtomic(config.OmoFile(), data, 0o600)`. Resolve as duas
falhas reusando o helper que já existe: preserva o modo de um arquivo existente, cria
novos em 0600, e escreve via arquivo temporário + `rename`.

**Teste de regressão:** `TestHuntRestorePreservesMode` em
`internal/backup/hunt_probe_test.go`

**Evidência:**

```
# antes
--- FAIL: TestHuntRestorePreservesMode (0.00s)
    Error:      Not equal:
                expected: 0x180      (0600)
                actual  : 0x1a4      (0644)
    Messages:   Restore widened a secrets-bearing document to -rw-r--r--

# depois
ok  	github.com/diogenes/omo-profiler/internal/backup	1.017s
```

---

### BUG-05 — CSRF: qualquer página aberta pelo usuário escreve no `omo.json` dele

- **Severidade:** alta
- **Módulo:** `internal/web`
- **Invariante violado:** I-web (só a SPA servida pelo próprio processo deve poder mutar o documento)
- **Oráculo que detectou:** I-web (invariante de segurança) + reprodução executável
- **Iteração:** 4

**Reprodução mínima**

Com `omo-profiler web` rodando (padrão `127.0.0.1:4747`), qualquer página em qualquer
origem executa:

```js
fetch('http://127.0.0.1:4747/api/profiles', {
  method: 'POST',
  headers: {'Content-Type': 'text/plain;charset=UTF-8'},  // requisição "simples": sem preflight
  body: '{"name":"attacker"}'
})
```

Equivalente em curl:

```
curl -X POST http://127.0.0.1:4747/api/profiles \
  -H 'Origin: https://evil.example' -H 'Content-Type: text/plain' \
  -d '{"name":"attacker"}'
```

**Comportamento observado:** `201 Created` — o perfil é criado no `~/.omo/omo.json` do
usuário. O mesmo vale para `PUT /api/profiles/{name}` (sobrescreve um perfil),
`DELETE /api/profiles/{name}`, `POST /api/import` e `POST /api/profiles/{name}/activate`.
A resposta é ilegível cross-origin pela SOP, mas a **escrita já aconteceu**.

**Comportamento esperado:** `403`. O servidor não tem autenticação — ele depende de estar
em loopback —, e loopback não protege contra o navegador do próprio usuário. Um corpo
`text/plain` transforma o POST em requisição CORS "simples", que é entregue sem preflight.

**Causa raiz:** `internal/web/server.go` — `newMux()` registrava as rotas sem nenhum
middleware; nenhum handler checa `Origin`, `Content-Type` ou token.

**Correção:** middleware `sameOriginOnly` envolvendo o roteador. Para métodos que não são
GET/HEAD/OPTIONS, um cabeçalho `Origin` presente precisa casar com o `Host` da requisição.
`Origin` é ausente em curl e em GET same-origin, e os navegadores sempre o enviam em
requisições cross-origin e em não-GET same-origin — então a comparação separa a SPA de uma
página hostil sem exigir infraestrutura de token. `newMux()` passou a devolver
`http.Handler` (com a guarda), e `newRoutes()` mantém o registro das rotas, de modo que os
testes existentes exercitam exatamente o que roda em produção.

**Teste de regressão:** `TestHuntCrossOriginWriteIsRejected` e
`TestHuntCrossOriginImportIsRejected` em `internal/web/hunt_probe_test.go`

**Evidência:**

```
# antes (teste)
--- FAIL: TestHuntCrossOriginWriteIsRejected (0.00s)
    Error:      Should not be: 201
    Messages:   cross-origin POST created a profile (body={"name":"attacker"})
--- FAIL: TestHuntCrossOriginImportIsRejected (0.10s)
    Error:      Should not be: 200
    Messages:   cross-origin POST /api/import succeeded (body={"hadCollision":false,"name":"pwned"})

# depois — servidor real, binário construído, curl contra 127.0.0.1:4748
GET  /api/profiles            same-origin   -> status=200
POST /api/profiles            SPA origin    -> status=201
POST /api/profiles            no origin     -> status=201
POST /api/profiles            evil origin   -> status=403
PUT  /api/profiles/dev        evil origin   -> status=403
DELETE /api/profiles/dev      evil origin   -> status=403
POST /api/import              evil origin   -> status=403
GET  /                        SPA fallback  -> status=200
profiles now: "dev" "fromcurl" "fromspa"
```

O bloco acima é a prova dupla que importa: a SPA (`Origin` == `Host`) e o curl (sem
`Origin`) continuam escrevendo — `fromspa` e `fromcurl` foram criados —, enquanto os
quatro ataques cross-origin são recusados e `dev` sobrevive intacto.

---

### BUG-06 — erro de execução imprime o bloco de uso e enterra a mensagem real

- **Severidade:** baixa
- **Módulo:** `internal/cli`
- **Invariante violado:** I-cli (contrato público: falha de execução ≠ erro de uso)
- **Oráculo que detectou:** O5 (o erro real é ofuscado) via uso real do binário
- **Iteração:** 6

**Reprodução mínima**

```
omo-profiler web --port 1 --no-open       # porta privilegiada
omo-profiler list                          # com ~/.omo/omo.json corrompido
```

**Comportamento observado:** a mensagem de erro correta é seguida do `Usage:` completo do
subcomando, sugerindo que o usuário errou a sintaxe quando na verdade a porta estava
indisponível ou o arquivo de configuração estava quebrado.

**Comportamento esperado:** só a mensagem de erro. O comando fez o parse corretamente; a
falha é de runtime.

**Causa raiz:** `internal/cli/root.go:16` — o comando raiz não define `SilenceUsage`, então
o Cobra imprime o *usage* para qualquer erro devolvido por um `RunE`. Afeta `list` e `web`
(os únicos que usam `RunE`); os demais usam `Run` + `os.Exit` e por isso escapavam,
deixando o comportamento inconsistente entre subcomandos.

**Correção:** `SilenceUsage: true` no `rootCmd` (o Cobra propaga aos filhos). O Cobra
continua imprimindo `Error: ...` e o código de saída segue 1.

**Teste de regressão:** `TestHuntRuntimeErrorDoesNotPrintUsage` em
`internal/cli/root_hunt_test.go` (primeiro teste do pacote `internal/cli`, que estava em
0% de cobertura)

**Evidência:**

```
# antes
$ omo-profiler web --port 1 --no-open
Error: failed to listen on 127.0.0.1:1: listen tcp 127.0.0.1:1: bind: permission denied
Usage:
  omo-profiler web [flags]

Flags:
  -h, --help          help for web
      --host string   Host to bind (default "127.0.0.1")
      --no-open       Do not open the browser automatically
      --port int      Port to listen on (default 4747)

-> exit=1

# depois
$ omo-profiler web --port 1 --no-open
Error: failed to listen on 127.0.0.1:1: listen tcp 127.0.0.1:1: bind: permission denied
-> exit=1

$ omo-profiler list      # documento corrompido
Error: failed to list profiles: parse .../.omo/omo.json: invalid character 'o' in literal null (expecting 'u')
-> exit=1
```

---

## Resumo

| | |
|---|---|
| Iterações do laço | 7 |
| Alvos cobertos | `internal/config` (JSONC/documento), `internal/backup`, `internal/web`, `internal/cli` + `internal/cli/cmd`, `internal/profile` (via round-trip) |
| Bugs confirmados, corrigidos e com teste verde | 6 |
| Por severidade | 1 crítica, 1 alta, 3 médias, 1 baixa |
| Motivo da parada | Saturação: os dois alvos de maior valor restantes (parser JSONC e round-trip export/import) resistiram a 31,8 M execuções de fuzzing sem uma única falha, e o laço não produziu achado novo depois da iteração 6 |
| Orçamento usado | ~7 iterações de 40; ~45 min de 90 |

Um teste que **passou** merece registro tanto quanto os que falharam, porque cada um
fechou uma hipótese: as escritas concorrentes de 8 perfis simultâneos preservaram todas as
8 (o `docMutex` faz o que promete), o round-trip export→import é de fato lossless
inclusive para chaves desconhecidas e zeros explícitos, e um documento criado do zero
nasce `0600`.

## Suspeitas não confirmadas (hipóteses, sem reprodução de dano)

- **S1 — `StripJSONC` não é idempotente em entrada já inválida.** `StripJSONC(",,}")`
  devolve `", }"`; uma segunda passada devolve `"  }"`. Sem impacto: as três formas são
  JSON inválido e `LoadDocumentFrom` rejeita todas. Em JSONC válido a vírgula não-final é
  sempre seguida por um início de valor, nunca por `}`/`]`, então nenhuma vírgula legítima
  é apagada — confirmado por 31,8 M execuções do oráculo metamórfico. O caso ficou no
  corpus (`internal/config/testdata/fuzz/`).
- **S2 — DNS rebinding contorna a correção do BUG-05.** Se o atacante faz `evil.com`
  resolver para `127.0.0.1`, `Origin: http://evil.com:4747` casa com
  `Host: evil.com:4747` e a guarda deixa passar. Fechar isso exige validar `Host` contra
  uma allowlist de loopback, o que quebraria quem usa `--host 0.0.0.0` para acessar pela
  LAN. Decisão de projeto — não tomei sozinho.
- **S3 — `models.Load` sobrescreve sempre o mesmo `models.json.bak`.**
  (`internal/models/models.go:76`) Uma segunda corrupção destrói o backup da primeira, e
  o registro vazio devolvido é persistido no próximo `Save`. O comportamento é deliberado
  e documentado no código; classifiquei como escolha de projeto, não defeito.
- **S4 — `readBody` usa `io.ReadAll` sem limite** (`internal/web/handlers.go:583`). Um
  corpo grande é lido inteiro em memória. Servidor de loopback, então o atacante realista
  é o próprio usuário; não gerei reprodução de dano.
- **S5 — um documento literalmente `null` é tratado como vazio em silêncio.**
  `omo-profiler list` imprime `(No profiles found)` em vez de sinalizar arquivo
  malformado.
- **Não é bug:** `omo-profiler create <nome>` sem `--from` responde "TUI mode not
  implemented". É um stub declarado, e o próprio texto longo do comando manda usar
  `--from`. Funcionalidade ausente, não defeito.

## Alvos de fuzzing criados

Em `internal/config/hunt_fuzz_test.go`:

```
go test -run xxx -fuzz FuzzStripJSONCInvariants   -fuzztime 60s ./internal/config/
go test -run xxx -fuzz FuzzJSONCCommentInjection -fuzztime 60s ./internal/config/
go test -run xxx -fuzz FuzzDocumentRoundTrip     -fuzztime 60s ./internal/config/
```

- `FuzzStripJSONCInvariants` — preservação de comprimento e de quebras de linha
  (as duas coisas que o comentário de `StripJSONC` promete para manter offsets e números
  de linha de erro corretos), idempotência sobre o domínio útil, e "JSON válido não pode
  ser alterado".
- `FuzzJSONCCommentInjection` — o oráculo metamórfico que vale: pega qualquer JSON
  válido, injeta comentários de linha, comentários de bloco e vírgulas finais nas posições
  de espaço em branco, e exige que o resultado decodifique para exatamente o mesmo valor.
  31,8 M execuções, nenhuma falha.
- `FuzzDocumentRoundTrip` — a forma canônica de `Document.Bytes()` reparseia e é estável
  na segunda serialização.

Corpus da falha do S1: `internal/config/testdata/fuzz/FuzzStripJSONCInvariants/36579a76c7cf530c`.

## Suíte final

```
### gofmt : gofmt -l .        RC=0
### build : go build ./...    RC=0
### vet   : go vet ./...      RC=0
### test1 : go test ./...     RC=0
### test2 : go test -count=1 ./...        RC=0
### race  : go test -race -count=1 ./...  RC=0
### cover : go test -count=1 -cover ./... RC=0
### lint  : golangci-lint run ./...       RC=0   (0 issues)
```

## Delta de cobertura por módulo

| Pacote | Antes | Depois |
|---|---|---|
| `internal/cli` | 0.0% | **69.2%** |
| `internal/web` | 45.6% | **51.8%** |
| `internal/backup` | 65.8% | **71.6%** |
| `cmd/omo-profiler` | 0.0% | 0.0% |
| `internal/cli/cmd` | 4.8% | 4.8% |
| `internal/config` | 82.6% | 82.6% |
| `internal/diff` | 100.0% | 100.0% |
| `internal/models` | 80.7% | 80.7% |
| `internal/profile` | 79.5% | 79.5% |
| `internal/schema` | 67.4% | 67.4% |
| `internal/tui` | 41.2% | 41.2% |
| `internal/tui/layout` | 40.8% | 40.8% |
| `internal/tui/views` | 60.3% | 60.3% |

"Antes" é a linha de base **após** o BUG-01, quando os 4 pacotes bloqueados voltaram a
compilar. Antes disso `cmd/omo-profiler`, `internal/cli`, `internal/cli/cmd` e
`internal/web` não rodavam teste nenhum.

## Como aplicar

As correções foram feitas e validadas **dentro do sandbox**, não na sua árvore. Nada em
`/home/diogo/dev/omo-profiler` foi modificado. Para aplicar:

```bash
cd /home/diogo/dev/omo-profiler
git apply --check .bughunt/bughunt.patch   # confere antes
git apply .bughunt/bughunt.patch           # internal/backup/backup.go, internal/web/server.go, internal/cli/root.go
cp -r .bughunt/novos/. .                   # arquivos novos (inclui frontend/dist/.gitkeep)
make build && make test && golangci-lint run ./...
```

O patch é um `diff -u` contra o conteúdo atual dos seus arquivos (não contra o `HEAD`),
porque `internal/web/server.go` ainda é um arquivo não rastreado.

`go.sum` **não** entra: o `go mod download all` do pré-aquecimento acrescentou hashes de
dependências transitivas de teste. Confirmei que as correções compilam e passam com o seu
`go.sum` original — as correções não adicionam dependência nenhuma (só `net/url`, da
biblioteca padrão).

Evidência bruta de todas as fases em `.bughunt/evidencia/`.

`internal/cli/cmd` continua em 4.8% e é o maior buraco restante: os subcomandos chamam
`os.Exit` direto de dentro de `Run`, o que impede testá-los em processo. Foi por isso que
a caça nesse alvo foi feita conduzindo o binário de verdade (`$SB/cli-hunt.sh`), e foi
assim que o BUG-06 apareceu. Tornar esses `Run` testáveis (devolver erro em vez de sair)
é o próximo passo óbvio, mas é refatoração, não correção de bug — fora do escopo desta
caça.
