import * as vscode from "vscode";
import { BackendManager } from "./backendManager";
import { LocalCopilotCompletionProvider } from "./completionProvider";
import axios from "axios";
import { IndexResponse } from "./types";

let backendManager: BackendManager;
let completionProvider: vscode.Disposable | undefined;

export async function activate(context: vscode.ExtensionContext) {
  console.log("Local Copilot extension is now active");

  backendManager = new BackendManager();

  // Register commands
  context.subscriptions.push(
    vscode.commands.registerCommand("local-copilot.startBackend", async () => {
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
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("local-copilot.stopBackend", async () => {
      await backendManager.stop();

      // Unregister completion provider
      if (completionProvider) {
        completionProvider.dispose();
        completionProvider = undefined;
      }
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand(
      "local-copilot.indexWorkspace",
      async () => {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (!workspaceFolders) {
          vscode.window.showErrorMessage("No workspace folder open");
          return;
        }

        // Check if backend is running
        const isRunning = await backendManager.isRunning();
        if (!isRunning) {
          const answer = await vscode.window.showWarningMessage(
            "Backend is not running. Start it now?",
            "Yes",
            "No",
          );

          if (answer === "Yes") {
            const started = await backendManager.start(
              workspaceFolders[0].uri.fsPath,
            );
            if (!started) {
              return;
            }
            registerCompletionProvider(context);
          } else {
            return;
          }
        }

        // Trigger indexing
        await vscode.window.withProgress(
          {
            location: vscode.ProgressLocation.Notification,
            title: "Indexing workspace...",
            cancellable: false,
          },
          async (progress) => {
            try {
              const response = await axios.post<IndexResponse>(
                `${backendManager.getBaseUrl()}/index`,
                { workspace_path: workspaceFolders[0].uri.fsPath },
                { timeout: 60000 },
              );

              const message = `Indexed ${response.data.files_processed} files, found ${response.data.symbols_found} symbols`;
              vscode.window.showInformationMessage(message);

              if (response.data.errors && response.data.errors.length > 0) {
                console.warn("Indexing errors:", response.data.errors);
              }
            } catch (error) {
              vscode.window.showErrorMessage(`Indexing failed: ${error}`);
            }
          },
        );
      },
    ),
  );

  // Auto-start backend if configured
  const workspaceFolders = vscode.workspace.workspaceFolders;
  if (workspaceFolders) {
    const isRunning = await backendManager.isRunning();
    if (!isRunning) {
      vscode.window.showInformationMessage(
        'Local Copilot: Backend not running. Use "Local Copilot: Start Backend" to enable suggestions.',
      );
    } else {
      registerCompletionProvider(context);
    }
  }
}

function registerCompletionProvider(context: vscode.ExtensionContext) {
  if (completionProvider) {
    return; // Already registered
  }

  const provider = new LocalCopilotCompletionProvider(
    backendManager.getBaseUrl(),
  );

  completionProvider = vscode.languages.registerInlineCompletionItemProvider(
    [
      { scheme: "file", language: "javascript" },
      { scheme: "file", language: "javascriptreact" },
      { scheme: "file", language: "typescript" },
      { scheme: "file", language: "typescriptreact" },
      { scheme: "file", language: "go" },
    ],
    provider,
  );

  context.subscriptions.push(completionProvider);
  console.log("Completion provider registered");
}

export function deactivate() {
  if (backendManager) {
    backendManager.stop();
  }
}
