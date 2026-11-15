// Chiamarla in un ramo di default costringe il compilatore a verificare che quel ramo sia
// irraggiungibile: quando l'insieme cresce, il codice smette di compilare qui invece di
// ricadere in silenzio su un comportamento di ripiego.
export function assertNever(value: never): never {
  throw new Error(`caso non gestito: ${JSON.stringify(value)}`);
}
