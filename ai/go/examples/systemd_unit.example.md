# Unit systemd

Il servizio che avvia il binario sulla macchina di destinazione, versionato nel repo (di norma
`deploy/<binario>.service`) e installato in `/etc/systemd/system/`. Vedi
[Distribuzione](../rules/deployment.rule.md).

```ini
# deploy/<binario>.service

[Unit]
Description=<descrizione del servizio>
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/<binario>
# Configurazione e segreti: KEY=VALUE, fuori dal repo, permessi 0600.
EnvironmentFile=/etc/<binario>/<binario>.env

# Utente effimero senza privilegi; i dati in /var/lib/<binario>, che systemd crea e
# assegna a quell'utente. Il binario riceve il percorso in $STATE_DIRECTORY.
DynamicUser=yes
StateDirectory=<binario>

# Il tetto vero lo fa rispettare il kernel; GOMEMLIMIT, un po' sotto, dice al garbage
# collector di lavorare di più prima di arrivarci invece di farsi uccidere.
MemoryMax=128M
Environment=GOMEMLIMIT=100MiB

Restart=on-failure
RestartSec=5s
# SIGTERM è il segnale che il binario gestisce per lo shutdown ordinato.
KillSignal=SIGTERM
TimeoutStopSec=30s

NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
PrivateDevices=yes

[Install]
WantedBy=multi-user.target
```

Installazione e aggiornamento:

```bash
sudo install -m 0755 dist/<binario> /usr/local/bin/<binario>
sudo install -m 0644 deploy/<binario>.service /etc/systemd/system/
sudo install -d -m 0750 /etc/<binario>
sudo install -m 0600 deploy/<binario>.env.example /etc/<binario>/<binario>.env   # solo la prima volta
sudo systemctl daemon-reload
sudo systemctl enable --now <binario>
sudo systemctl restart <binario>                     # dopo ogni aggiornamento del binario
curl -fsS http://127.0.0.1:<porta>/healthz           # deve stampare la versione appena installata
journalctl -u <binario> -f                           # il log, che slog scrive su stderr
```

-   **`MemoryMax`** è un valore da tarare sul progetto: si parte stretti e si alza guardando
    `systemctl status <binario>` (riga `Memory:`) dopo qualche giorno di uso.
-   **Il binario ascolta su `127.0.0.1`** se davanti c'è un proxy o un tunnel che termina TLS:
    la porta non deve essere raggiungibile dalla rete senza passare da lì.
