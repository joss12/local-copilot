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
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const backendManager_1 = require("./backendManager");
const completionProvider_1 = require("./completionProvider");
const axios_1 = __importDefault(require("axios"));
let backendManager;
let completionProvider;
async function activate(context) {
    console.log("Local Copilot extension is now active");
    backendManager = new backendManager_1.BackendManager();
    // Register commands
    context.subscriptions.push(vscode.commands.registerCommand("local-copilot.startBackend", async () => {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (!workspaceFolders) {
            vscode.window.showErrorMessage("No workspace folder open");
            return;
        }
        const workspacePath = workspaceFolders[0].uri.fsPath;
        const started = await backendManager.start(workspacePath);
        if (started) {
            // Register completion provider
            registerCompletionProvider(context);
        }
    }));
    context.subscriptions.push(vscode.commands.registerCommand("local-copilot.stopBackend", async () => {
        await backendManager.stop();
        // Unregister completion provider
        if (completionProvider) {
            completionProvider.dispose();
            completionProvider = undefined;
        }
    }));
    context.subscriptions.push(vscode.commands.registerCommand("local-copilot.indexWorkspace", async () => {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (!workspaceFolders) {
            vscode.window.showErrorMessage("No workspace folder open");
            return;
        }
        // Check if backend is running
        const isRunning = await backendManager.isRunning();
        if (!isRunning) {
            const answer = await vscode.window.showWarningMessage("Backend is not running. Start it now?", "Yes", "No");
            if (answer === "Yes") {
                const started = await backendManager.start(workspaceFolders[0].uri.fsPath);
                if (!started) {
                    return;
                }
                registerCompletionProvider(context);
            }
            else {
                return;
            }
        }
        // Trigger indexing
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Indexing workspace...",
            cancellable: false,
        }, async (progress) => {
            try {
                const response = await axios_1.default.post(`${backendManager.getBaseUrl()}/index`, { workspace_path: workspaceFolders[0].uri.fsPath }, { timeout: 60000 });
                const message = `Indexed ${response.data.files_processed} files, found ${response.data.symbols_found} symbols`;
                vscode.window.showInformationMessage(message);
                if (response.data.errors && response.data.errors.length > 0) {
                    console.warn("Indexing errors:", response.data.errors);
                }
            }
            catch (error) {
                vscode.window.showErrorMessage(`Indexing failed: ${error}`);
            }
        });
    }));
    // Auto-start backend if configured
    const workspaceFolders = vscode.workspace.workspaceFolders;
    if (workspaceFolders) {
        const isRunning = await backendManager.isRunning();
        if (!isRunning) {
            vscode.window.showInformationMessage('Local Copilot: Backend not running. Use "Local Copilot: Start Backend" to enable suggestions.');
        }
        else {
            registerCompletionProvider(context);
        }
    }
}
function registerCompletionProvider(context) {
    if (completionProvider) {
        return; // Already registered
    }
    const provider = new completionProvider_1.LocalCopilotCompletionProvider(backendManager.getBaseUrl());
    completionProvider = vscode.languages.registerInlineCompletionItemProvider([
        { scheme: "file", language: "javascript" },
        { scheme: "file", language: "javascriptreact" },
        { scheme: "file", language: "typescript" },
        { scheme: "file", language: "typescriptreact" },
        { scheme: "file", language: "go" },
    ], provider);
    context.subscriptions.push(completionProvider);
    console.log("Completion provider registered");
}
function deactivate() {
    if (backendManager) {
        backendManager.stop();
    }
}
//# sourceMappingURL=extension.js.map