// SOURCE: backend (do not edit locally)
// See: API_CONTRACT.md

export interface Provider {
  name: string;
  type: string;
}

export interface ProvidersResponse {
  providers: Provider[];
}
