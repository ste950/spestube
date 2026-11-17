export type Channel = {
  readonly id: string;
  readonly title: string;
  // `null` se il canale è stato aggiunto dal suo id.
  readonly handle: string | null;
  // L'ultimo download del feed non è riuscito: i video di questo canale potrebbero mancare.
  readonly fetchFailed: boolean;
};
