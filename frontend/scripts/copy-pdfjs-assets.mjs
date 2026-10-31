// Copia para public/pdfjs os arquivos que o pdf.js busca em tempo de execução: os decodificadores em WebAssembly (JBIG2, que também
// decodifica o CCITT fax dos PDFs de scanner, JPEG 2000 e perfis de cor ICC), os CMaps e as fontes padrão. Sem eles, páginas de
// scanner saem em branco. Roda antes do dev e do build; public/pdfjs não vai para o Git (ver .gitignore).
import { cpSync, existsSync, mkdirSync, rmSync } from 'node:fs';
import { createRequire } from 'node:module';
import path from 'node:path';

const require = createRequire(import.meta.url);
const root = path.dirname(require.resolve('pdfjs-dist/package.json'));
const out = path.resolve(import.meta.dirname, '../public/pdfjs');

rmSync(out, { recursive: true, force: true });
mkdirSync(out, { recursive: true });
for (const dir of ['wasm', 'cmaps', 'standard_fonts']) {
  const from = path.join(root, dir);
  if (!existsSync(from)) throw new Error(`pdfjs-dist sem a pasta ${dir}: o visualizador de PDF não funcionará.`);
  cpSync(from, path.join(out, dir), { recursive: true });
}
