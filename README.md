# UNCONF CLI - Spec-Driven Development Experiment

> "Choose your room, choose your roommate, focus on the conference."

Questo repository ospita lo sviluppo di **UNCONF**, una CLI Application reale, utilizzata come base di partenza per sperimentare e validare la metodologia **Spec-Driven Development (SDD)**.

L'obiettivo non è solo scrivere codice, ma derivarlo rigorosamente da documenti di specifica, piani e task definiti *prima* dell'implementazione.

## 🧪 Il Metodo: Spec-Driven Development (SDD)

In questo progetto, nessuna riga di codice viene scritta senza una specifica approvata. Il flusso di lavoro segue rigorosamente questi step:

1.  **Specifica**: Definizione dei requisiti e user stories nella cartella `specs/`.
2.  **Pianificazione**: Creazione di un piano tecnico e suddivisione in task atomici.
3.  **TDD (Test-Driven Development)**: Scrittura dei test basati sui criteri di accettazione delle spec.
4.  **Implementazione**: Scrittura del codice per soddisfare i test.

Troverai la "memoria" e le regole del progetto in `.specify/` e le specifiche delle feature in `specs/`.

---

## 🚀 L'Applicazione: UNCONF

**UNCONF** è uno strumento da riga di comando (CLI) con interfaccia grafica testuale (TUI) pensato per le conferenze developer "Open Space" (come *SoCraTes Italia* o *Polenta e Deploy*).

Risolve il problema della gestione manuale delle prenotazioni alberghiere, trasformando la burocrazia (email, fogli Excel) in un'esperienza "nerd" e social.

### Funzionalità Chiave
*   **Velocità**: Registrazione e prenotazione hotel in < 3 minuti.
*   **TUI Interattiva**: Visualizzazione grafica delle stanze nel terminale (stile "Crush").
*   **Social Discovery**: Vedi chi ha prenotato in quale stanza e scegli i tuoi roommate.
*   **Privacy-first**: Controllo granulare sulla visibilità del proprio nome.

---

## 🛠 Tech Stack

Il progetto è scritto interamente in **Go (Golang)** v1.21+.

*   **CLI Framework**: [Cobra](https://github.com/spf13/cobra) per la gestione dei comandi.
*   **TUI & Styling**: Stack [Charm Bracelet](https://charm.sh/) (Bubble Tea, Lip Gloss, Bubbles) per interfacce testuali ricche e interattive.
*   **Configurazione**: [Viper](https://github.com/spf13/viper).
*   **Backend**: Gin Gonic + SQLite (containerizzato con Docker).

---

## 🏗 Architettura

L'architettura segue i principi della **Hexagonal Architecture (Ports and Adapters)** e impone regole ferree definite nella "Costituzione" del progetto:

1.  **Repository Pattern (Non-Negoziabile)**: L'accesso ai dati è astratto tramite interfacce. Nessuna query SQL risiede nella business logic. Questo prepara il sistema a future migrazioni (es. da SQLite a PostgreSQL).
2.  **CLI-First Design**: L'interfaccia segue le convenzioni Unix. Utilizza un sistema di **contesto** (simile a `kubectl` o `git`) tramite il comando `checkout` per mantenere lo stato della conferenza attiva.
3.  **Separazione CLI/Backend**:
    *   La **CLI** agisce come un client "stupido" ma bello, che comunica via REST API.
    *   Il **Backend** gestisce la logica di business, la persistenza e l'invio delle email.
4.  **Testing Strategy**:
    *   **Unit Test**: Per la logica interna.
    *   **Contract Test**: Per garantire che CLI e Backend si parlino correttamente (Pact).
    *   **Integration Test**: Per i comandi CLI end-to-end.

### Struttura Comandi

L'albero dei comandi riflette il flusso utente:

```bash
unconf
├── config       # Setup utente locale
├── list         # Elenco conferenze
├── checkout     # Selezione contesto conferenza
├── rooms        # TUI interattiva selezione stanze
├── book         # Wizard di prenotazione
├── status       # Verifica stato prenotazione
└── cancel       # Cancellazione
```
