// Ciò che l'installazione rende disponibile, e che cambia quello che l'app mostra.
export type Installation = {
  readonly version: string;
  // Gli strumenti di sviluppo sono attivi.
  readonly dev: boolean;
  // C'è la chiave API: un canale si può aggiungere dal suo @handle.
  readonly handleLookup: boolean;
};
