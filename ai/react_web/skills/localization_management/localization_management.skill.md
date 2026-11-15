---
description: Skill per la gestione delle etichette localizzate di un sito React, con cataloghi tipizzati, prefisso della feature e accesso dai componenti tramite hook.
goal: Garantire che ogni etichetta mostrata all'utente viva nei cataloghi di shared/i18n/, con la stessa chiave in tutte le lingue e nessuna stringa scritta dentro un componente.
---

# Skill: Gestione delle Etichette Localizzate

Questa skill descrive come aggiungere, rinominare ed eliminare un'etichetta in
`web/src/shared/i18n/`.

## 📋 Obiettivo

Tenere tutte le stringhe mostrate all'utente in cataloghi tipizzati, così che una chiave
mancante in una lingua o rimasta dopo una rinomina **non compili**, invece di comparire in
pagina.

## 📐 Regole per una chiave

1.  **Prefisso obbligatorio**: `<nome_feature>_` (es. `orders_title`), oppure `common_` se
    l'etichetta non appartiene a nessuna feature.
2.  **`snake_case`** dopo il prefisso, e un nome che dica **cosa è**, non dove sta: `orders_empty`
    e non `orders_center_text`.
3.  **Parametri nominati** fra graffe: `{amount}`, `{count}`. Mai posizionali.
4.  **Nessuna concatenazione**: una frase è una chiave sola. Due mezze frasi unite nel codice si
    traducono solo in una lingua.
5.  **Niente etichette dentro `shared/components/`**: un componente condiviso riceve il testo già
    tradotto da chi lo usa.

## 🛠️ Aggiungere un'etichetta

1.  Apri il catalogo della **lingua base** (`src/shared/i18n/<lingua_base>.ts`) e aggiungi la
    chiave nel blocco del suo prefisso.
2.  Aggiungila in **tutti gli altri cataloghi**: sono annotati
    `Record<TranslationKey, string>`, quindi finché manca non compilano.
3.  Usala nel componente con l'hook, mai con una stringa scritta a mano:

    ```tsx
    const t = useTranslate();
    // …
    <p>{t('orders_empty')}</p>
    ```

4.  Il tipo delle chiavi nasce dalla lingua base: una chiave sbagliata è un errore di `tsc`,
    e a dirlo è la CI — vedi [Verifica delle Modifiche](../../rules/verification.rule.md).

-   [Formato dei cataloghi](./examples/i18n_catalog.example.md)
-   [Accesso dai componenti](../../examples/i18n_hook.example.md)

## ♻️ Rinominare o eliminare

-   **Rinominare**: cambia la chiave nel catalogo base e poi negli altri. `tsc` elenca da solo
    tutti i punti da aggiornare: il tipo `TranslationKey` nasce dal catalogo base, quindi la
    vecchia chiave smette di esistere.
-   **Eliminare**: togli la chiave da tutti i cataloghi. Se `tsc` passa, non era usata da
    nessuna parte — è l'unico controllo che serve.
-   Quando si elimina una feature, si eliminano le chiavi con il suo prefisso: è il motivo per
    cui il prefisso esiste.

## ⚠️ Vincoli

-   Il catalogo **non contiene dominio**: nessuna entità, nessuna regola. Quale etichetta
    corrisponda a un fallimento lo decide `features/<f>/state/`, ed è l'ultimo punto in cui il
    vocabolario delle `Failure` esiste
    ([Result e fallimenti](../../examples/result_pattern.example.md)).
-   `src/domain/` e `src/data/` **non conoscono le etichette**: un messaggio di errore tradotto
    che nasce nel data layer è una regola di presentazione finita nel posto sbagliato.
-   Le lingue supportate e quale sia la base le dichiara il `CLAUDE.md` del progetto.

## ✅ Verifica

Il controllo dei tipi gira nella CI — vedi
[Verifica delle Modifiche](../../rules/verification.rule.md). La `grep` qui sotto non è un
check: legge, e serve mentre si lavora. Dalla cartella `web/`:

```bash
# nessuna stringa visibile scritta direttamente nel JSX
grep -rn ">[A-Za-zÀ-ÿ][^<>{}]\{3,\}<" src/features src/shared/components --include=*.tsx
```

La `grep` è un aiuto, non una garanzia: legge il JSX più semplice e non vede una stringa passata
come proprietà. Quello che garantisce davvero è il tipo — una chiave inventata non compila.
