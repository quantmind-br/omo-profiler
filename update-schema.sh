#!/bin/bash

# Script para atualizar o schema do oh-my-opencode
# Faz o download da versão mais recente do schema, gera um diff e substitui o arquivo local

set -e

URL="https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/refs/heads/dev/assets/oh-my-opencode.schema.json"
OUTPUT_FILE="oh-my-opencode-schema.json"
DIFF_FILE="oh-my-opencode-schema-diff.md"
TEMP_FILE=$(mktemp)

echo "Baixando schema de: $URL"

# Download para arquivo temporário
if command -v curl &> /dev/null; then
    curl -fsSL "$URL" -o "$TEMP_FILE"
elif command -v wget &> /dev/null; then
    wget -q "$URL" -O "$TEMP_FILE"
else
    echo "Erro: curl ou wget é necessário para fazer o download" >&2
    exit 1
fi

# Gera o diff se o arquivo existir
if [ -f "$OUTPUT_FILE" ]; then
    echo "Gerando diff em: $DIFF_FILE"
    
    # Cria o arquivo markdown com o diff
    {
        echo "# Diff do Schema oh-my-opencode"
        echo ""
        echo "**Data:** $(date -Iseconds)"
        echo ""
        echo "## Comparação"
        echo ""
        echo "\`\`\`diff"
        diff -u "$OUTPUT_FILE" "$TEMP_FILE" || true
        echo "\`\`\`"
    } > "$DIFF_FILE"
    
    echo "Diff salvo em: $DIFF_FILE"
fi

# Move o arquivo temporário para o destino final
mv "$TEMP_FILE" "$OUTPUT_FILE"

echo "Schema atualizado com sucesso: $OUTPUT_FILE"
