#!/usr/bin/env python3
import os
import re
import sys

def update_imports_in_file(filepath):
    """Update imports from @/shared/constants/* to @/shared/constants"""
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Pattern to match static imports from @/shared/constants with any subpath
    # This handles both single and double quotes, with or without .js extension
    static_pattern = r"from ['\"]@/shared/constants/([^'\"]+)['\"]"
    
    # Pattern to match dynamic imports like await import("@/shared/constants/pricing.js")
    dynamic_pattern = r"import\(['\"]@/shared/constants/([^'\"]+)['\"]\)"
    
    updated = False
    
    def replace_static_import(match):
        nonlocal updated
        updated = True
        return 'from "@/shared/constants"'
    
    new_content = re.sub(static_pattern, replace_static_import, content)
    
    def replace_dynamic_import(match):
        nonlocal updated
        updated = True
        return 'import("@/shared/constants")'
    
    new_content = re.sub(dynamic_pattern, replace_dynamic_import, new_content)
    
    if updated:
        with open(filepath, 'w') as f:
            f.write(new_content)
    
    return updated

def main():
    frontend_dir = "/home/ubuntu/ai_proxy/frontend"
    src_dir = os.path.join(frontend_dir, "src")
    
    if not os.path.exists(src_dir):
        print(f"Error: src directory not found at {src_dir}")
        sys.exit(1)
    
    updated_files = []
    
    # Walk through all files in src directory
    for root, dirs, files in os.walk(src_dir):
        for file in files:
            if file.endswith(('.js', '.jsx', '.ts', '.tsx')):
                filepath = os.path.join(root, file)
                if update_imports_in_file(filepath):
                    updated_files.append(filepath)
    
    print(f"Updated {len(updated_files)} files:")
    for f in updated_files:
        print(f"  - {os.path.relpath(f, frontend_dir)}")
    
    # Verify no deep imports remain
    print("\nVerifying no deep imports remain...")
    os.chdir(frontend_dir)
    result = os.system('grep -rn "from [\\"\\\']@/shared/constants/" src/ 2>/dev/null | head -10')
    
    if result == 0:
        print("WARNING: Some deep imports may still exist!")
    else:
        print("SUCCESS: No deep imports found.")

if __name__ == "__main__":
    main()