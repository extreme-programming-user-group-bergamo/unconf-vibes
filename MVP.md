# UNCONF CLI

> "Choose your room, choose your roommate, focus on the conference."

**Documento di Specifiche Funzionali & CLI Design**

---

## 1. Il Problema

Le conferenze Open Space per sviluppatori software come "Polenta e Deploy" o "SoCraTes Italia" richiedono un sistema di registrazione che vada oltre la semplice raccolta di partecipanti. La scelta dell'alloggio è parte integrante dell'esperienza: i partecipanti vogliono vedere chi c'è, scegliere con chi condividere la camera, e completare la registrazione in modo trasparente e immediato.

Attualmente, questo processo richiede scambi multipli di email, fogli Excel condivisi, e coordinamento manuale con l'hotel: friction per i partecipanti e overhead organizzativo per il team.

---

## 2. La Soluzione

Una applicazione da riga di comando (CLI) dedicata con una interfaccia utente testuale (TUI) che trasforma la registrazione alla conferenza in un'esperienza fluida e social, automatizzando la comunicazione con l'hotel partner.

L'applicazione è molto "nerd", pensata per divertire e coinvolgere sviluppatori software, ma è anche estremamente pratica e funzionale.

**Core Value Proposition:**

1. **Velocità**: registrazione completa in meno di 3 minuti
2. **Social Discovery**: visualizzazione grafica (TUI) delle stanze e dei partecipanti
3. **Automazione**: comunicazione strutturata e automatica verso le strutture alberghiere

---

## 3. Funzionalità

### Per i Partecipanti

- Lista pubblica delle conferenze disponibili
- Visualizzazione dettagli conferenza (date, location, costi)
- Flusso di creazione account guidato:
  - Dati personali (nome, email)
  - Necessità alimentari speciali
  - Necessità di accessibilità
- Flusso di registrazione guidato:
  - Selezione date di arrivo e partenza
  - Visualizzazione real-time delle camere disponibili (tipologia, prezzo, caratteristiche, occupazione)
  - Visualizzazione di chi ha già prenotato ogni camera (se ha dato consenso)
  - **Privacy selettiva**: ogni partecipante decide se rendere pubblico il proprio nome
  - Conferma finale con riepilogo dettagliato, inviato anche via email

### Per l'Hotel

- Notifica email automatica alla prenotazione con dati strutturati:
  - Dati del partecipante
  - Tipologia camera selezionata
  - Date di arrivo/partenza
  - Note speciali
- Template email personalizzabile

### Per gli Organizzatori

- Dashboard amministrativa per monitorare registrazioni
- Configurazione conferenza (date, camere disponibili, prezzi)
- Gestione template email
- Notifiche via email per ogni nuova registrazione
- Esportazione dati registrazioni in formato CSV

---

## 4. Architettura dei Comandi

L'interfaccia segue lo standard **Unix-like**. La sintassi generale è:

```bash
unconf <comando> [sotto-comando] [opzioni]
```

### Concetto di "Context"

Per evitare ripetizioni, UNCONF utilizza un sistema di **contesto** (simile a `kubectl` o `git`). L'utente seleziona una conferenza attiva una volta sola (`checkout`), e tutti i comandi successivi (`rooms`, `book`, `status`) si riferiscono a quel contesto finché non viene cambiato.

---

## 5. Specifiche dei Comandi

### 5.1 `config`

Configura le impostazioni globali dell'utente. Questi dati vengono salvati localmente (es. `~/.unconf.yaml`) e usati per pre-compilare le registrazioni.

**Sinossi:**

```bash
unconf config [flags]
```

**Opzioni:**

- `--user <string>`: imposta il nome completo (es. "Mario Rossi")
- `--email <string>`: imposta l'email di contatto
- `--privacy <public|private>`: (default: `private`) imposta la visibilità di default sulla room list
- `--view`: mostra la configurazione attuale

---

### 5.2 `list` (alias: `ls`)

Mostra l'elenco delle conferenze disponibili sulla piattaforma.

**Sinossi:**

```bash
unconf list [flags]
```

**Opzioni:**

- `--all` (`-a`): mostra anche le conferenze passate o chiuse
- `--json`: output in formato JSON (utile per scripting)

