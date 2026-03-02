"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.LocalCopilotCompletionProvider = void 0;
const vscode = __importStar(require("vscode"));
const axios_1 = __importDefault(require("axios"));
class LocalCopilotCompletionProvider {
    constructor(backendUrl) {
        this.backendUrl = backendUrl;
    }
    async provideInlineCompletionItems(document, position, context, token) {
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
        const request = {
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
            const response = await axios_1.default.post(`${this.backendUrl}/suggest`, request, { timeout: 2000 });
            const items = response.data.suggestions.map((suggestion) => {
                const completionText = this.extractCompletionText(suggestion.text, suggestion.type);
                return new vscode.InlineCompletionItem(completionText, new vscode.Range(position, position));
            });
            return items;
        }
        catch (err) {
            console.error("Failed to get suggestions:", err);
            return null;
        }
    }
    mapLanguage(vscodeLanguageId) {
        const languageMap = {
            javascript: "javascript",
            javascriptreact: "javascript",
            typescript: "typescript",
            typescriptreact: "typescript",
            go: "go",
        };
        return languageMap[vscodeLanguageId] || null;
    }
    extractCompletionText(fullText, type) {
        if (type === "function") {
            const match = fullText.match(/func\s+(?:\(.*?\)\s+)?(\w+\(.*?\))/);
            if (match) {
                return match[1];
            }
        }
        return fullText;
    }
}
exports.LocalCopilotCompletionProvider = LocalCopilotCompletionProvider;
//# sourceMappingURL=completionProvider.js.map