---
description: Skill per creare un repository completo — entità e interfaccia nel dominio, migrazione, mapper, implementazione SQL, fake e cablaggio nel composition root.
goal: Dare al dominio una nuova capacità di accesso ai dati mantenendolo puro, con l'implementazione confinata in data/ e la conformità verificata dal compilatore.
---
# Skill: Creazione Repository

## Obiettivo

Rendere disponibile un dato al resto dell'applicazione senza che nessuno, fuori da `data/`,
sappia da dove arriva.

## Quando usarla

-   Un soggetto ha bisogno di leggere o scrivere dati che ancora non espone.
-   Un servizio o un handler sta per usare direttamente un `*sql.DB` o un client HTTP.

Se la fonte è un servizio di terze parti, il flusso si allunga di una
[sorgente](../external_source_integration/external_source_integration.skill.md), da scrivere
fra il passo 4 e il passo 5.

## Procedura

1.  **Verifica il contesto**
    -   Leggi [Struttura del progetto](../../rules/project_structure.rule.md) e
        [Persistenza](../../rules/persistence.rule.md).
    -   Identifica il **soggetto**: se non ne esiste uno adatto, chiediti se è davvero nuovo
        prima di crearlo.

2.  **Entità**
    -   In `internal/domain/<soggetto>/<entità>.go`: struct senza tag, insiemi chiusi come tipi
        `string` con `Valid()`, logica derivata come metodi.
    -   [Vedi Esempio](../../examples/domain_entity.example.md)

3.  **Interfaccia**
    -   In `internal/domain/<soggetto>/repository.go`: solo i metodi che un consumatore usa,
        con la convenzione sull'assenza (`Get` → `NotFound`, `FindBy…` → `ok`).
    -   [Vedi Esempio](../../examples/domain_repository.example.md)

4.  **Migrazione**
    -   Un file nuovo `NNNN_<cosa>.sql` in `internal/data/migrations/`, mai una modifica a uno
        esistente. `UNIQUE` per ogni invariante, indici per i filtri frequenti.
    -   [Vedi Esempio](../../examples/migration.example.md)

5.  **Mapper e implementazione**
    -   In `internal/data/<soggetto>/`: la riga con i tipi delle colonne, `toEntity`, e il
        repository che traduce `sql.ErrNoRows` e avvolge ogni altro errore.
    -   [Vedi Esempio](../../examples/repository_impl.example.md)

6.  **Fake**
    -   In `internal/domain/<soggetto>/<soggetto>test/`, con la riga
        `var _ <soggetto>.<Interfaccia> = (*Fake)(nil)` che si rompe quando il contratto cambia.
    -   [Vedi Esempio](../../examples/table_test.example.md)

7.  **Cablaggio**
    -   In `internal/bootstrap/run.go`: `New<Nome>Repository(db)` passato al costruttore del
        servizio. È qui che il compilatore verifica la conformità.
    -   [Vedi Esempio](../../examples/main_and_shutdown.example.md)

## Vincoli

-   Il repository non contiene logica di business: ordinamenti e filtri in SQL, decisioni nel
    servizio.
-   Nessun `*sql.Rows` o tipo di `database/sql` nella firma di un metodo pubblico.
-   Un errore imprevisto dell'archivio è un errore, non una lista vuota.

## Verifica

-   Un test in `internal/data/<soggetto>/` su SQLite vero in `t.TempDir()`, con le migrazioni
    vere: lettura, scrittura, assenza, vincolo di unicità.
-   In locale `gofmt -w .`; il resto in CI.