**Output Tabellare:**

Mostra colonne: `ID`, `NAME`, `DATE`, `LOCATION`, `STATUS` (Open/SoldOut).

---

### 5.3 `checkout`

Imposta il contesto attivo su una specifica conferenza. Scarica i metadati necessari in cache locale.

**Sinossi:**

```bash
unconf checkout <conference_id>
```

**Comportamento:**

- Verifica che l'ID esista
- Salva l'ID come contesto corrente
- Mostra un messaggio di conferma con i dettagli brevi dell'evento

---

### 5.4 `info`

Mostra i dettagli completi della conferenza attualmente in "checkout".

**Sinossi:**

```bash
unconf info
```

**Output:**

Renderizza il Markdown della descrizione conferenza, inclusi prezzi, indirizzo hotel, policy di cancellazione e link utili.

---

### 5.5 `rooms` (alias: `map`)

Il cuore dell'esperienza utente. Visualizza la disponibilità delle camere.

**Sinossi:**

```bash
unconf rooms [flags]
```

**Modalità:**

1. **Standard (TUI)**: se lanciato senza flag o con `--interactive`, apre un'interfaccia grafica nel terminale (basata su *Bubble Tea*). Permette di navigare con le frecce, vedere la mappa delle stanze, chi le occupa, e i prezzi.
2. **Scripting**: se usata con flag di filtro, restituisce una lista testuale.

**Opzioni:**

- `--available`: mostra solo camere libere
- `--type <single|double|triple>`: filtra per tipologia
- `--mates`: mostra i nomi degli occupanti (dove la privacy lo consente)

---

### 5.6 `book`

Effettua la prenotazione di una camera e la registrazione alla conferenza.

**Sinossi:**

```bash
unconf book [room_id] [flags]
```

**Comportamento:**

- Se lanciato senza `room_id`, avvia un wizard interattivo (`--interactive` implicito)
- Se lanciato con `room_id` (es. `104`), tenta la prenotazione diretta
- Chiede conferma finale prima di chiamare l'API

**Opzioni:**

- `--private`: forza la privacy (nasconde il nome) per questa specifica prenotazione
- `--notes <string>`: aggiunge note per l'hotel (es. "Celiaco", "Letto aggiuntivo")
- `--force`: salta il prompt di conferma (solo per utenti esperti/script)

---

### 5.7 `status` (alias: `whoami`)

Mostra lo stato della tua registrazione per la conferenza corrente.

**Sinossi:**

```bash
unconf status
```

**Output:**

- Stato registrazione: (None / Pending / Confirmed)
- Camera assegnata: (es. "202 - Sea View")
- Roommate: (es. "@luigi_verdi")
- Importo dovuto: (es. "180EUR")

---

### 5.8 `cancel`

Cancella la prenotazione corrente.

**Sinossi:**

```bash
unconf cancel
```

**Comportamento:**

- Richiede una conferma esplicita ("Sei sicuro? Questa azione è irreversibile")
- Invia la richiesta di cancellazione all'API
- Libera la camera per altri utenti

---

## 6. Scenari di Utilizzo (User Journey)

### Scenario A: Il "First Timer" (Interattivo)

L'utente vuole esplorare e lasciarsi guidare.

```bash
# 1. Setup iniziale
$ unconf config --user "Giulia Bianchi" --email "giulia@tech.it"

# 2. Cerca conferenze
$ unconf ls
ID              NAME           STATUS
socrates-26     SoCraTes IT    OPEN
ruby-day-26     RubyDay        OPEN

# 3. Seleziona SoCraTes
$ unconf checkout socrates-26
Context set to 'SoCraTes IT 2026'.
Location: Hotel Rimini, 20-22 Maggio.

# 4. Esplora le stanze (TUI)
$ unconf rooms
# (Si apre l'interfaccia grafica. Giulia vede che la stanza 204 ha vista mare
# ed è occupata da un sua amica, @paola. Esce dalla TUI memorizzando l'ID).

# 5. Prenota (Wizard)
$ unconf book 204
> Room 204 is a Double Room (Sea View).
> Currently occupied by: Paola.
> Price: 150EUR.
> Confirm booking? [Y/n]: Y
> Success! Confirmation email sent to giulia@tech.it.
```

