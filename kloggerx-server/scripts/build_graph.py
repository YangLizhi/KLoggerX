#!/usr/bin/env python3
"""
build_graph.py - 知识图谱构建与社区检测脚本

用法:
    python3 build_graph.py --input /path/to/entities.json --output /path/to/graph.json

功能:
    1. 读取 entities.json (实体+关系)
    2. 构建 NetworkX 有向图，同名实体自动合并
    3. Leiden 社区检测 (fallback 到连通组件分组)
    4. 输出 graph.json (含节点、边、社区信息、元数据)
"""

import argparse
import json
import sys
from collections import defaultdict

try:
    import networkx as nx
except ImportError:
    print("错误: 缺少 networkx 依赖，请执行 pip install networkx", file=sys.stderr)
    sys.exit(1)

try:
    import igraph as ig
    import leidenalg
    LEIDEN_AVAILABLE = True
except ImportError:
    print("警告: leidenalg 或 python-igraph 不可用，将使用连通组件分组作为替代", file=sys.stderr)
    LEIDEN_AVAILABLE = False

COMMUNITY_COLORS = [
    "#4285f4", "#ea4335", "#fbbc04", "#34a853", "#ff6d01",
    "#46bdc6", "#7baaf7", "#f07b72", "#fdd663", "#57bb8a",
    "#ff8a50", "#78d9e0", "#a0c4ff", "#f4a261", "#2a9d8f",
    "#e76f51", "#264653", "#e9c46a", "#f4a261", "#606c38"
]


def load_input(input_path: str) -> dict:
    """读取并验证 entities.json"""
    try:
        with open(input_path, "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"错误: 输入文件不存在: {input_path}", file=sys.stderr)
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"错误: JSON 解析失败: {e}", file=sys.stderr)
        sys.exit(1)

    if not isinstance(data, dict):
        print("错误: 输入文件顶层必须是 JSON 对象", file=sys.stderr)
        sys.exit(1)

    if "entities" not in data:
        print("错误: 输入文件缺少 'entities' 字段", file=sys.stderr)
        sys.exit(1)

    if "relationships" not in data:
        data["relationships"] = []

    return data


def build_graph(data: dict) -> nx.DiGraph:
    """构建 NetworkX 有向图，同名实体自动合并"""
    G = nx.DiGraph()

    # 用于按 label 合并同名实体
    label_to_id = {}  # label -> 首次出现的 id
    id_remap = {}     # 原始 id -> 合并后的 id

    for entity in data.get("entities", []):
        eid = entity.get("id", "")
        label = entity.get("label", "").strip()
        etype = entity.get("type", "")
        description = entity.get("description", "")
        source_doc_id = entity.get("source_doc_id", "")

        if not eid or not label:
            continue

        if label in label_to_id:
            # 合并：将此实体映射到已有节点
            canonical_id = label_to_id[label]
            id_remap[eid] = canonical_id
            # 合并 source_docs
            if source_doc_id:
                G.nodes[canonical_id]["source_docs"].add(source_doc_id)
            # 如果描述更长，更新描述
            if len(description) > len(G.nodes[canonical_id].get("description", "")):
                G.nodes[canonical_id]["description"] = description
        else:
            # 新节点
            label_to_id[label] = eid
            id_remap[eid] = eid
            source_docs = set()
            if source_doc_id:
                source_docs.add(source_doc_id)
            G.add_node(eid, label=label, type=etype,
                       description=description, source_docs=source_docs)

    # 添加边
    for rel in data.get("relationships", []):
        source = rel.get("source", "")
        target = rel.get("target", "")
        rel_type = rel.get("type", "")
        description = rel.get("description", "")
        confidence = rel.get("confidence", 0.0)

        if not source or not target:
            continue

        # 重映射到合并后的 id
        source = id_remap.get(source, source)
        target = id_remap.get(target, target)

        # 只添加两端都存在的边
        if source not in G.nodes or target not in G.nodes:
            continue

        # 跳过自环
        if source == target:
            continue

        # 确定 confidence_tag
        if confidence >= 0.8:
            confidence_tag = "EXTRACTED"
        elif confidence >= 0.5:
            confidence_tag = "INFERRED"
        else:
            confidence_tag = "LOW_CONFIDENCE"

        G.add_edge(source, target,
                   relationship=rel_type,
                   label=description,
                   confidence=confidence,
                   confidence_tag=confidence_tag)

    return G


