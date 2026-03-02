import * as vscode from "vscode";
import axios from "axios";
import { SuggestionRequest, SuggestionResponse } from "./types";

export class LocalCopilotCompletionProvider
  implements vscode.InlineCompletionItemProvider
{
  private backendUrl: string;

  constructor(backendUrl: string) {
    this.backendUrl = backendUrl;
  }

  async provideInlineCompletionItems(
    document: vscode.TextDocument,
    position: vscode.Position,
    context: vscode.InlineCompletionContext,
    token: vscode.CancellationToken,
  ): Promise<
    vscode.InlineCompletionItem[] | vscode.InlineCompletionList | null
  > {
    //Only provide suggestions for supported languages
    const language = this.mapLanguage(document.languageId);
    if (!language) {
      return null;
    }

    //Get current word being typed
    const wordRange = document.getWordRangeAtPosition(position);
    const partialSymbol = wordRange ? document.getText(wordRange) : "";

    if (partialSymbol.length < 2) {
      return null;
    }

    const lineText = document.lineAt(position.line).text;
    const contextBefore = lineText.substring(0, position.character);
    const contextAfter = lineText.substring(position.character);

    // Build request
    const request: SuggestionRequest = {
      file_path: document.uri.fsPath,
      language: language,
      current_line: position.line,
      current_column: position.character,
      context_before: contextBefore,
      context_after: contextAfter,
      partial_symbol: partialSymbol,
    };

    try {
      //Call backend API
      const response = await axios.post<SuggestionResponse>(
        `${this.backendUrl}/suggest`,
        request,
        { timeout: 2000 },
      );

      const items = response.data.suggestions.map((suggestion) => {
        const completionText = this.extractCompletionText(
          suggestion.text,
          suggestion.type,
        );

        return new vscode.InlineCompletionItem(
          completionText,
          new vscode.Range(position, position),
        );
      });
      return items;
    } catch (err) {
      console.error("Failed to get suggestions:", err);
      return null;
    }
  }

  private mapLanguage(vscodeLanguageId: string): string | null {
    const languageMap: { [key: string]: string } = {
      javascript: "javascript",
      javascriptreact: "javascript",
      typescript: "typescript",
      typescriptreact: "typescript",
      go: "go",
    };
    return languageMap[vscodeLanguageId] || null;
  }

  private extractCompletionText(fullText: string, type: string): string {
    if (type === "function") {
      const match = fullText.match(/func\s+(?:\(.*?\)\s+)?(\w+\(.*?\))/);
      if (match) {
        return match[1];
      }
    }

    return fullText;
  }
}
