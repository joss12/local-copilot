import * as vscode from "vscode";
import * as path from "path";
import { spawn, ChildProcess } from "child_process";
import axios from "axios";

export class BackendManager {
  private backendProcess: ChildProcess | null = null;
  private backendPort: number;
  private backendPath: string;

  constructor() {
    const config = vscode.workspace.getConfiguration("localCopilot");
    this.backendPort = config.get<number>("backendPort", 8089);
    this.backendPath = config.get<string>("backendPath", "");
  }

  public getBaseUrl(): string {
    return `http://localhost:${this.backendPort}`;
  }

  public async start(workspacePath: string): Promise<boolean> {
    if (this.backendProcess) {
      vscode.window.showInformationMessage("Backend is already  running");
      return true;
    }

    //Find backend binary
    const binaryPath = this.findBackendBinary(workspacePath);
    if (!binaryPath) {
      vscode.window.showErrorMessage("Backend binary not found.");
      return false;
    }

    try {
      //Start backend process
      this.backendProcess = spawn(
        binaryPath,
        ["--workspace", workspacePath, "--port", this.backendPort.toString()],
        {
          cwd: path.dirname(binaryPath),
        },
      );

      //handle stdout
      this.backendProcess.stdout?.on("data", (data) => {
        console.log(`[Backend Error] ${data}`);
      });

      //handle stderr
      this.backendProcess.stderr?.on("data", (data) => {
        console.log(`[Backend Error] ${data}`);
      });

      //Handle exit
      this.backendProcess.on("exit", (code) => {
        console.log(`Backend existed with code ${code}`);
        this.backendProcess = null;
      });

      //wait for backend to be ready
      await this.waitForBacckend();

      vscode.window.showInformationMessage("Local Copilot backend started");
      return true;
    } catch (err) {
      vscode.window.showErrorMessage(`Failed to start backend: ${err}`);
      return false;
    }
  }

  public async stop(): Promise<void> {
    if (this.backendProcess) {
      this.backendProcess.kill();
      this.backendProcess = null;
      vscode.window.showInformationMessage("Local Copilot backend stopped");
    }
  }

  public async isRunning(): Promise<boolean> {
    try {
      const response = await axios.get(`${this.getBaseUrl()}/health`, {
        timeout: 1000,
      });
      return response.status === 200;
    } catch {
      return false;
    }
  }

  private findBackendBinary(workspacePath: string): string | null {
    //If user specified path in settings, use that
    if (this.backendPath) {
      return this.backendPath;
    }

    //Try to find in workspace
    const possiblePaths = [
      path.join(workspacePath, "backend", "copilot"),
      path.join(workspacePath, "..", "backend", "copilot"),
      path.join(workspacePath, "copilot"),
    ];

    for (const p of possiblePaths) {
      try {
        if (require("fs").existsSync(p)) {
          return p;
        }
      } catch {
        continue;
      }
    }

    return null;
  }

  private async waitForBacckend(maxAttemps = 10): Promise<void> {
    for (let i = 0; i < maxAttemps; i++) {
      if (await this.isRunning()) {
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, 500));
    }
    throw new Error("Backend failed to start in time");
  }
}
