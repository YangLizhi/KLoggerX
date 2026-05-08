#!/usr/bin/env python3
import os
import re
import json

# 匹配中文字符的正则表达式
chinese_pattern = re.compile(r'[\u4e00-\u9fff]+')

# 需要检查的文件扩展名
file_extensions = ('.vue', '.ts', '.tsx')

# 需要排除的目录
exclude_dirs = {'node_modules', 'dist', '.vite', '__tests__', '.nuxt', '.output'}

# 结果存储
results = {}

def is_in_excluded_dir(path):
    """检查路径是否在排除目录中"""
    for excluded in exclude_dirs:
        if f'/{excluded}/' in path or path.startswith(excluded):
            return True
    return False

def is_code_comment(line, position):
    """检查是否在注释中"""
    # 检查 // 注释
    comment_start = line.find('//')
    if comment_start != -1 and position > comment_start:
        return True
    # 检查 /* */ 注释 - 更复杂，这里简化处理
    return False

def should_skip_match(line, match):
    """判断是否应该跳过这个匹配"""
    match_pos = match.start()
    
    # 跳过注释中的中文
    if is_code_comment(line, match_pos):
        return True
    
    # 跳过 console.log 中的中文
    if 'console.log' in line or 'console.warn' in line or 'console.error' in line:
        return True
    
    return False

def scan_file(filepath):
    """扫描单个文件中的硬编码中文"""
    file_results = []
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            lines = f.readlines()
        
        for line_num, line in enumerate(lines, 1):
            matches = list(chinese_pattern.finditer(line))
            for match in matches:
                if not should_skip_match(line, match):
                    # 获取上下文（前后几个字符）
                    start = max(0, match.start() - 20)
                    end = min(len(line), match.end() + 20)
                    context = line[start:end].strip()
                    
                    file_results.append({
                        'line': line_num,
                        'chinese': match.group(),
                        'context': context,
                        'full_line': line.strip()
                    })
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
    
    return file_results

def main():
    src_path = 'src'
    
    # 遍历所有文件
    for root, dirs, files in os.walk(src_path):
        # 移除排除的目录
        dirs[:] = [d for d in dirs if d not in exclude_dirs]
        
        if is_in_excluded_dir(root):
            continue
        
        for file in files:
            if file.endswith(file_extensions):
                filepath = os.path.join(root, file)
                file_results = scan_file(filepath)
                
                if file_results:
                    results[filepath] = file_results
    
    # 输出结果
    if results:
        print(f"发现 {len(results)} 个文件包含硬编码中文\n")
        print("=" * 100)
        
        for filepath, items in sorted(results.items()):
            print(f"\n文件: {filepath}")
            print(f"发现 {len(items)} 处硬编码中文:")
            print("-" * 100)
            
            for item in items:
                print(f"  行号: {item['line']}")
                print(f"  中文: {item['chinese']}")
                print(f"  上下文: {item['context']}")
                print(f"  完整行: {item['full_line'][:80]}")
                print()
    else:
        print("未发现硬编码中文！")

if __name__ == '__main__':
    main()
