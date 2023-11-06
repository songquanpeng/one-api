# Check if local folder contains any Chinese characters
# Check files with extension *.go, *.js
# Except build and node_modules folders

import os
import re
import sys
import argparse

# Regular expression for matching Chinese characters
chinese_char_regex = re.compile(r'[\u4e00-\u9fff]+')

# Regular expressions for matching comments in Go and JavaScript
line_comment_regex = re.compile(r'//.*')
block_comment_regex = re.compile(r'/\*.*?\*/', re.DOTALL)

def remove_comments(content):
    content = line_comment_regex.sub('', content)
    content = block_comment_regex.sub('', content)
    return content

def is_chinese_present_in_file(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            content = file.read()
            content = remove_comments(content)
            return chinese_char_regex.search(content) is not None
    except Exception as e:
        print(f"Error reading file {file_path}: {e}")
        return False  # Continue to check other files even if one file cannot be read

def scan_directory_for_chinese_characters(directory, extensions, excludes):
    found_chinese = False  # Track whether Chinese characters have been found
    for root, dirs, files in os.walk(directory, topdown=True):
        dirs[:] = [d for d in dirs if d not in excludes]  # Exclude specific directories

        for file in files:
            if any(file.endswith(ext) for ext in extensions):
                file_path = os.path.join(root, file)
                if is_chinese_present_in_file(file_path):
                    print(f"Chinese characters found in: {file_path}")
                    found_chinese = True
    return found_chinese

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Check for Chinese characters in local files.')
    parser.add_argument('directory', help='Directory to check')
    parser.add_argument('--extensions', nargs='+', default=['.go', '.js'], help='File extensions to check')
    parser.add_argument('--excludes', nargs='+', default=['build', 'node_modules'], help='Directories to exclude')
    
    args = parser.parse_args()
    
    if scan_directory_for_chinese_characters(args.directory, args.extensions, args.excludes):
        sys.exit(1)  # Exit with error code 1 if Chinese characters are found
    else:
        sys.exit(0)  # Exit normally (with code 0) if no Chinese characters are found
        