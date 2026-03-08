import { cpSync, existsSync, mkdirSync, rmSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const webRoot = path.resolve(scriptDir, "..");
const sourceDir = path.resolve(webRoot, "..", "docs", "assets");
const targetDir = path.resolve(webRoot, "public", "assets");

if (!existsSync(sourceDir)) {
  throw new Error(`Kaynak assets dizini bulunamadi: ${sourceDir}`);
}

rmSync(targetDir, { recursive: true, force: true });
mkdirSync(path.dirname(targetDir), { recursive: true });
cpSync(sourceDir, targetDir, { recursive: true });

console.log(`Assets sync tamamlandi: ${sourceDir} -> ${targetDir}`);
