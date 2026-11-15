---
description: Skill per creare use case di dominio in TypeScript, come funzioni pure o fabbriche che ricevono le porte, tenendo la logica di business fuori da store e componenti.
goal: Guidare la creazione di use case piccoli, tipizzati e testabili senza render, che restituiscono entità di dominio.
---

# Skill: Creazione Use Case

Questa skill descrive come creare uno use case in
`web/src/domain/<soggetto>/use-cases/`.

## 📋 Obiettivo

Estrarre una regola di business — un calcolo, un ordinamento, una selezione, un coordinamento
fra entità e porte — in una funzione dedicata, prima che finisca dentro uno store o un
componente.

## 🕐 Quando usarla

-   La logica sta per essere scritta (o vive già) dentro uno store o un componente.
-   Una regola di dominio è duplicata, o si può provare solo rendendo qualcosa.
-   Serve coordinare più entità o più porte per un singolo caso d'uso.

Non serve per il mapping di presentazione (un'etichetta formattata, una frazione per una barra):
quello è un view model e resta in `features/<f>/state/`. Non serve nemmeno per un passaggio
diretto a un repository senza logica aggiuntiva.

## 🛠️ Procedura

1.  **Verifica il contesto**
    -   Leggi [domain/ – Domain Layer](../../rules/domain_layer.rule.md).
    -   Guarda gli use case già presenti nel soggetto e riusa naming e forma.
    -   Se il progetto non ha ancora il contratto base, parti dal
        [Contratto UseCase](../../examples/use_case_contract.example.md).

2.  **Scegli la forma**
    -   Senza dipendenze esterne: una **funzione esportata**, annotata
        `UseCase<In, Out>`.
    -   Con delle porte: una **fabbrica `make<Verbo><Soggetto>(porte)`** che le riceve come
        argomenti e restituisce la funzione finale, annotata `AsyncUseCase<In, Out>`.
    -   Mai una `class` con un solo metodo, e mai un'`interface` per lo use case: sarebbe la
        copia dell'unica implementazione possibile. Se ti serve davvero, quello che hai in mano
        è una porta.
    -   [Vedi Esempio](../../examples/use_case.example.md)

3.  **Crea il file**
    -   `web/src/domain/<soggetto>/use-cases/<verbo>-<soggetto>.ts`, simbolo esportato
        `<verbo><Soggetto>`.
    -   Più ingressi si passano come **un oggetto**, così `In` resta un tipo solo.
    -   Uno `switch` su un insieme chiuso finisce con `assertNever`
        ([Result e fallimenti](../../examples/result_pattern.example.md)).

4.  **Rispetta i confini**
    -   Una responsabilità: un caso d'uso, un output.
    -   Ritorna **entità di dominio**, mai forme pensate per la UI.
    -   Se l'output è un concetto di dominio nuovo, l'entità nasce in
        `domain/<soggetto>/entities/` — [Vedi Esempio](../../examples/domain_entity.example.md).
    -   Import ammessi: solo `@core/*` e il proprio soggetto. Mai `react`, `react-dom`,
        `react-router`, `@data/*`, `@features/*`, `@shared/*`.
    -   Nessuna copia mutata di ciò che riceve: `sort`, `reverse`, `splice` mutano, e quello che
        arriva è un'entità.

5.  **Collega lo use case**
    -   Se è puro, lo store lo importa e lo chiama.
    -   Se riceve porte, lo costruisce il
        [composition root](../../examples/bootstrap_dependencies.example.md), oppure la fabbrica
        dello store a cui le porte sono già arrivate. In nessun caso lo use case importa
        `@bootstrap/dependencies`.

## ✅ Verifica

-   Scrivi il test: caso nominale e casi limite (lista vuota, valori a zero, input che non va
    modificato) — [Vedi Esempio](../../examples/use_case_test.example.md).
-   I check dello stack girano nella CI, non qui: vedi
    [Verifica delle Modifiche](../../rules/verification.rule.md). In locale resta il solo
    `npx prettier --write .`. La `grep` invece legge, e serve mentre si lavora — dalla
    cartella `web/`:

    ```bash
    grep -rn "from 'react\|@data/\|@features/\|@shared/" src/domain
    ```

    La `grep` deve essere vuota.
