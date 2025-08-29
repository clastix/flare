#!/bin/bash

# Script to create FLARE Documentation DOCX with internal cross-references
# Run this from the flare-internal root directory

# Check if pandoc is installed
if ! command -v pandoc &> /dev/null; then
    echo "❌ Error: pandoc is not installed or not in PATH"
    echo ""
    echo "Please install pandoc to generate DOCX documentation:"
    echo "  • Ubuntu/Debian: sudo apt install pandoc"
    echo "  • macOS: brew install pandoc"
    echo "  • Windows: Download from https://pandoc.org/installing.html"
    echo ""
    exit 1
fi

echo "Creating FLARE Documentation..."

# Create structured markdown file with proper hierarchy
cat > FLARE_Structured_Documentation.md << 'DOC_EOF'
# Part I: Getting Started

DOC_EOF

# Function to process markdown files and fix cross-references
process_markdown() {
    local input_file="$1"
    
    # Convert ALL markdown file links to internal references
    # Handle both docs/ prefixed and direct file references
    # First handle links with section anchors, then whole file links
    sed -E \
        -e 's|\[([^]]+)\]\(FLARE_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLUIDOS_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(NVIDIA_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(AMD_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(docs/FLARE_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(docs/FLUIDOS_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(docs/NVIDIA_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(docs/AMD_[^)#]*#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(docs/FLARE_placeholder\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_placeholder\.md#([^)]+)\)|[\1](#\2)|g' \
        "$input_file" | \
    sed -E \
        -e 's|\[([^]]+)\]\(FLARE_API_Reference\.md\)|[\1](#chapter-6-flare-api-reference)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Architecture\.md\)|[\1](#chapter-5-flare-architecture)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Admin_Guide\.md\)|[\1](#chapter-11-admin-guide)|g' \
        -e 's|\[([^]]+)\]\(FLARE_GPU_Annotations_Reference\.md\)|[\1](#chapter-7-gpu-annotations-reference)|g' \
        -e 's|\[([^]]+)\]\(FLARE_GPU_Pooling_Guide\.md\)|[\1](#chapter-4-flare-gpu-pooling-guide)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Project_Overview\.md\)|[\1](#chapter-1-project-overview)|g' \
        -e 's|\[([^]]+)\]\(FLARE_QuickStart_Guide\.md\)|[\1](#chapter-2-quickstart-guide)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Sample_Use_Cases\.md\)|[\1](#chapter-12-sample-use-cases)|g' \
        -e 's|\[([^]]+)\]\(FLUIDOS_Basic_Workflow\.md\)|[\1](#chapter-3-fluidos-basic-workflow)|g' \
        -e 's|\[([^]]+)\]\(NVIDIA_GPU_Labels_Mapping\.md\)|[\1](#chapter-8-nvidia-gpu-labels-mapping)|g' \
        -e 's|\[([^]]+)\]\(AMD_GPU_Labels_Mapping\.md\)|[\1](#chapter-9-amd-gpu-labels-mapping)|g' \
        -e 's|\[([^]]+)\]\(FLARE_placeholder\.md\)|[\1](#chapter-10-efficient-gpu-management)|g' \
        -e 's|\[([^]]+)\]\(docs/FLARE_placeholder\.md\)|[\1](#chapter-10-efficient-gpu-management)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Final_Project_Review\.md\)|[\1](#chapter-13-final-project-review)|g' \
        -e 's|\[([^]]+)\]\(docs/FLARE_Final_Project_Review\.md\)|[\1](#chapter-13-final-project-review)|g' \
        -e 's|\[([^]]+)\]\(FLARE_API_Reference\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Architecture\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Admin_Guide\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_GPU_Annotations_Reference\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_GPU_Pooling_Guide\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Project_Overview\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_QuickStart_Guide\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLARE_Sample_Use_Cases\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(FLUIDOS_Basic_Workflow\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(NVIDIA_GPU_Labels_Mapping\.md#([^)]+)\)|[\1](#\2)|g' \
        -e 's|\[([^]]+)\]\(AMD_GPU_Labels_Mapping\.md#([^)]+)\)|[\1](#\2)|g'
}

# Add Chapter 1: Project Overview
echo "## Chapter 1: Project Overview" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_Project_Overview.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

# Add Chapter 2: QuickStart Guide
echo "## Chapter 2: QuickStart Guide" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_QuickStart_Guide.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

# Part II: Core Concepts
echo "# Part II: Core Concepts" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md

echo "## Chapter 3: FLUIDOS Basic Workflow" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLUIDOS_Basic_Workflow.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

echo "## Chapter 4: FLARE GPU Pooling Guide" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_GPU_Pooling_Guide.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

# Part III: Architecture & API
echo "# Part III: Architecture & API" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md

echo "## Chapter 5: FLARE Architecture" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_Architecture.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

echo "## Chapter 6: FLARE API Reference" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_API_Reference.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

# Part IV: GPU Resource Management
echo "# Part IV: GPU Resource Management" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md

echo "## Chapter 7: GPU Annotations Reference" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_GPU_Annotations_Reference.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

echo "## Chapter 8: NVIDIA GPU Labels Mapping" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/NVIDIA_GPU_Labels_Mapping.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

echo "## Chapter 9: AMD GPU Labels Mapping" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/AMD_GPU_Labels_Mapping.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

echo "## Chapter 10: Efficient GPU Management" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_placeholder.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

# Part V: Operations & Administration
echo "# Part V: Operations & Administration" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md

echo "## Chapter 11: Admin Guide" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_Admin_Guide.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

# Part VI: Use Cases & Examples
echo "# Part VI: Use Cases & Examples" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md

echo "## Chapter 12: Sample Use Cases" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_Sample_Use_Cases.md" >> FLARE_Structured_Documentation.md
cat >> FLARE_Structured_Documentation.md << 'PAGEBREAK_EOF'


```{=openxml}
<w:p><w:r><w:br w:type="page"/></w:r></w:p>
```

PAGEBREAK_EOF

# Part VII: Project Documentation
echo "# Part VII: Project Documentation" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md

echo "## Chapter 13: Final Project Review" >> FLARE_Structured_Documentation.md
echo "" >> FLARE_Structured_Documentation.md
process_markdown "docs/FLARE_Final_Project_Review.md" >> FLARE_Structured_Documentation.md

# Convert to DOCX with pandoc - balance between functionality and clean appearance
echo "Converting to DOCX format..."
pandoc FLARE_Structured_Documentation.md \
  -o FLARE_Documentation.docx \
  -f markdown-yaml_metadata_block \
  --toc \
  --toc-depth=2 \
  --highlight-style=tango \
  --standalone

# Clean up temporary files
rm -f FLARE_Structured_Documentation.md

echo "✅ FLARE_Documentation.docx created successfully!"
echo ""
echo "Documentation features:"
echo "  • Professional hierarchical structure with Parts and Chapters"
echo "  • Working table of contents with navigation links"
echo "  • Internal cross-references throughout all documents"
echo "  • Page breaks between major sections"
echo "  • Clean formatting optimized for Word/Office viewers"
echo ""
echo "The complete FLARE documentation is now available in DOCX format."