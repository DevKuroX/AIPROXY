// SOURCE: backend (do not edit locally)
// See: API_CONTRACT.md

export interface OAuthProvider {
  // TODO: Define OAuthProvider shape per backend implementation
  id: string;
  name: string;
  provider: string;
  created_at: string;
}

export interface OAuthProvidersResponse {
  providers: OAuthProvider[];
}
