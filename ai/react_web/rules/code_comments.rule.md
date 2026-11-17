# Commenti nel Codice

Un commento è codice che nessun compilatore controlla: nessuno lo aggiorna quando la riga
sotto cambia, e il giorno in cui mente costa più di quanto sia mai valso. Se ne scrivono
pochi, e solo quelli che nessun nome e nessun tipo potrebbero dire al posto loro.

È lo stesso principio degli altri stack
(Flutter,
Python); cambia ciò che in TypeScript dice già da
sé quello che altrove andrebbe scritto a parole. Il linguaggio è lo stesso di
React Native, e così la regola: quando una
delle due cambia, l'altra va riletta.

## ✍️ Il commento spiega il *perché*, mai il *cosa*

Se il commento descrive quello che il codice già dice a chi lo legge, va cancellato.

```ts
// ❌ dice quello che si vede
// Filtra gli elementi attivi e li ordina per data.
const visible = items.filter((i) => i.isActive).sort(byDate);

// ✅ dice quello che non si vede
// Ordinati per data e non per rilevanza: la lista si aggiorna sotto gli occhi di chi la
// sta leggendo, e un ordine che cambia sposterebbe le righe già lette.
const visible = items.filter((i) => i.isActive).sort(byDate);
```

Si commenta:

-   una **decisione non ovvia**, e la strada più semplice che è stata scartata;
-   un **caso limite** e perché è trattato così;
-   un **vincolo esterno** (una fonte lenta, un formato che non si controlla, un
    comportamento che cambia da un browser all'altro);
-   un **numero scelto** — una cadenza, una soglia, una dimensione di pagina — che altrimenti
    sembrerebbe arbitrario;
-   una **rinuncia al controllo dei tipi**, che senza il perché è indistinguibile da una resa.

## 🏷️ I tipi non si riscrivono a parole

`tsc -b` gira su tutto il progetto con `strict: true`, quindi la firma è già
documentazione **verificata**. Un blocco JSDoc che ne ripete i tipi è la stessa informazione
una seconda volta, non controllata da nessuno, che invecchia al primo cambio di firma.

```ts
// ❌ cinque righe per dire quello che la firma dice meglio
/**
 * Trova un ordine.
 * @param orderId l'identificativo dell'ordine
 * @returns l'ordine, oppure undefined
 */
function findOrder(orderId: string): Order | undefined;

// ✅ il commento dice ciò che la firma non può dire
/** `undefined` se l'ordine non esiste: l'assenza non è un fallimento, e chi chiama decide. */
function findOrder(orderId: string): Order | undefined;
```

Le annotazioni `@param` e `@returns` non si scrivono mai: sono i tipi, copiati a mano in un
posto dove nessuno li verifica.

## 🚫 Nessun commento che ripete il nome del simbolo

```ts
// ❌
/** Un ordine. */
export type Order = {
  /** Il titolo dell'ordine. */
  readonly title: string;
};

// ✅ — niente commento: i nomi e i tipi bastano
export type Order = {
  readonly title: string;
};
```

Un commento su un tipo o su una funzione **esportata** serve quando il nome non basta a usarla
correttamente: un'unità di misura, il significato di un `undefined`, un invariante che chi
chiama deve rispettare, il motivo per cui un contratto esiste. Su ciò che non esce dal modulo
è rumore.

## 🤫 Ogni silenziamento porta con sé la sua ragione

`@ts-expect-error`, `@ts-ignore` e `eslint-disable-next-line` spengono l'unico controllo
automatico che questo stack ha. Non si scrivono mai da soli:

```ts
// ✅ si sa cosa si sta accettando, e quando finirà
// I tipi dell'SDK dichiarano il parametro obbligatorio, ma a runtime è opzionale e ometterlo
// è l'unico modo di ottenere il comportamento predefinito.
// @ts-expect-error
client.connect();
```

Fra i due, si usa **`@ts-expect-error`** e non `@ts-ignore`: il primo fallisce in compilazione
quando l'errore che stava coprendo sparisce, quindi si rimuove da sé il giorno in cui la
libreria si sistema. `@ts-ignore` resta lì per sempre a nascondere errori che nel frattempo
sono diventati altri.

Un `as` non è un silenziamento formale ma si comporta come tale: quando serve davvero — e
sul confine di una sorgente non serve, lì si controlla la forma
([Cosa attraversa il confine](./data_layer.rule.md#cosa-attraversa-il-confine)) — porta con
sé la ragione per cui chi scrive sa più del compilatore.

## 🧪 Nei test, il nome è il commento

Un test si chiama con una frase che descrive il comportamento atteso
(`'restituisce una lista vuota quando nessun ordine è attivo'`). Un commento che spiega cosa
sta provando un test è quasi sempre il segno che il nome è sbagliato: si sistema il nome, non
si aggiunge la spiegazione.

Resta utile il commento che dice perché quel caso limite **esiste** — da quale comportamento
reale è nato, quale difetto ha prodotto — perché quello il nome non può dirlo.

## 🎫 Niente task, issue o PR nel codice

Mai, in nessuna forma: né `(#412)`, né `// TODO(#88)`, né «vedi la PR», né il nome di un
branch.

Quei codici vivono nei **messaggi di commit**, dove restano attaccati alla modifica che li ha
motivati. Dentro un file le sopravvivono: l'issue si chiude, il codice intorno cambia, e resta
un numero che rimanda a una discussione che non descrive più niente — e nessuno se ne accorge,
perché niente lo verifica.

Se il motivo di una riga sta in un'issue, allora **il motivo va scritto nel commento**, a
parole, e l'issue non serve più.

Stessa sorte per `TODO` e `FIXME`: o la cosa si fa, o diventa un task tracciato fuori dal
codice.

È un divieto su un pattern dentro un file, quindi è esprimibile come test e va reso tale — il
criterio è in [Test di Architettura](../../quality/rules/architecture_tests.rule.md). Mentre si
lavora, la verifica al volo da `web/`:

```bash
grep -rn "#[0-9]\|TODO\|FIXME" src --include=*.ts --include=*.tsx
```

## 📄 Rimandi al documento dei requisiti

Sono ammessi, ma solo dove portano un peso reale: quando un numero o un comportamento non
ovvio verrebbe altrimenti scambiato per arbitrario, il rimando dice a chi legge che quella
scelta ha una fonte e dove trovarla.

```ts
// ✅ il numero ha una fonte, e il rimando la indica
// Ogni quanto si rilegge lo stato di ciò che si sta seguendo (PRD §10).
const POLL_INTERVAL_MS = 1500;

// ❌ decorazione: il rimando non aggiunge niente a un nome già chiaro
readonly title: string; // il titolo dell'ordine (PRD §4.2)
```

Mai come decorazione ripetuta su ogni campo. La forma del rimando e il documento a cui punta
li dichiara il `CLAUDE.md` del progetto: il framework non impone quale sia.

## 🌍 Lingua

Una sola lingua per tutto il codice di un progetto, quella dichiarata nel suo `CLAUDE.md`.
Metà commenti in una e metà nell'altra è il segno che il file è passato di mano senza che
nessuno rileggesse.

## 🔗 Correlate

-   [Convenzioni UI](./ui_conventions.rule.md): il limite di dimensione dei componenti conta le
    righe di **codice**, non i commenti — sfoltirli non è il modo di rientrarci.
-   [Verifica delle Modifiche](./verification.rule.md): `tsc -b` è il controllo che rende
    superflua la documentazione dei tipi, ed è anche quello che i silenziamenti spengono.
