---
description: Skill per dare a un soggetto una nuova fonte di dati in un sito React — porta di dominio, modello, controllo di forma, sorgente, mapper, repository e cablaggio nel composition root.
goal: Rendere un dato disponibile alle feature senza che nessuno, sopra il repository, sappia da dove arriva, con fetch e storage del browser confinati nelle sorgenti e nessun passo del flusso dimenticato.
---

# Skill: Creazione di Repository e Sorgenti

Questa skill copre il flusso completo con cui un soggetto acquisisce una nuova fonte di dati:
la porta in `web/src/domain/<soggetto>/` e tutto ciò che serve a soddisfarla in
`web/src/data/<soggetto>/`.

## 📋 Obiettivo

Portare un dato fino alle feature attraverso un **contratto di dominio**, tenendo il
trasporto — endpoint, formato, `localStorage`, SDK di terzi — confinato nelle sorgenti, dove
può essere sostituito senza che nessun altro file cambi.

## 🕐 Quando usarla

-   Un soggetto deve leggere o scrivere dati che ancora non espone.
-   Uno store, una pagina o uno use case sta per chiamare `fetch` o leggere `localStorage`
    da sé — il linter lo rifiuterà comunque, e questa è la strada giusta.
-   Un soggetto che ha già un repository guadagna un secondo confine tecnico (una copia
    locale, l'SDK di un servizio): bastano i passi da 4 a 7.

## 🛠️ Procedura

1.  **Verifica il contesto**
    -   Leggi [domain/ – Domain Layer](../../rules/domain_layer.rule.md) e
        [data/ – Data Layer](../../rules/data_layer.rule.md).
    -   Identifica il **soggetto**: un concetto di business, non una pagina. Se nessuno di
        quelli esistenti è adatto, controlla i criteri prima di crearne uno nuovo.
    -   Guarda i soggetti già presenti e riusa il loro naming.

2.  **Definisci o estendi le entità**
    -   In `domain/<soggetto>/entities/`, un `type` con campi `readonly`, la logica derivata
        come funzione pura accanto, gli insiemi chiusi come unioni di letterali.
    -   [Vedi Esempio](../../examples/domain_entity.example.md)

3.  **Scrivi la porta di dominio**
    -   In `domain/<soggetto>/repositories/<soggetto>-repository.ts`, un'`interface` i cui
        metodi parlano di **entità** e restituiscono `Promise<Result<T>>`.
    -   La porta dichiara *cosa serve al dominio*, nella lingua del dominio: non nomina
        pagine, endpoint, cache o status code.
    -   [Vedi Esempio](../../examples/domain_repository.example.md)

4.  **Definisci il modello e il suo controllo di forma**
    -   In `data/<soggetto>/models/`, un `type` per ogni forma che la fonte restituisce, con i
        campi e i nomi che arrivano davvero (`placed_at`, non `placedAt`).
    -   Accanto, la funzione che verifica che un `unknown` abbia quella forma. È codice, non
        una `as`: quello che arriva da `fetch` o da `JSON.parse` è `any`, e va fermato qui.
    -   Un modello non esce mai da `data/`.

5.  **Scrivi la sorgente: contratto e implementazione**
    -   Una sorgente per ogni **confine tecnico**: la fonte remota e la copia locale sono due
        sorgenti anche quando servono lo stesso dato.
    -   Il contratto prende il nome del confine (`<soggetto>-source.ts`, `<soggetto>-cache.ts`),
        l'implementazione ci mette davanti il trasporto (`http-<soggetto>-source.ts`,
        `local-<soggetto>-cache.ts`) ed è una factory che annota il tipo di ritorno.
    -   La sorgente restituisce `Promise<Result<Modello>>` e **non solleva**: è qui, e solo qui,
        che vivono il `try/catch` e la traduzione di uno stato HTTP in una `Failure`.
    -   L'URL di base e ogni altra configurazione si **ricevono** come argomento: le legge
        `bootstrap/config.ts`, non la sorgente.
    -   [Vedi Esempio](../../examples/source_impl.example.md)

6.  **Scrivi il mapper**
    -   In `data/<soggetto>/mappers/`, una funzione pura modello → entità. È lui a
        **interpretare**: default, valori sconosciuti di un insieme chiuso, date.
    -   [Vedi Esempio](../../examples/repository_impl.example.md)

7.  **Scrivi il repository**
    -   In `data/<soggetto>/repositories/`, una factory che riceve le sorgenti e restituisce la
        porta. Il nome dice la strategia di composizione (`cached-<soggetto>-repository.ts`).
    -   Decide quale sorgente interrogare e quando ripiegare, chiama il mapper, lascia passare
        le `Failure` così come sono. Niente logica di business: quella è uno use case.
    -   [Vedi Esempio](../../examples/repository_impl.example.md)

8.  **Cabla nel composition root**
    -   In `bootstrap/dependencies.ts`, costruisci il repository con le sue sorgenti e
        **annotalo con la porta**: è lì che TypeScript verifica la conformità.
    -   Se la sorgente chiede configurazione nuova, aggiungi la variabile a `config.ts`, a
        `vite-env.d.ts` e a `.env.example` — e chiediti prima se è davvero pubblica
        ([Configurazione](../../rules/configuration.rule.md)).
    -   Una libreria nuova che parla con l'esterno va aggiunta a `onlyInSources` della
        [tabella dei confini](../../examples/architecture_test.example.md).
    -   [Vedi Esempio](../../examples/bootstrap_dependencies.example.md)

## ⚠️ Vincoli

-   **Nessuno sopra il repository sa che le sorgenti esistono**: né le feature, né gli use
    case, né gli store.
-   **Le sorgenti restituiscono modelli, il repository restituisce entità.** Il mapping è uno,
    per tutte le fonti.
-   **Niente segreti nelle sorgenti**: un token che autorizza qualcosa non può stare in un
    sito, perché finisce nel pacchetto che ogni visitatore scarica.

## ✅ Verifica

-   Scrivi il test del repository: la composizione si prova con le sorgenti finte, come
    oggetti letterali annotati con il loro contratto — fonte a posto, fonte caduta con copia
    locale piena, fonte caduta con copia vuota.
-   I check dello stack girano nella CI: vedi
    [Verifica delle Modifiche](../../rules/verification.rule.md). La `grep` qui sotto legge, e
    serve mentre si lavora — dalla cartella `web/`:

    ```bash
    grep -rn "fetch(\|localStorage\|indexedDB" src --include=*.ts --include=*.tsx \
      | grep -v "src/data/.*/sources/"
    ```

    La `grep` deve essere vuota.
