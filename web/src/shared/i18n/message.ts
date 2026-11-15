import { assertNever } from '@core/exhaustive';
import type { Failure } from '@core/result/failure';
import type { TranslationKey } from './keys';
import type { Translate } from './use-translate';

// Un testo da mostrare: un'etichetta del catalogo, oppure parole che arrivano già scritte
// dal server. È ciò che uno store tiene al posto di una `Failure`.
export type Message =
  | { readonly key: TranslationKey; readonly params?: Record<string, string | number> }
  | { readonly text: string };

export const renderMessage = (t: Translate, message: Message): string =>
  'text' in message ? message.text : t(message.key, message.params);

// L'ultimo punto in cui il vocabolario delle `Failure` esiste: sopra di qui si parla solo di
// cosa mostrare. Sta in `shared/` perché la traduzione è la stessa per ogni feature.
export function describeFailure(failure: Failure): Message {
  switch (failure.kind) {
    case 'network':
      return { key: 'common_error_network' };
    case 'notFound':
      return { key: 'common_error_not_found' };
    case 'rejected':
      // Il server ha già scritto il perché nella lingua di chi usa l'app.
      return { text: failure.message };
    case 'parse':
    case 'storage':
      // Un dato illeggibile e uno storage rotto sono lo stesso fatto per chi guarda: non si
      // può fare niente, e riprovare non cambia nulla.
      return { key: 'common_error_unexpected' };
    default:
      return assertNever(failure);
  }
}
