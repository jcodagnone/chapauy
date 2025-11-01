# ChapaUY

### Scripts disponibles

\`\`\`bash

# Desarrollo

pnpm dev # Inicia el servidor de desarrollo

# Build

pnpm build # Genera el build de producción
pnpm start # Ejecuta el build de producción

# Calidad de código

pnpm lint # Ejecuta ESLint
pnpm lint:fix # Corrige problemas de ESLint automáticamente
pnpm format # Formatea el código con Prettier
pnpm format:check # Verifica el formato sin modificar archivos
pnpm typecheck # Verifica tipos de TypeScript
\`\`\`

### Configuración del editor

#### Visual Studio Code

Instala las extensiones recomendadas:

- ESLint
- Prettier - Code formatter
- EditorConfig for VS Code

La configuración en `.vscode/settings.json` se aplicará automáticamente al abrir el proyecto.

#### Zed

Agrega esta configuración a tu `settings.json` de Zed:

\`\`\`json
{
"format_on_save": "on",
"formatter": "prettier",
"lsp": {
"eslint": {
"settings": {
"codeActionsOnSave": {
"source.fixAll.eslint": true
}
}
}
}
}
\`\`\`

#### Otros editores

Asegurate de tener soporte para:

- **EditorConfig** - Lee `.editorconfig` para configuración básica
- **Prettier** - Para formateo automático
- **ESLint** - Para linting de código

### Estándares de código

El proyecto usa:

- **Prettier** para formateo consistente
- **ESLint** con reglas de Next.js y TypeScript
- **EditorConfig** para configuración básica de editor

Antes de hacer commit, ejecuta:

\`\`\`bash
pnpm format && pnpm lint:fix && pnpm typecheck
\`\`\`

Para más detalles sobre convenciones y flujo de trabajo, consulta [CONTRIBUTING.md](./CONTRIBUTING.md)

## Tecnologías

- **Framework**: Next.js 15 (App Router)
- **Lenguaje**: TypeScript
- **Estilos**: Tailwind CSS v4
- **UI Components**: Radix UI + shadcn/ui
- **Formularios**: React Hook Form + Zod
- **Gestión de paquetes**: pnpm
