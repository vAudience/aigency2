#!/bin/bash

# Default values
LANGUAGE=""
declare -a INCLUDE_PATTERNS
declare -a EXCLUDE_PATTERNS
TARGET_FILE="code2markdown.md"
BASE_PATH="."

# Usage message
usage() {
    echo "Usage: $0 -lang <language> -include <pattern1> [<pattern2> ...] [-exclude <patternX> [<patternY> ...]] [-target <targetfile>] [-basepath <path>]"
    echo "Example: $0 -lang typescript -include '*.ts' 'README.md' 'SYSTEMPROMPT.md' -exclude 'node_modules/*' -target textpatch_code.md -basepath ."
    exit 1
}

# Parse options
while [[ $# -gt 0 ]]; do
    case $1 in
        -lang)
            LANGUAGE="$2"
            shift 2
            ;;
        -include)
            shift
            while [[ $# -gt 0 && ! $1 =~ ^- ]]; do
                INCLUDE_PATTERNS+=("$1")
                shift
            done
            ;;
        -exclude)
            shift
            while [[ $# -gt 0 && ! $1 =~ ^- ]]; do
                EXCLUDE_PATTERNS+=("$1")
                shift
            done
            ;;
        -target)
            TARGET_FILE="$2"
            shift 2
            ;;
        -basepath)
            BASE_PATH="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Check if language and include patterns are provided
if [[ -z "$LANGUAGE" || ${#INCLUDE_PATTERNS[@]} -eq 0 ]]; then
    echo "Error: Language (-lang) and at least one include pattern (-include) must be provided."
    usage
fi

# Clear the target file content or create it if it doesn't exist
> "$TARGET_FILE"

# Add H1 header to the top of the file with the name of the file
echo "# $(basename "$TARGET_FILE")" >> "$TARGET_FILE"
echo "" >> "$TARGET_FILE" # Add a newline for better formatting

# Helper function to check if a file matches any exclude patterns
matches_exclude_pattern() {
    local file=$1
    for pattern in "${EXCLUDE_PATTERNS[@]}"; do
        if [[ $file == $pattern ]]; then
            return 0 # File matches exclude pattern
        fi
    done
    return 1 # File does not match any exclude pattern
}

# Helper function to check if a file matches any include patterns
matches_include_pattern() {
    local file=$1
    for pattern in "${INCLUDE_PATTERNS[@]}"; do
        if [[ $file == $pattern ]]; then
            return 0 # File matches include pattern
        fi
    done
    return 1 # File does not match any include pattern
}

# Helper function to append file content to the target file
append_file_to_markdown() {
    local file=$1
    if matches_exclude_pattern "$file"; then
        return # Skip files that match exclude patterns
    fi
    if ! matches_include_pattern "$file"; then
        return # Skip files that don't match include patterns
    fi
    echo "## $file" >> "$TARGET_FILE"
    echo "" >> "$TARGET_FILE" # Add a newline for better formatting
    echo '```'"$LANGUAGE" >> "$TARGET_FILE"
    cat "$file" >> "$TARGET_FILE"
    echo '```' >> "$TARGET_FILE"
    echo "" >> "$TARGET_FILE" # Add another newline for spacing
}

# Process files recursively
find "$BASE_PATH" -type f | while read -r file; do
    relative_path=${file#"$BASE_PATH/"}
    append_file_to_markdown "$relative_path"
done

echo "Markdown file created: $TARGET_FILE"