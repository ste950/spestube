# Accessibilità

Un'interfaccia coerente e un'interfaccia usabile non sono la stessa cosa. Le regole sulla UI
degli stack coprono tema, token, dimensione dei componenti e riuso: tutto ciò che la rende
coerente, niente di ciò che la rende usabile da chi non la vede bene, non la tocca con
precisione o la naviga senza guardarla.

È un requisito con un costo asimmetrico. Un contrasto sbagliato o un bersaglio tattile troppo
piccolo si correggono in un token, ma solo finché quel token esiste: quando i valori sono
sparsi nei widget, la stessa correzione è un rifacimento. Per questo l'accessibilità è una
regola del framework e non una revisione finale — ed è anche il motivo per cui vive in
`quality/`: è una **garanzia**, come i test di architettura, e non appartiene a nessuno stack.

Quello che segue non dipende dal linguaggio. I rapporti di contrasto, le dimensioni minime, le
etichette, l'ordine di lettura sono gli stessi ovunque si disegni un'interfaccia: cambia solo
l'API con cui si esprimono, e quella sta nello stack.

## 🎨 Contrasto

I minimi, misurati come rapporto fra la luminanza del primo piano e quella dello sfondo:

| Cosa | Rapporto minimo |
|---|---|
| testo normale | **4.5:1** |
| testo grande — da 24 px, o da 18.7 px se in grassetto | **3:1** |
| elementi non testuali che veicolano informazione: icone significative, bordi di un campo, tracciato di un grafico | **3:1** |
| indicatore di focus, contro lo sfondo che copre | **3:1** |

Fuori dal vincolo restano solo il testo di un controllo **disabilitato** e la grafica
puramente decorativa: nel primo caso il contrasto basso è l'informazione, nel secondo non c'è
informazione da perdere.

Due precisazioni che decidono se la regola vale davvero qualcosa:

-   **si verifica sui token del tema, non a occhio.** Il rapporto si calcola fra due valori
    dichiarati — un colore di primo piano e il colore di superficie su cui sta — e il posto
    dove quei valori vivono è il tema. Un controllo fatto su uno screenshot verifica una
    schermata; un controllo fatto sulle coppie di token verifica tutte quelle che quei token
    comporranno.
-   **le coppie vanno dichiarate.** Il tema non sa quale primo piano finirà su quale sfondo:
    l'elenco delle combinazioni ammesse è parte del tema quanto i colori. Una coppia che non è
    nell'elenco non è verificata, e prima o poi qualcuno la userà.

Vale per **ogni modalità** che il tema dichiara: un progetto con tema chiaro e scuro ha due
serie di coppie, e passare l'una non dice niente sull'altra.

Questo criterio è **deterministico** — due numeri, una formula, un confronto — ed è quindi
esattamente il genere di cosa che va difesa da un test e non da un promemoria: vedi
[Test di Architettura](./architecture_tests.rule.md).

## 👆 Bersagli tattili

-   **48 × 48** unità logiche è il minimo per qualsiasi elemento interattivo. Le piattaforme
    dichiarano soglie leggermente diverse (44 su alcune, 48 su altre): si prende la più alta e
    si smette di distinguere.
-   **8** unità logiche di spazio fra due bersagli adiacenti. Due bersagli grandi abbastanza ma
    appiccicati producono lo stesso errore che si voleva evitare, e su un elenco di azioni
    ravvicinate è l'errore più costoso: colpirne una al posto dell'altra.
-   **L'area toccabile può essere più grande del disegno.** È la risposta al caso che si
    presenta sempre: il design chiede un'icona da 24, e va bene — si disegna l'icona da 24 e le
    si dà intorno un bersaglio da 48. Quello che non si fa è rimpicciolire il bersaglio per
    farlo combaciare con il disegno.
-   Il minimo si applica al componente **condiviso**, una volta sola. Se il bersaglio nasce
    giusto nel componente, nessuna schermata deve ricordarsene.

Il valore sta in un token di dimensione come tutti gli altri: scritto a mano dentro un widget
non è verificabile, e vale la regola dello stack sui valori di stile letterali.

## 🏷️ Etichette

**Ogni controllo senza testo visibile ha un'etichetta testuale.** Un bottone di sola icona,
un campo con il solo segnaposto, un'immagine che è un link: senza etichetta, per chi usa uno
screen reader quel controllo non esiste o si annuncia come un elemento anonimo.

Come si scrive un'etichetta utile:

-   dice **l'azione e il suo oggetto**, non il nome dell'icona: «Elimina il documento», mai
    «cestino». Chi ascolta non vede il disegno e non gliene importa;
-   **non ripete il ruolo**: la piattaforma annuncia già «bottone», «campo di testo»,
    «interruttore». Un'etichetta che lo ripete produce «bottone Elimina bottone»;
-   **segue lo stato** quando il controllo ne ha uno: lo stesso bottone è «Riproduci» o
    «Metti in pausa», non «Riproduci» sempre. Se l'etichetta non cambia con l'icona, è
    sbagliata per metà del tempo;
-   è una **stringa localizzata** come tutte le altre. Un'etichetta scritta in chiaro nel
    widget è testo non tradotto che nessuno vede mancare;
-   la grafica **decorativa** si toglie esplicitamente dall'albero semantico. Un'icona accanto
    a un testo che dice già tutto non va etichettata: va nascosta, altrimenti raddoppia
    l'annuncio.

Anche qui c'è una parte deterministica: che un componente condiviso di sola icona **esponga un
parametro di etichetta obbligatorio** è un pattern verificabile nel sorgente, ed è il modo di
rendere impossibile la dimenticanza invece di cercarla dopo.

## 🧭 Struttura e focus

