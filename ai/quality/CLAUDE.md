# Framework AI — Garanzie di Qualità

Questo file è l'indice della cartella: elenca regole ed esempi su come le regole
architetturali smettono di essere un promemoria e diventano qualcosa che fallisce da solo.
Nessun contenuto sostanziale va scritto qui — solo link.

È materiale **trasversale**: non appartiene a nessuno stack e vale per qualsiasi progetto che
aggancia il framework, qualunque tecnologia usi. Cambia la sintassi con cui si scrive un
test; non cambia quali confini vanno difesi.

## 🎯 Il principio

Una regola di dipendenza che nessuno verifica automaticamente è un suggerimento. Le regole
degli stack si chiudono spesso con un `grep` da lanciare a mano: resta utile come verifica al
volo, ma la garanzia è il test, perché gira con la suite e non si dimentica.

## 🔌 Come è agganciato a questo repo

Questa cartella è una copia tenuta dentro il repo: il `CLAUDE.md` alla radice la importa con
`@ai/quality/CLAUDE.md`. Quello che è specifico del progetto — la radice dei test, i nomi dei
livelli, il comando con cui si lancia la suite — resta nel `CLAUDE.md` alla radice e nelle
regole dello stack.

## 📖 Regole

-   [Test di Architettura](./rules/architecture_tests.rule.md): cosa va coperto sempre, dove vivono i test e come si chiamano, quando se ne aggiunge uno e cosa fare quando è la regola a dover cambiare.
-   [Accessibilità](./rules/accessibility.rule.md): i criteri che rendono un'interfaccia usabile — contrasto, bersagli tattili, etichette, ordine di lettura e focus, movimento — con i numeri e con quello che di essi può diventare un test.

## 📦 Esempi

-   [Tabella degli import vietati](./examples/forbidden_imports_table.example.md): la forma neutra di un test di architettura — la tabella dei divieti e i tre passi che la eseguono.

Le versioni eseguibili vivono nello stack che le parla — per Go
[Test di architettura](../go/examples/architecture_test.example.md), per React web
[Confini fra livelli](../react_web/examples/architecture_test.example.md) e
[Componente Accessibile](../react_web/examples/accessible_component.example.md).
