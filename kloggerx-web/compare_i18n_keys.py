#!/usr/bin/env python3
import json

# 读取两个语言文件
with open('src/locales/zh-CN.json', 'r', encoding='utf-8') as f:
    zh_cn = json.load(f)

with open('src/locales/en-US.json', 'r', encoding='utf-8') as f:
    en_us = json.load(f)

def get_all_keys(obj, prefix=''):
    """递归获取所有 key 路径"""
    keys = set()
    if isinstance(obj, dict):
        for k, v in obj.items():
            full_key = f"{prefix}.{k}" if prefix else k
            keys.add(full_key)
            if isinstance(v, dict):
                keys.update(get_all_keys(v, full_key))
    return keys

# 获取所有 key
zh_keys = get_all_keys(zh_cn)
en_keys = get_all_keys(en_us)

# 找差异
only_in_zh = zh_keys - en_keys
only_in_en = en_keys - zh_keys

print("语言文件 Key 对比结果")
print("=" * 80)

if only_in_zh:
    print(f"\n只在 zh-CN.json 中存在的 key ({len(only_in_zh)} 个):")
    for key in sorted(only_in_zh):
        print(f"  - {key}")

if only_in_en:
    print(f"\n只在 en-US.json 中存在的 key ({len(only_in_en)} 个):")
    for key in sorted(only_in_en):
        print(f"  - {key}")

if not only_in_zh and not only_in_en:
    print("\n✓ 两个语言文件的 key 结构完全一致！")
else:
    print(f"\n发现不一致！zh-CN 中独有: {len(only_in_zh)} 个，en-US 中独有: {len(only_in_en)} 个")

