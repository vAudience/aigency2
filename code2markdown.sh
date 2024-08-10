#!/bin/bash

# Initial setup
LANGUAGE="go" # default language
INCLUDE_PATTERN='*' # default include pattern
declare -a IGNORE_PATTERNS # declare an array for ignore patterns
RECURSIVE=false

# Usage message
usage() {
    echo "Usage: $0 [-l language] [-R include-pattern] [-i ignore-pattern]... [source-files-and-directories] <target-file>"
    echo "Example: $0 -l go -R '*.go' -i 'test' -i 'aigentchatcli' . aigentchat_code.md"
    echo "Multiple -i options can be used to specify multiple ignore patterns."
    exit 1
}

# Parse options
while getopts ":l:R:i:" opt; do
    case ${opt} in
        l ) # Programming language
            LANGUAGE=$OPTARG
            ;;
        R ) # Recursive include pattern
            INCLUDE_PATTERN=$OPTARG
            RECURSIVE=true
            ;;
        i ) # Ignore pattern (accumulate multiple patterns into an array)
            IGNORE_PATTERNS+=("$OPTARG")
            ;;
        \? ) usage
            ;;
    esac
done
shift $((OPTIND -1))

# Check if at least one argument is provided for source
if [ "$#" -lt 2 ]; then
    usage
fi

# The target file is the last argument
TARGET_FILE="${@: -1}"

# Clear the target file content or create it if doesn't exist
> "$TARGET_FILE"

# Add H1 header to the top of the file with the name of the file
echo "# $(basename "$TARGET_FILE")" >> "$TARGET_FILE"
echo "" >> "$TARGET_FILE" # Add a newline for better formatting

# Helper function to check if a file matches any ignore patterns
matches_ignore_pattern() {
    local file=$1
    for pattern in "${IGNORE_PATTERNS[@]}"; do
        if [[ $file =~ $pattern ]]; then
            return 0 # File matches ignore pattern
        fi
    done
    return 1 # File does not match any ignore pattern
}

# Helper function to append file content to the target file
append_file_to_markdown() {
    local file=$1
    if matches_ignore_pattern "$file"; then
        return # Skip files that match ignore patterns
    fi
    echo "## $file" >> "$TARGET_FILE"
    echo "" >> "$TARGET_FILE" # Add a newline for better formatting
    echo '```'"$LANGUAGE" >> "$TARGET_FILE"
    cat "$file" >> "$TARGET_FILE"
    echo '```' >> "$TARGET_FILE"
    echo "" >> "$TARGET_FILE" # Add another newline for spacing
}

# Process files and directories
for item in "${@:1:$#-1}"; do
    if [ -f "$item" ]; then
        append_file_to_markdown "$item"
    elif [ -d "$item" ] && [ "$RECURSIVE" = true ]; then
        while IFS= read -r -d '' file; do
            append_file_to_markdown "$file"
        done < <(find "$item" -type f -name "$INCLUDE_PATTERN" -print0)
    fi
done
