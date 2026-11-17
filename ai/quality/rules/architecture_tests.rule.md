# Test di Architettura

## 🎯 Una regola che nessuno verifica è un suggerimento

Le regole di dipendenza di questo framework si possono controllare con un `grep`. Il
problema del `grep` è che va ricordato: funziona finché qualcuno, prima di aprire una pull
request, si ferma e lo lancia. Con un agente che lavora senza presidio non se lo ricorda
nessuno, e una regola che vale solo quando qualcuno la controlla non è una regola — è un
suggerimento.

Un **test di architettura** chiude la distanza: legge i file del progetto, cerca il divieto
e fallisce come un qualunque altro test. Il `grep` resta comodo per una verifica al volo
mentre si lavora; la garanzia è il test, perché gira con la suite e non si dimentica.

## 🧱 Cosa va coperto, sempre

Tre famiglie, in qualsiasi progetto e qualsiasi stack:

1.  **I confini fra livelli.** Per ogni livello, la lista degli import che non può contenere.
    È la traduzione diretta della tabella delle dipendenze dichiarata dallo stack.
2.  **L'indipendenza fra feature.** Nessuna feature importa un'altra feature: quello che
    condividono passa dal dominio, non da un import laterale.
3.  **L'assenza di valori di stile scritti a mano nei componenti condivisi.** Colori
    letterali, misure, accessi al tema che scavalcano i token: nei componenti condivisi non
    ci vanno, e un test che cerca quei simboli nel sorgente lo dimostra.

Le prime due sono divieti di import, la terza è un pattern proibito dentro un file. Sono le
due forme che un test di architettura sa esprimere — e sono anche il criterio con cui se ne
aggiunge uno nuovo.

## 📁 Dove vivono e come si chiamano

I test di architettura stanno **tutti insieme**, in una cartella dedicata sotto la radice
dei test del progetto:

```
<radice dei test>/architecture/
```

Un file per famiglia di regole, non uno per regola: il nome dice quale regola difende, con
la convenzione di test dello stack (`layer_purity_test.dart`, `test_layer_purity.py`). Una
cartella con lo stesso nome in ogni progetto rende questi test riconoscibili a colpo
d'occhio, e distingue ciò che difende l'architettura da ciò che verifica il comportamento.

Girano con **la suite normale**, senza un comando dedicato e senza un tag che li escluda:
un test di architettura che si lancia a parte è tornato a essere un `grep`. E la suite gira
nella CI (Dove gira la verifica): è quello che
chiude il cerchio aperto in cima a questa regola — non c'è più nessuno che debba ricordarsi
di lanciarla.

## 🧪 Come è fatto

Sempre gli stessi tre passi, in qualunque linguaggio:

1.  **Dichiarare la tabella dei divieti** in cima al file — livello → import vietati. È la
    regola scritta una volta sola, in un posto dove si legge senza eseguire niente.
2.  **Enumerare i file** del livello, ricorsivamente, e leggerne gli import (o il sorgente,
    per i pattern proibiti).
3.  **Asserire**, con un messaggio che nomina il file colpevole e l'import che non doveva
    esserci: chi legge il fallimento deve sapere cosa cancellare, non dove cercare.

Il passo che si dimentica è un altro: **fallire se non si è letto niente.** Un test che
scandisce una cartella vuota — perché il layout è cambiato, perché il percorso è sbagliato —
passa senza aver verificato nulla, ed è peggio di non averlo. Ogni scansione asserisce prima
di tutto che l'insieme dei file trovati non è vuoto.

Vedi [Tabella degli import vietati](../examples/forbidden_imports_table.example.md) per la
forma neutra, e le materializzazioni per stack più sotto.

## 🔧 Quando lo strumento migliore non è un test

In alcuni ecosistemi il **linter** sa già imporre un divieto di import per cartella. Dove
esiste, è lo strumento da preferire: dice la stessa cosa del test, ma la dice sulla riga
dell'import mentre la si scrive, invece che dopo. Valgono due condizioni, entrambe non
negoziabili — la configurazione è **committata nel progetto**, non impostata nell'editor di
qualcuno, e il linter **gira nella CI**, altrimenti il fallimento è un suggerimento che si
supera premendo «commit».

Resta comunque un test, più piccolo: un linter i cui percorsi non corrispondono più a nessun
file non protesta: tace. È lo stesso difetto della cartella vuota, e si chiude allo stesso
modo — asserendo che c'è ancora qualcosa da controllare.

## ➕ Quando se ne aggiunge uno

Il criterio è meccanico: **ogni volta che una regola nuova è esprimibile come divieto di
import o come pattern proibito in un file, nasce con il suo test.** Non dopo, non in un task
a parte — nello stesso commit della regola, altrimenti non arriva mai.

Se una regola non è esprimibile in quelle due forme (una convenzione di naming sfumata, una
scelta di design che richiede giudizio), non si forza: resta una regola scritta, e si evita
un test che approssima e produce falsi allarmi.

## 💥 Quando è la regola a dover cambiare

Un test di architettura che fallisce dice una di due cose: il codice ha attraversato un
confine, oppure il confine era sbagliato. Il primo caso si sistema nel codice. Il secondo è
quello delicato.

Quando è la regola a dover cambiare, **si cambiano la regola e il test nello stesso commit**,
e si scrive nel messaggio perché il confine si è spostato. Mai il test da solo: togliere una
riga dalla tabella dei divieti per far passare la suite è il modo in cui un'architettura si
sfalda senza che nessuno se ne accorga — il diff sembra una manutenzione, e la regola scritta
continua a dire il contrario di quello che il progetto fa davvero. Se lo spostamento del
confine è una scelta strutturale, merita anche un decision record
(Decision Record).

## 🚧 Questa non è una regola sui test

Qui si parla **solo** dei test che difendono l'architettura: import, confini, pattern
proibiti. Cosa vada testato del comportamento del prodotto, con quale copertura e con quali
fake, è un'altra questione e la decide lo stack.

## 🔗 Materializzazioni per stack

-   **Flutter / Dart**:
    Test di purezza dei livelli —
    la tabella dei divieti, l'indipendenza fra feature e i componenti condivisi senza colori
    letterali, in due file sotto `app/test/architecture/`.
-   **Python**: Test di architettura —
    la stessa tabella letta sull'albero sintattico dei moduli invece che sul testo, e i
    costrutti vietati dentro un router, sotto `backend/tests/architecture/`.
-   **React Native / TypeScript**:
    Confini fra livelli — la stessa
    tabella come configurazione del linter invece che come test, perché lì gli import li legge
    già qualcuno; il test resta per il buco che il linter ha su sé stesso, cioè il glob che non
    corrisponde più a nessun file.
-   **Go**: [Test di architettura](../../go/examples/architecture_test.example.md) — la stessa
    tabella letta con `go/parser` sull'intestazione dei file, in un package di solo test sotto
    `backend/architecture/`: il compilatore vieta i cicli di import, non le dipendenze nella
    direzione sbagliata.
-   **React web / TypeScript**:
    [Confini fra livelli](../../react_web/examples/architecture_test.example.md) — la tabella
    di React Native come configurazione di ESLint, con un livello in più (`src/app/`, che di
    una feature vede solo le rotte), e stylelint per il divieto dei valori scritti a mano, che
    sul web sta nei fogli CSS e non nei componenti.
