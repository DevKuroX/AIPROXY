// SOURCE: backend (do not edit locally)
// See: API_CONTRACT.md

export interface Combo {
  // TODO: Define Combo shape per backend implementation
  id: string;
  name: string;
  providers: string[];
}

export interface CombosResponse {
  combos: Combo[];
}
