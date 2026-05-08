#!/usr/bin/env python3
import os
import re

# 匹配中文字符的正则表达式
chinese_pattern = re.compile(r'[\u4e00-\u9fff]+')

# 需要检查的文件扩展名
file_extensions = ('.vue', '.ts', '.tsx')

# 需要排除的目录
exclude_dirs = {'node_modules', 'dist', '.vite', '__tests__', '.nuxt', '.output'}

# 结果存储
all_results = {}

def should_skip_line(line):
    """判断是否应该跳过整行"""
    line_stripped = line.strip()
    # 跳过纯注释行
    if line_stripped.startswith('//') or line_stripped.startswith('/*') or line_stripped.startswith('*'):
        return True
    # 跳过console输出
    if 'console.' in line and any(x in line for x in ['.log', '.warn', '.error', '.info']):
        return True
    return False

def extract_context(line):
    """提取有意义的代码上下文"""
    # 移除注释
    if '//' in line:
        line = line[:line.index('//')]
    return line.strip()

def scan_file(filepath):
    """扫描单个文件中的硬编码中文"""
    file_results = []
    
    try:
        with open(filepath, 'r', encoding='utf-8', errors='ignore') as f:
            lines = f.readlines()
        
        for line_num, line in enumerate(lines, 1):
            if should_skip_line(line):
                continue
            
            # 查找所有中文
            for match in chinese_pattern.finditer(line):
                chinese_text = match.group()
                context = extract_context(line)
                
                file_results.append({
                    'line_num': line_num,
                    'chinese': chinese_text,
                    'context': context[:100]  # 限制长度
                })
    except Exception as e:
        pass
    
    return file_results

def main():
    src_path = 'src'
    
    # 遍历所有文件
    for root, dirs, files in os.walk(src_path):
        # 移除排除的目录
        dirs[:] = [d for d in dirs if d not in exclude_dirs]
        
        for file in files:
            if file.endswith(file_extensions):
                filepath = os.path.join(root, file)
                file_results = scan_file(filepath)
                
                if file_results:
                    all_results[filepath] = file_results
    
    # 输出结果为JSON格式，便于后续处理
    import json
    print(json.dumps(all_results, ensure_ascii=False, indent=2))

if __name__ == '__main__':
    main()
