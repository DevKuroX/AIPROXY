// SOURCE: backend (do not edit locally)
// See: API_CONTRACT.md

export interface Alias {
  // TODO: Define Alias shape per backend implementation
  id: string;
  name: string;
  value: string;
}

export interface AliasesResponse {
  aliases: Alias[];
}
