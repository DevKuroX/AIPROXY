// SOURCE: backend (do not edit locally)
// See: API_CONTRACT.md

// StreamChunk represents normalized backend SSE stream chunks
// Format: { id: string; object: string; created: number; type: string; data: string }
export interface StreamChunk {
  id: string;
  object: string;
  created: number;
  type: string;
  data: string;
}
