export type UseCase<In, Out> = (input: In) => Out;

// La variante asincrona esiste per gli use case che ricevono una porta: l'I/O è della porta,
// ma l'attesa attraversa lo use case e deve comparire nella sua firma.
export type AsyncUseCase<In, Out> = (input: In) => Promise<Out>;