### Scenario B: Il "Power User" (Rapido)

L'utente sa già cosa vuole e usa comandi diretti.

```bash
# Assume che config e checkout siano già fatti
$ unconf status
No booking found for 'socrates-26'.

# Cerca rapidamente una singola libera
$ unconf rooms --type single --available
ID    TYPE      PRICE   FEATURES
301   Single    100     No View
305   Single    120     Balcony

# Prenota direttamente la 305 nascondendo il nome
$ unconf book 305 --private --notes "Vegan breakfast" --force
Booking confirmed. Room 305 reserved.
```

### Scenario C: L'Organizzatore (Analisi)

Usa il tool per estrarre dati.

```bash
$ unconf checkout socrates-26
$ unconf list --json > report_conferenze.json
```

---

## 7. Tech Stack

- **Language**: Go (Golang) 1.21+
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) (standard de facto per CLI Go)
- **TUI Library (MANDATORY)**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (interfaccia grafica), [Bubbles](https://github.com/charmbracelet/bubbles) (componenti come liste/input), [Lip Gloss](https://github.com/charmbracelet/lipgloss) (styling). **All interactive views must follow this stack.**
- **Config Mgmt**: [Viper](https://github.com/spf13/viper) (gestione file YAML/env vars)
- **Backend**: Go con [Gin](https://github.com/gin-gonic/gin) (REST API), containerizzato con Docker
- **Database**: SQLite (con repository pattern per futura migrazione a PostgreSQL)
- **Email**: SendGrid, Mailgun o SMTP standard; MailHog per test locali
- **Backend Communication**: Resty o net/http standard

---

## 8. Differenziatori

1. **Social Room Selection**: non solo prenotazione, ma scelta consapevole del compagno di camera
2. **Privacy by Design**: ogni partecipante controlla la propria visibilità
3. **Handoff Smart**: automazione fino all'hotel, poi processo tradizionale (nessun lock-in tecnologico)
4. **Open Space Spirit**: riflette i valori di trasparenza e collaborazione della community

---

## 9. Success Metrics (PoC)

- Tempo medio di registrazione < 3 minuti
- 80% dei partecipanti usa la funzionalità "vedi chi c'è"
- 0 errori di trascrizione dati verso l'hotel
- Riduzione del 90% del lavoro manuale per gli organizzatori

---

## 10. Value Proposition

**Per i Partecipanti:**

- "Prenota la tua camera in 2 minuti e scegli con chi dormire"
- Trasparenza totale su disponibilità e costi
- Controllo sulla propria privacy

**Per gli Organizzatori:**

- Riduzione drastica del lavoro manuale di coordinamento
- Nessun foglio Excel da aggiornare
- Automazione completa della comunicazione con l'hotel

**Per l'Hotel:**

- Ricezione dati strutturati e completi
- Riduzione errori di trascrizione
- Processo di conferma standard già conosciuto (via email)

---

## 11. Prossimi Passi (Implementation Plan)

1. Inizializzare il modulo Go: `go mod init github.com/username/unconf`
2. Setup di Cobra: creare lo scheletro dei comandi (`root`, `config`, `list`)
3. Implementare il mock dei dati (senza backend reale all'inizio) per testare la TUI `rooms`
4. Sviluppare la vista TUI con Bubble Tea

---

## 12. CI/CD & Distribution

**Strategy**: Use [GoReleaser](https://goreleaser.com/) with GitHub Actions for automated cross-platform distribution.

**Workflows:**
1.  **PR Check**: Run tests (`go test ./...`) and linting (`golangci-lint`) on every PR.
2.  **Release**:
    *   Triggered by Git Tags (e.g., `v1.0.0`).
    *   **GoReleaser Action**:
        *   Cross-compiles for Linux (amd64/arm64), macOS (amd64/arm64), and Windows.
        *   Injects version metadata via `ldflags`.
        *   Creates a GitHub Release with artifacts (tar.gz/zip).
        *   (Optional) Updates Homebrew tap.

**Note**: Setup GoReleaser configuration (`.goreleaser.yaml`) when preparing the first alpha release.
