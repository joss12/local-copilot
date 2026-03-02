export interface IndexRequest {
  workspace_path: string;
}

export interface IndexResponse {
  files_processed: number;
  symbols_found: number;
  errors?: string[];
  duration: string;
}

export interface SuggestionRequest {
  file_path: string;
  language: string;
  current_line: number;
  current_column: number;
  context_before: string;
  context_after: string;
  partial_symbol: string;
}

export interface Suggestion {
  text: string;
  type: string;
  source: string;
  confidence: number;
  line_number: number;
}

export interface SuggestionResponse {
  suggestions: Suggestion[];
  duration: string;
}