def detect_communities(G: nx.DiGraph) -> dict:
    """
    社区检测：优先使用 Leiden 算法，fallback 到连通组件分组。
    返回 {node_id: community_id}
    """
    if len(G.nodes) < 3 or not LEIDEN_AVAILABLE:
        # Fallback: 使用弱连通组件作为社区
        undirected = G.to_undirected()
        components = list(nx.connected_components(undirected))
        node_community = {}
        for cid, comp in enumerate(components):
            for node in comp:
                node_community[node] = cid
        return node_community

    # 使用 Leiden 算法
    # 将 NetworkX 图转为 igraph
    node_list = list(G.nodes())
    node_index = {n: i for i, n in enumerate(node_list)}

    # 创建 igraph 图（无向，Leiden 需要）
    ig_graph = ig.Graph(n=len(node_list), directed=False)

    edges = set()
    for u, v in G.edges():
        ui, vi = node_index[u], node_index[v]
        if ui != vi:
            edge = (min(ui, vi), max(ui, vi))
            edges.add(edge)

    ig_graph.add_edges(list(edges))

    # 运行 Leiden
    try:
        partition = leidenalg.find_partition(ig_graph, leidenalg.ModularityVertexPartition)
        node_community = {}
        for cid, members in enumerate(partition):
            for idx in members:
                node_community[node_list[idx]] = cid
        return node_community
    except Exception as e:
        print(f"警告: Leiden 算法执行失败 ({e})，使用连通组件分组", file=sys.stderr)
        undirected = G.to_undirected()
        components = list(nx.connected_components(undirected))
        node_community = {}
        for cid, comp in enumerate(components):
            for node in comp:
                node_community[node] = cid
        return node_community


def analyze_graph(G: nx.DiGraph, node_community: dict) -> dict:
    """分析图：god nodes、社区统计、社区标签"""
    # 计算 degree
    degree_map = dict(G.degree())

    # God nodes: degree 排名前 5
    sorted_nodes = sorted(degree_map.items(), key=lambda x: x[1], reverse=True)
    god_nodes = [n for n, _ in sorted_nodes[:5]]

    # 社区统计
    community_nodes = defaultdict(list)
    for node, cid in node_community.items():
        community_nodes[cid].append(node)

    communities = []
    for cid in sorted(community_nodes.keys()):
        nodes_in_community = community_nodes[cid]
        node_count = len(nodes_in_community)

        # 社区标签：取该社区中 degree 最高的节点 label
        best_node = max(nodes_in_community, key=lambda n: degree_map.get(n, 0))
        community_label = G.nodes[best_node].get("label", f"Community {cid}")

        color = COMMUNITY_COLORS[cid % len(COMMUNITY_COLORS)]

        communities.append({
            "id": cid,
            "label": community_label,
            "color": color,
            "node_count": node_count
        })

    return {
        "god_nodes": god_nodes,
        "communities": communities,
        "degree_map": degree_map
    }


def build_output(G: nx.DiGraph, node_community: dict, analysis: dict) -> dict:
    """构建最终输出 JSON"""
    degree_map = analysis["degree_map"]

    nodes = []
    for nid, attrs in G.nodes(data=True):
        nodes.append({
            "id": nid,
            "label": attrs.get("label", ""),
            "type": attrs.get("type", ""),
            "description": attrs.get("description", ""),
            "source_docs": sorted(list(attrs.get("source_docs", set()))),
            "community": node_community.get(nid, 0),
            "degree": degree_map.get(nid, 0)
        })

    edges = []
    for u, v, attrs in G.edges(data=True):
        edges.append({
            "source": u,
            "target": v,
            "relationship": attrs.get("relationship", ""),
            "label": attrs.get("label", ""),
            "confidence": attrs.get("confidence", 0.0),
            "confidence_tag": attrs.get("confidence_tag", "LOW_CONFIDENCE")
        })

    output = {
        "nodes": nodes,
        "edges": edges,
        "communities": analysis["communities"],
        "metadata": {
            "node_count": len(nodes),
            "edge_count": len(edges),
            "community_count": len(analysis["communities"]),
            "god_nodes": analysis["god_nodes"]
        }
    }

    return output


def main():
    parser = argparse.ArgumentParser(description="知识图谱构建与社区检测")
    parser.add_argument("--input", required=True, help="输入文件路径 (entities.json)")
    parser.add_argument("--output", required=True, help="输出文件路径 (graph.json)")
    args = parser.parse_args()

    # 1. 读取输入
    data = load_input(args.input)

    # 2. 构建图
    G = build_graph(data)

    if len(G.nodes) == 0:
        print("警告: 图中没有有效节点，生成空结果", file=sys.stderr)
        output = {
            "nodes": [],
            "edges": [],
            "communities": [],
            "metadata": {
                "node_count": 0,
                "edge_count": 0,
                "community_count": 0,
                "god_nodes": []
            }
        }
    else:
        # 3. 社区检测
        node_community = detect_communities(G)

        # 4. 分析
        analysis = analyze_graph(G, node_community)

        # 5. 构建输出
        output = build_output(G, node_community, analysis)

    # 6. 写入输出文件
    try:
        with open(args.output, "w", encoding="utf-8") as f:
            json.dump(output, f, ensure_ascii=False, indent=2)
    except IOError as e:
        print(f"错误: 无法写入输出文件: {e}", file=sys.stderr)
        sys.exit(1)

    print(f"图构建完成: {output['metadata']['node_count']} 节点, "
          f"{output['metadata']['edge_count']} 边, "
          f"{output['metadata']['community_count']} 社区")


if __name__ == "__main__":
    main()
