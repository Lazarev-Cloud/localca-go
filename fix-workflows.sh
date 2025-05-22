#!/bin/bash

# Fix workflow YAML issues
echo "Fixing GitHub Actions workflows..."

WORKFLOW_DIR=".github/workflows"

# Add document start and fix common issues for each workflow
for workflow in $WORKFLOW_DIR/*.yml; do
    if [[ -f "$workflow" && ! "$workflow" =~ "old" ]]; then
        filename=$(basename "$workflow")
        echo "Processing $filename..."
        
        # Add document start if missing
        if ! head -1 "$workflow" | grep -q "^---"; then
            sed -i '' '1i\
---
' "$workflow"
        fi
        
        # Remove trailing spaces
        sed -i '' 's/[[:space:]]*$//' "$workflow"
        
        # Add final newline if missing
        if [ "$(tail -c1 "$workflow" | wc -l)" -eq 0 ]; then
            echo "" >> "$workflow"
        fi
        
        # Fix action versions
        sed -i '' 's|securego/gosec@master|securego/gosec@v2.21.4|g' "$workflow"
        sed -i '' 's|aquasecurity/trivy-action@master|aquasecurity/trivy-action@0.24.0|g' "$workflow"
        sed -i '' 's|SonarSource/sonarqube-scan-action@master|SonarSource/sonarqube-scan-action@v2.3.0|g' "$workflow"
        sed -i '' 's|trufflesecurity/trufflehog@main|trufflesecurity/trufflehog@v3.82.6|g' "$workflow"
        
        echo "  ✅ Fixed $filename"
    fi
done

echo
echo "All workflows processed. Running validation..."

# Quick validation
VALID_COUNT=0
TOTAL_COUNT=0

for workflow in $WORKFLOW_DIR/*.yml; do
    if [[ -f "$workflow" && ! "$workflow" =~ "old" ]]; then
        TOTAL_COUNT=$((TOTAL_COUNT + 1))
        filename=$(basename "$workflow")
        
        # Check if it starts with --- now
        if head -1 "$workflow" | grep -q "^---"; then
            VALID_COUNT=$((VALID_COUNT + 1))
            echo "✅ $filename has document start"
        else
            echo "❌ $filename missing document start"
        fi
    fi
done

echo
echo "Summary: $VALID_COUNT/$TOTAL_COUNT workflows have proper YAML headers"
echo "Run 'yamllint .github/workflows/*.yml' for detailed validation"