-   **L'ordine di lettura segue l'ordine visivo.** Chi naviga da tastiera o con un lettore
    attraversa l'interfaccia nell'ordine in cui i suoi elementi sono dichiarati: quando quello
    non coincide con quello che si vede, l'unico modo di capire dove si è è indovinare. Se il
    layout mette a destra qualcosa che nel codice viene prima, è il codice a doversi riordinare.
-   **Lo stato di focus è sempre visibile**, su ogni elemento che può riceverlo, con il **3:1**
    di contrasto della tabella sopra. Rimuovere l'indicatore di focus perché «non è nel design»
    rende l'app inutilizzabile da tastiera: se il disegno non lo prevede, il disegno è
    incompleto.
-   **Niente trappole.** Ogni contenitore che cattura il focus — dialogo, pannello, foglio
    modale — dichiara dove il focus entra all'apertura e dove torna alla chiusura, e si può
    sempre chiudere con il gesto standard della piattaforma. La verifica è una sola: entrare
    con la tastiera e uscirne senza toccare il puntatore.
-   **Un titolo per schermata**, e una gerarchia di intestazioni che non salta livelli: è
    l'indice con cui chi non vede la pagina si fa un'idea di cosa contiene.

## 🚦 Non solo colore

**Nessuna informazione è veicolata dal solo colore.** Uno stato, un errore, una serie in un
grafico, un elemento selezionato: il colore è un rinforzo, mai l'unico segnale. Accanto ci va
sempre un secondo canale — un'icona, un'etichetta, una forma, una posizione.

La verifica sta in una domanda e in un comando, entrambi a costo zero: *si riesce a descrivere
questo stato senza nominare un colore?* Se la risposta è no, manca il secondo canale. E una
schermata in scala di grigi mostra la stessa cosa in un secondo.

È il criterio che salta fuori più spesso nei grafici e negli elenchi di stato, dove il colore
sembra bastare perché chi lo disegna sa già cosa significa.

## 🎞️ Movimento

Ogni sistema ha una preferenza per le **animazioni ridotte**, e l'app la legge invece di
ignorarla: chi l'ha attivata lo ha fatto perché il movimento gli dà fastidio o lo disorienta,
non per estetica.

Cosa vuol dire, in pratica:

-   con la preferenza attiva, le transizioni **non si accorciano: si annullano**. Una
    dissolvenza al posto di uno scorrimento, o niente del tutto. Quello che sparisce è il
    movimento nello spazio — parallasse, traslazioni ampie, rimbalzi, zoom — non il
    cambiamento di stato;
-   **niente si muove da solo per più di 5 secondi** senza un modo di fermarlo: caroselli,
    testi scorrevoli, animazioni in ciclo;
-   **niente lampeggia più di 3 volte al secondo.** Questa non è una preferenza: è la soglia
    oltre la quale un'animazione può scatenare una crisi, e vale sempre, anche senza la
    preferenza attiva.

La lettura della preferenza è una riga che sta nel componente condiviso di animazione, non in
ogni schermata: se ogni chiamante deve ricordarsene, qualcuno non lo farà.

## ✅ Cosa può diventare un test

Non tutto: una parte di questa regola richiede giudizio e resta una regola scritta. Ma la
parte che ha un numero è difendibile da sola, e va difesa — nello stesso commit in cui il
criterio entra nel progetto, secondo il criterio di
[Test di Architettura](./architecture_tests.rule.md).

| Criterio | Come si difende |
|---|---|
| contrasto delle coppie di token, in ogni modalità del tema | test: calcola il rapporto sulle coppie dichiarate e fallisce sotto la soglia |
| bersaglio minimo dei componenti interattivi condivisi | test di widget: misura il riquadro del bersaglio, non quello del disegno |
| etichetta obbligatoria sui componenti di sola icona | test sul sorgente: il parametro esiste ed è obbligatorio |
| valori di stile letterali al posto dei token | già coperto dai test di architettura dello stack |
| ordine di lettura, qualità dell'etichetta, secondo canale oltre il colore | giudizio: restano regole scritte, si verificano in revisione |

Vale l'avvertenza di sempre: un test che scandisce un insieme vuoto — nessuna coppia
dichiarata, nessun componente trovato — passa senza aver verificato niente. Prima di tutto si
asserisce che qualcosa è stato letto.

## 🔗 Materializzazioni per stack

-   **Flutter / Dart**:
    Widget accessibile — etichetta
    obbligatoria, bersaglio da 48 intorno a un'icona più piccola, stato che non si affida al
    solo colore e animazione che rispetta la preferenza di sistema.
-   **React Native / TypeScript**:
    Componente accessibile — gli
    stessi criteri con ruoli ed etichette dichiarati uno per uno, perché non c'è un albero HTML
    da cui dedurli, e la dichiarazione esplicita di quale criterio su un'app touch non si
    applica.
-   **React web / TypeScript**:
    [Componente accessibile](../../react_web/examples/accessible_component.example.md) — gli
    stessi criteri a partire dall'elemento HTML giusto, con il focus visibile nei fogli, il
    salto della navigazione, il focus riportato sul contenuto dopo un cambio di pagina e le
    coppie di contrasto dichiarate per nome di token.

Un nuovo stack aggiunge il proprio esempio qui sotto e non riscrive i criteri: i numeri di
questa regola non cambiano con il linguaggio.

## 🔗 Correlate

-   [Test di Architettura](./architecture_tests.rule.md): la forma con cui i criteri numerici
    di questa regola smettono di essere un promemoria.
-   Convenzioni UI: dove i token di tema, a cui
    contrasto e dimensioni fanno riferimento, sono dichiarati per lo stack Flutter.
