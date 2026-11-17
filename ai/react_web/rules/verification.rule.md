# Verifica delle Modifiche

## 🎯 Il cancello non è su questa macchina

I check di questo stack girano **nella CI**, sul [job di verifica](../examples/ci_job.example.yml).
In locale se ne esegue uno solo, il formattatore, e non è un cancello: il perché — e perché
l'agente apre la pull request in draft invece di dichiarare che la verifica passa — sta in
Dove gira la verifica.

## 📋 I check dello stack

Dalla cartella `web/`:

| Famiglia | Comando | Dove gira |
|---|---|---|
| Formattazione | `npx prettier --check .` | CI — **e in locale** nella forma che riscrive, `npx prettier --write .` |
| Lint e confini fra livelli | `npx eslint .` | solo CI |
| Fogli di stile | `npx stylelint "src/**/*.css"` | solo CI |
| Analisi statica | `npx tsc -b` | solo CI |
| Test (architettura compresa) | `npm test -- --run` | solo CI |
| Build di verifica | `npx vite build` | solo CI |

I comandi assumono `npm`. Un progetto che usa un altro gestore di pacchetti lo dichiara nel
proprio `CLAUDE.md`: cambia il prefisso, non cambia l'elenco.

`tsc -b` è il check che in questo stack pesa di più, ed è il motivo per cui
`tsconfig.app.json` ha `strict: true` insieme a `noUncheckedIndexedAccess` e
`exactOptionalPropertyTypes`. Senza `strict`, metà delle regole di questo framework — il
restringimento di [`Result`](./domain_layer.rule.md#%EF%B8%8F-result-e-fallimenti), l'esaustività
degli `switch`, il confine in cui `unknown` diventa un modello — smettono di essere verificate
e tornano a essere abitudini.

È `-b` e non `--noEmit`, e la differenza non è di gusto. Il `tsconfig.json` di un progetto
Vite non contiene file: contiene **riferimenti** a `tsconfig.app.json` e `tsconfig.node.json`.
`tsc --noEmit` lanciato sulla radice controlla l'insieme vuoto ed esce verde; `tsc -b` segue i
riferimenti e controlla entrambi i progetti. È il caso esemplare di un check che passa senza
aver verificato niente. Vedi [Vite e TypeScript](../examples/vite_and_tsconfig.example.md).

`npx vite build` **non** controlla i tipi — toglie le annotazioni e impacchetta — ed è per
questo che i due step sono separati: la build dice che i moduli si risolvono e che il sito si
costruisce, `tsc` dice che il codice è corretto. Uno non sostituisce l'altro.

`npx eslint .` e `npx stylelint` non verificano lo stile: la tabella delle dipendenze fra
livelli è una regola del primo, il divieto dei valori scritti a mano nei fogli una regola del
secondo, e una violazione è un errore come un altro — vedi
[Confini fra livelli](../examples/architecture_test.example.md) e
[Test di Architettura](../../quality/rules/architecture_tests.rule.md). `npm test` copre la
parte che i linter non possono darsi da soli, cioè che i file da controllare esistano ancora.

## ✅ L'unico comando locale

```bash
npx prettier --write .
```

Sta in locale perché costa secondi, non controlla nessun tipo e non avvia nessuna suite, e la
sua violazione è certa: un giro di CI per una riga da riformattare è tutto sprecato.
Nient'altro — niente `eslint`, niente `tsc`, niente suite.

## 📦 Quando cambiano le dipendenze

```bash
npm install <pacchetto>          # o npm install -D per gli strumenti di sviluppo
```

Il file di lock va committato insieme al `package.json`. In CI l'installazione è `npm ci`, che
fallisce se il lock non è allineato a `package.json`: un lock dimenticato non è un dettaglio di
ambiente, è un check rosso.

Una dipendenza nuova finisce nel pacchetto che ogni visitatore scarica. Prima di aggiungerla si
guarda quanto pesa e se la piattaforma fa già la stessa cosa: una libreria di date per
formattare un giorno, quando `Intl.DateTimeFormat` lo fa nella lingua giusta, è peso pagato da
chiunque apra il sito.

## ⚙️ Comandi di lavoro

Servono a scrivere il codice, non a dichiararlo verificato. Nessuno di questi è un cancello, e
nessuno va eseguito «per sicurezza» prima di consegnare.

```bash
npx vite                        # server di sviluppo
npx vite preview                # serve la build di produzione, per provarla com'è
npm test -- tests/domain        # il singolo test che si sta scrivendo
npx tsc -b --watch              # il controllo dei tipi mentre si lavora
npx eslint . --fix              # le correzioni automatiche, quando le si vuole
```

Il server di sviluppo e la build non sono lo stesso programma: il primo serve i moduli uno per
uno, la seconda li impacchetta. Un problema che appare solo dopo la build — un percorso di
asset, un import dinamico, una variabile d'ambiente mancante — si riproduce con
`vite build && vite preview`, non con il server di sviluppo.

## 🚧 Verifica dei confini fra layer

La garanzia è `npx eslint .` in CI. Mentre si lavora restano comodi i `grep`, che rispondono a
«dove sta questa cosa» e non a «questa modifica regge». Da `web/`:

```bash
grep -rn "@data/" src/features src/domain src/app       # deve essere vuoto
grep -rn "from 'react" src/domain src/data              # deve essere vuoto
grep -rn "fetch(\|localStorage\|indexedDB" src --include=*.ts --include=*.tsx \
  | grep -v "src/data/.*/sources/"                      # deve essere vuoto
grep -rn "import.meta.env" src | grep -v "src/bootstrap/config.ts"   # deve essere vuoto
grep -rn "#[0-9]\|TODO\|FIXME" src --include=*.ts --include=*.tsx
```

L'ultimo è il divieto di [Commenti nel Codice](./code_comments.rule.md); gli altri sono le
righe della [tabella delle dipendenze](./project_structure.rule.md#%EF%B8%8F-regole-di-dipendenza-tra-livelli)
che si sbagliano più spesso.

## 🖥️ Verifica nel browser

È l'unica verifica che nessun runner fa, e con i check in CI diventa la sola cosa che resta
davvero a una persona. Quando la modifica tocca la UI o la navigazione, va guardata almeno una
volta in un browser vero, e queste quattro cose non hanno un equivalente in nessuna suite:

-   **due larghezze**: quella di un telefono e quella di un desktop. Un layout che regge solo
    alla larghezza di chi l'ha scritto è il difetto più comune di un sito;
-   **solo tastiera**: si attraversa la pagina con `Tab`, si attiva con `Invio` e `Spazio`, si
    chiude con `Esc`, e il focus si vede sempre;
-   **ricarica su una rotta profonda**: un URL interno aperto da zero deve arrivare alla stessa
    pagina. Se fallisce in produzione e non in sviluppo, è il server che non rimanda a
    `index.html` — vedi [Distribuzione](./deployment.rule.md);
-   **i due schemi di colore**, quando la modifica tocca colori o superfici.

## 🔗 Correlate

-   Dove gira la verifica: perché il cancello è uno
    solo, e cosa resta in locale.
-   Quando la CI è rossa: cosa si fa di un check
    fallito, e le mosse che lo fanno tornare verde senza verificare niente.
-   [Job di CI](../examples/ci_job.example.yml): gli stessi comandi, su macchina pulita.
-   [Confini fra livelli](../examples/architecture_test.example.md): la tabella delle dipendenze
    come regola di ESLint, e i fogli di stile sotto stylelint.
