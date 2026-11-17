La tabella dei divieti è la parte di un test di architettura che non dipende dal linguaggio:
la regola scritta una volta sola, in cima al file, leggibile senza eseguire niente (vedi
[Test di Architettura](../rules/architecture_tests.rule.md)). Cambia solo la sintassi con cui
la si dichiara.

Per un layout a livelli condivisi, con i nomi usati dagli stack di questo framework:

| Livello       | Import vietati                                        |
| ------------- | ----------------------------------------------------- |
| `domain/`     | il framework UI, `data/`, `features/`, `bootstrap/`   |
| `data/`       | `features/`, `shared/`, `bootstrap/`                  |
| `core/`       | `domain/`, `data/`, `features/`                       |
| `shared/`     | `core/`, `domain/`, `data/`, `features/`              |
| `features/<f>`| ogni altra `features/<g>`, `data/`                    |

`bootstrap/` non compare: è il composition root, l'unico livello che può importare tutto.
I nomi dei livelli e la tabella esatta li dichiara lo stack — questa è la forma, non il
contenuto.

Tradotta in un test, sempre in tre passi:

```
divieti = {
  "<percorso del livello>": ["<prefisso di import vietato>", ...],
  ...
}

per ogni (livello, vietati) in divieti:
    test "<livello> non importa <vietati>":
        file = tutti i sorgenti sotto <livello>, ricorsivamente

        # Una cartella vuota — un livello rinominato, un percorso sbagliato — farebbe
        # passare il test senza che abbia letto una riga.
        asserisci che file non è vuoto

        per ogni f in file:
            per ogni riga di import in f:
                per ogni v in vietati:
                    asserisci che la riga non contiene v
                    # Il messaggio nomina f e v: chi legge il fallimento deve sapere
                    # quale riga cancellare, non dove cercarla.
```

Stessa struttura per un pattern proibito dentro un file: cambia solo cosa si legge — il
sorgente intero invece delle sole righe di import — e la lista dei simboli vietati.

La tabella non cambia quando un confine passa **dentro** un livello invece che fra due — una
dipendenza tecnica ammessa solo in certe cartelle, per esempio. Lì si rovescia: invece di una
lista di divieti per livello si dichiara la lista dei **percorsi ammessi**, e si scorrono
tutti i sorgenti saltando quelli che ci ricadono. Il resto dei tre passi è identico.

Le versioni eseguibili sono in
Test di purezza dei livelli per
Dart/Flutter, in
Test di architettura per Python, in
[Test di architettura](../../go/examples/architecture_test.example.md) per Go, e in
Confini fra livelli per
React Native e nel suo gemello [per il web](../../react_web/examples/architecture_test.example.md)
— dove la tabella diventa configurazione del linter invece che test.
