// SOURCE: backend (do not edit locally)
// See: API_CONTRACT.md

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    email: string;
  };
}

export interface MeResponse {
  id: string;
  email: string;
  name: string;
}
