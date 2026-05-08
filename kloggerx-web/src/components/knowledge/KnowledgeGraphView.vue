<template>
  <div class="kg-graph-view" ref="containerRef">
    <!-- Loading -->
    <div v-if="loading" class="kg-loading">
      <el-icon class="kg-loading-icon" :size="32"><Loading /></el-icon>
      <span>{{ $t('knowledge.graph.loading') }}</span>
    </div>

    <!-- Empty State -->
    <div v-else-if="!graphData || !graphData.nodes.length" class="kg-empty">
      <el-icon :size="48" color="#c0c4cc"><Connection /></el-icon>
      <p>{{ $t('knowledge.graph.noData') }}</p>
    </div>

    <!-- Graph Content -->
    <template v-else>
      <!-- Toolbar -->
      <div class="kg-toolbar">
        <el-input
          v-model="searchQuery"
          :placeholder="$t('knowledge.graph.searchNode')"
          prefix-icon="Search"
          clearable
          size="small"
          class="kg-search"
          @input="handleSearch"
          @clear="clearSearch"
        />
        <div class="kg-stats">
          <el-tag size="small" type="info">{{ graphData.metadata.node_count }} {{ $t('knowledge.graph.nodes') }}</el-tag>
          <el-tag size="small" type="info">{{ graphData.metadata.edge_count }} {{ $t('knowledge.graph.edges') }}</el-tag>
          <el-tag size="small" type="info">{{ graphData.metadata.community_count }} {{ $t('knowledge.graph.communities') }}</el-tag>
        </div>
        <el-popover placement="bottom-end" :width="280" trigger="click">
          <template #reference>
            <el-button size="small" text><el-icon><Filter /></el-icon>{{ $t('knowledge.graph.filter') }}</el-button>
          </template>
          <div class="kg-filter-panel">
            <div class="kg-filter-section">
              <div class="kg-filter-title">{{ $t('knowledge.graph.community') }}</div>
              <el-checkbox-group v-model="visibleCommunities" @change="updateVisibility">
                <el-checkbox
                  v-for="c in graphData.communities"
                  :key="c.id"
                  :value="c.id"
                >
                  <span class="kg-filter-dot" :style="{ background: c.color }"></span>
                  {{ c.label }} ({{ c.node_count }})
                </el-checkbox>
              </el-checkbox-group>
            </div>
            <div class="kg-filter-section">
              <div class="kg-filter-title">{{ $t('knowledge.graph.nodeType') }}</div>
              <el-checkbox-group v-model="visibleTypes" @change="updateVisibility">
                <el-checkbox v-for="nt in nodeTypes" :key="nt" :value="nt">{{ typeLabel(nt) }}</el-checkbox>
              </el-checkbox-group>
            </div>
          </div>
        </el-popover>
        <el-button size="small" text @click="resetZoom"><el-icon><FullScreen /></el-icon></el-button>
      </div>

      <!-- Main Area -->
      <div class="kg-main">
        <div class="kg-svg-wrap" ref="svgWrapRef">
          <svg ref="svgRef" class="kg-svg"></svg>
          <!-- Tooltip -->
          <div
            v-show="tooltip.visible"
            class="kg-tooltip"
            :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
          >
            <div class="kg-tooltip-label">{{ tooltip.label }}</div>
            <div class="kg-tooltip-type">{{ typeLabel(tooltip.type) }}</div>
            <div class="kg-tooltip-degree">{{ $t('knowledge.graph.degree') }}: {{ tooltip.degree }}</div>
            <div class="kg-tooltip-desc" v-if="tooltip.desc">{{ tooltip.desc }}</div>
          </div>

          <!-- Zoom Controls -->
          <div class="graph-controls">
            <button @click="zoomIn" :title="$t('knowledge.graph.zoomIn')">+</button>
            <button @click="zoomOut" :title="$t('knowledge.graph.zoomOut')">-</button>
            <button @click="resetZoom" :title="$t('knowledge.graph.resetZoom')">{{ $t('knowledge.graph.resetZoom') }}</button>
            <button @click="toggleFullscreen" :title="$t('knowledge.graph.fullscreen')">{{ $t('knowledge.graph.fullscreen') }}</button>
          </div>

          <!-- Legend overlay -->
          <div class="kg-legend-overlay" v-if="graphData.communities.length">
            <div class="kg-legend-title">{{ $t('knowledge.graph.legend') }}</div>
            <div class="kg-legend-items">
              <span v-for="c in graphData.communities" :key="c.id" class="kg-legend-item">
                <span class="kg-legend-dot" :style="{ background: c.color }"></span>
                {{ c.label }}
              </span>
            </div>
            <div class="kg-legend-size">
              <span class="kg-legend-size-small"></span>
              <span class="kg-legend-size-label">{{ $t('knowledge.graph.fewConnections') }}</span>
              <span class="kg-legend-size-large"></span>
              <span class="kg-legend-size-label">{{ $t('knowledge.graph.manyConnections') }}</span>
            </div>
            <template v-if="graphData.metadata.god_nodes?.length">
              <div class="kg-legend-core">
                <el-icon :size="12"><Star /></el-icon>
                {{ $t('knowledge.graph.coreNodes') }}: {{ graphData.metadata.god_nodes.join(', ') }}
              </div>
            </template>
          </div>
        </div>

        <!-- Detail Panel -->
        <transition name="kg-slide">
          <div v-if="selectedNode" class="kg-detail">
            <div class="kg-detail-header">
              <span class="kg-detail-title">{{ selectedNode.label }}</span>
              <el-icon class="kg-detail-close" @click="deselectNode"><Close /></el-icon>
            </div>
            <div class="kg-detail-body">
              <div class="kg-detail-row">
                <span class="kg-detail-key">{{ $t('knowledge.graph.type') }}</span>
                <el-tag size="small">{{ typeLabel(selectedNode.type) }}</el-tag>
              </div>
              <div class="kg-detail-row" v-if="selectedNode.description">
                <span class="kg-detail-key">{{ $t('knowledge.graph.descLabel') }}</span>
                <span class="kg-detail-val">{{ selectedNode.description }}</span>
              </div>
              <div class="kg-detail-row">
                <span class="kg-detail-key">{{ $t('knowledge.graph.degree') }}</span>
                <span class="kg-detail-val">{{ selectedNode.degree }}</span>
              </div>
              <div class="kg-detail-row">
                <span class="kg-detail-key">{{ $t('knowledge.graph.community') }}</span>
                <span class="kg-detail-val">
                  <span class="kg-filter-dot" :style="{ background: getCommunityColor(selectedNode.community) }"></span>
                  {{ getCommunityLabel(selectedNode.community) }}
                </span>
              </div>
              <div class="kg-detail-section" v-if="selectedNeighbors.length">
                <div class="kg-detail-key">{{ $t('knowledge.graph.relatedEntities') }} ({{ selectedNeighbors.length }})</div>
                <div
                  v-for="nb in selectedNeighbors"
                  :key="nb.id"
                  class="kg-neighbor-item"
                  @click="focusNode(nb.id)"
                >
                  <span class="kg-filter-dot" :style="{ background: getCommunityColor(nb.community) }"></span>
                  <span>{{ nb.label }}</span>
                  <el-tag size="small" type="info">{{ nb.rel }}</el-tag>
                </div>
              </div>
              <div class="kg-detail-section" v-if="selectedNode.source_docs?.length">
                <div class="kg-detail-key">{{ $t('knowledge.graph.sourceDocs') }} ({{ selectedNode.source_docs.length }})</div>
                <div v-for="doc in selectedNode.source_docs" :key="doc" class="kg-source-item">
                  <el-icon><Document /></el-icon>
                  <span>{{ doc }}</span>
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <!-- Bottom Legend (kept for compatibility) -->
      <div class="kg-legend" v-if="graphData.communities.length">
        <span v-for="c in graphData.communities" :key="c.id" class="kg-legend-item">
          <span class="kg-legend-dot" :style="{ background: c.color }"></span>
          {{ c.label }}
        </span>
        <template v-if="graphData.metadata.god_nodes?.length">
          <span class="kg-legend-sep">|</span>
          <span class="kg-legend-item kg-legend-god">
            <el-icon :size="12"><Star /></el-icon>
            {{ $t('knowledge.graph.coreNodes') }}: {{ graphData.metadata.god_nodes.join(', ') }}
          </span>
        </template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import * as d3 from 'd3'

// ── Types ──────────────────────────────────────────────
interface GraphNode {
  id: string
  label: string
  type: string
  description: string
  source_docs: string[]
  community: number
  degree: number
  x?: number
  y?: number
  fx?: number | null
  fy?: number | null
}

interface GraphEdge {
  source: string | GraphNode
  target: string | GraphNode
  relationship: string
  label: string
  confidence: number
  confidence_tag: string
}

interface Community {
  id: number
  label: string
  color: string
  node_count: number
}

interface GraphMetadata {
  node_count: number
  edge_count: number
  community_count: number
  god_nodes?: string[]
}

interface GraphData {
  nodes: GraphNode[]
  edges: GraphEdge[]
  communities: Community[]
  metadata: GraphMetadata
}

// ── Props ──────────────────────────────────────────────
const props = withDefaults(defineProps<{
  graphData: GraphData | null
  loading?: boolean
}>(), {
  loading: false
})

const { t } = useI18n()
const router = useRouter()

// ── Refs ───────────────────────────────────────────────
const containerRef = ref<HTMLDivElement>()
const svgWrapRef = ref<HTMLDivElement>()
const svgRef = ref<SVGSVGElement>()

const searchQuery = ref('')
const visibleCommunities = ref<number[]>([])
const visibleTypes = ref<string[]>([])
const nodeTypes = ['concept', 'person', 'technology', 'event', 'organization']

const selectedNode = ref<GraphNode | null>(null)
const tooltip = ref({ visible: false, x: 0, y: 0, label: '', type: '', desc: '', degree: 0 })

let simulation: d3.Simulation<GraphNode, GraphEdge> | null = null
let svgSelection: d3.Selection<SVGSVGElement, unknown, null, undefined> | null = null
let gRoot: d3.Selection<SVGGElement, unknown, null, undefined> | null = null
let zoomBehavior: d3.ZoomBehavior<SVGSVGElement, unknown> | null = null
let resizeObserver: ResizeObserver | null = null

// working copies
let workingNodes: GraphNode[] = []
let workingEdges: GraphEdge[] = []

// ── Helpers ────────────────────────────────────────────
function typeLabel(t2: string): string {
  const map: Record<string, string> = {
    concept: t('knowledge.graph.concept'), person: t('knowledge.graph.person'), technology: t('knowledge.graph.technology'), event: t('knowledge.graph.event'), organization: t('knowledge.graph.organization')
  }
  return map[t2] || t2
}

function getNodeRadius(d: GraphNode): number {
  const minRadius = 8
  const maxRadius = 30
  const maxDegree = Math.max(1, ...workingNodes.map(n => n.degree))
  return minRadius + (d.degree / maxDegree) * (maxRadius - minRadius)
}

function getFontSize(d: GraphNode): number {
  return Math.max(9, Math.min(14, 9 + d.degree))
}

function getCommunityColor(cid: number): string {
  const c = props.graphData?.communities.find(x => x.id === cid)
  return c?.color || '#999'
}

function getCommunityLabel(cid: number): string {
  const c = props.graphData?.communities.find(x => x.id === cid)
  return c?.label || t('knowledge.graph.communityLabel', { id: cid })
}

function darkenColor(hex: string, amount = 0.3): string {
  const c = d3.color(hex)
  return c ? (c.darker(amount) as d3.RGBColor).formatHex() : hex
}

// ── Selected neighbors ─────────────────────────────────
const selectedNeighbors = computed(() => {
  if (!selectedNode.value || !props.graphData) return []
  const sid = selectedNode.value.id
  const neighbors: { id: string; label: string; community: number; rel: string }[] = []
  for (const e of props.graphData.edges) {
    const srcId = typeof e.source === 'string' ? e.source : e.source.id
    const tgtId = typeof e.target === 'string' ? e.target : e.target.id
    if (srcId === sid) {
      const n = props.graphData.nodes.find(x => x.id === tgtId)
      if (n) neighbors.push({ id: n.id, label: n.label, community: n.community, rel: e.label || e.relationship })
    } else if (tgtId === sid) {
      const n = props.graphData.nodes.find(x => x.id === srcId)
      if (n) neighbors.push({ id: n.id, label: n.label, community: n.community, rel: e.label || e.relationship })
    }
  }
  return neighbors
})

// ── Visibility ─────────────────────────────────────────
function isNodeVisible(d: GraphNode): boolean {
  return visibleCommunities.value.includes(d.community) && visibleTypes.value.includes(d.type)
}

function isEdgeVisible(e: GraphEdge): boolean {
  const src = e.source as GraphNode
  const tgt = e.target as GraphNode
  return isNodeVisible(src) && isNodeVisible(tgt)
}

function updateVisibility() {
  if (!gRoot) return
  gRoot.selectAll<SVGGElement, GraphNode>('.kg-node')
    .attr('visibility', d => isNodeVisible(d) ? 'visible' : 'hidden')
  gRoot.selectAll<SVGLineElement, GraphEdge>('.kg-edge')
    .attr('visibility', d => isEdgeVisible(d) ? 'visible' : 'hidden')
}

// ── Search ─────────────────────────────────────────────
function handleSearch() {
  if (!gRoot || !searchQuery.value.trim()) {
    clearSearch()
    return
  }
  const q = searchQuery.value.toLowerCase()
  gRoot.selectAll<SVGGElement, GraphNode>('.kg-node')
    .classed('kg-dimmed', d => !d.label.toLowerCase().includes(q))
    .classed('kg-highlighted', d => d.label.toLowerCase().includes(q))
    .classed('kg-search-pulse', d => d.label.toLowerCase().includes(q))
  gRoot.selectAll<SVGLineElement, GraphEdge>('.kg-edge')
    .classed('kg-dimmed', true)

  // center on first match
  const match = workingNodes.find(n => n.label.toLowerCase().includes(q))
  if (match && match.x != null && match.y != null && svgSelection && zoomBehavior) {
    const wrap = svgWrapRef.value
    if (wrap) {
      const w = wrap.clientWidth
      const h = wrap.clientHeight
      svgSelection.transition().duration(500).call(
        zoomBehavior.transform as any,
        d3.zoomIdentity.translate(w / 2 - match.x, h / 2 - match.y)
      )
    }
  }
}

function clearSearch() {
  if (!gRoot) return
  gRoot.selectAll('.kg-node').classed('kg-dimmed', false).classed('kg-highlighted', false).classed('kg-search-pulse', false)
  gRoot.selectAll('.kg-edge').classed('kg-dimmed', false)
}

// ── Node selection ─────────────────────────────────────
function selectNode(d: GraphNode) {
  selectedNode.value = d
  if (!gRoot) return
  const sid = d.id
  const neighborIds = new Set<string>()
  workingEdges.forEach(e => {
    const srcId = (e.source as GraphNode).id
    const tgtId = (e.target as GraphNode).id
    if (srcId === sid) neighborIds.add(tgtId)
    if (tgtId === sid) neighborIds.add(srcId)
  })
  gRoot.selectAll<SVGGElement, GraphNode>('.kg-node')
    .classed('kg-dimmed', n => n.id !== sid && !neighborIds.has(n.id))
    .classed('kg-highlighted', n => n.id === sid)
  gRoot.selectAll<SVGLineElement, GraphEdge>('.kg-edge')
    .classed('kg-dimmed', e => {
      const srcId = (e.source as GraphNode).id
      const tgtId = (e.target as GraphNode).id
      return srcId !== sid && tgtId !== sid
    })
    .classed('kg-edge-highlight', e => {
      const srcId = (e.source as GraphNode).id
      const tgtId = (e.target as GraphNode).id
      return srcId === sid || tgtId === sid
    })
}

function deselectNode() {
  selectedNode.value = null
  if (!gRoot) return
  gRoot.selectAll('.kg-node').classed('kg-dimmed', false).classed('kg-highlighted', false)
  gRoot.selectAll('.kg-edge').classed('kg-dimmed', false).classed('kg-edge-highlight', false)
}

function focusNode(nodeId: string) {
  const n = workingNodes.find(x => x.id === nodeId)
  if (n) {
    selectNode(n)
    if (n.x != null && n.y != null && svgSelection && zoomBehavior && svgWrapRef.value) {
      const w = svgWrapRef.value.clientWidth
      const h = svgWrapRef.value.clientHeight
      svgSelection.transition().duration(500).call(
        zoomBehavior.transform as any,
        d3.zoomIdentity.translate(w / 2 - n.x, h / 2 - n.y)
      )
    }
  }
}

// ── Zoom controls ───────────────────────────────────────
function resetZoom() {
  if (svgSelection && zoomBehavior) {
    svgSelection.transition().duration(400).call(
      zoomBehavior.transform as any,
      d3.zoomIdentity
    )
  }
}

function zoomIn() {
  if (svgSelection && zoomBehavior) {
    svgSelection.transition().duration(300).call(
      zoomBehavior.scaleBy as any,
      1.4
    )
  }
}

function zoomOut() {
  if (svgSelection && zoomBehavior) {
    svgSelection.transition().duration(300).call(
      zoomBehavior.scaleBy as any,
      1 / 1.4
    )
  }
}

function toggleFullscreen() {
  const el = containerRef.value
  if (!el) return
  if (!document.fullscreenElement) {
    el.requestFullscreen?.().catch(() => {})
  } else {
    document.exitFullscreen?.().catch(() => {})
  }
}

// ── Build graph ────────────────────────────────────────
function buildGraph() {
  if (!props.graphData || !svgRef.value || !svgWrapRef.value) return

  // cleanup
  destroyGraph()

  const wrap = svgWrapRef.value
  const width = wrap.clientWidth
  const height = wrap.clientHeight

  // deep clone data for D3 mutation
  workingNodes = props.graphData.nodes.map(n => ({ ...n }))
  workingEdges = props.graphData.edges.map(e => ({ ...e }))

  // init filters
  visibleCommunities.value = props.graphData.communities.map(c => c.id)
  visibleTypes.value = [...nodeTypes]

  // SVG setup
  svgSelection = d3.select(svgRef.value)
    .attr('width', width)
    .attr('height', height)

  // Defs – arrowhead marker
  const defs = svgSelection.append('defs')
  defs.append('marker')
    .attr('id', 'kg-arrow')
    .attr('viewBox', '0 -5 10 10')
    .attr('refX', 20)
    .attr('refY', 0)
    .attr('markerWidth', 6)
    .attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path')
    .attr('d', 'M0,-5L10,0L0,5')
    .attr('fill', '#ccc')

  // Zoom
  zoomBehavior = d3.zoom<SVGSVGElement, unknown>()
    .scaleExtent([0.1, 5])
    .on('zoom', (event) => {
      if (gRoot) gRoot.attr('transform', event.transform)
    })
  svgSelection.call(zoomBehavior)

  gRoot = svgSelection.append('g')

  // Edges
  const edgeSelection = gRoot.selectAll<SVGLineElement, GraphEdge>('.kg-edge')
    .data(workingEdges)
    .enter()
    .append('line')
    .attr('class', 'kg-edge')
    .attr('stroke', '#ccc')
    .attr('stroke-width', 1)
    .attr('stroke-opacity', d => Math.max(0.2, d.confidence))
    .attr('marker-end', 'url(#kg-arrow)')

  // Node groups
  const nodeSelection = gRoot.selectAll<SVGGElement, GraphNode>('.kg-node')
    .data(workingNodes)
    .enter()
    .append('g')
    .attr('class', 'kg-node')
    .style('cursor', 'pointer')

  // Drag
  const dragBehavior = d3.drag<SVGGElement, GraphNode>()
    .on('start', (event, d) => {
      if (!event.active && simulation) simulation.alphaTarget(0.3).restart()
      d.fx = d.x
      d.fy = d.y
    })
    .on('drag', (event, d) => {
      d.fx = event.x
      d.fy = event.y
    })
    .on('end', (event, d) => {
      if (!event.active && simulation) simulation.alphaTarget(0)
      d.fx = null
      d.fy = null
    })
  nodeSelection.call(dragBehavior)

  // Circle
  nodeSelection.append('circle')
    .attr('r', d => getNodeRadius(d))
    .attr('fill', d => getCommunityColor(d.community))
    .attr('stroke', d => darkenColor(getCommunityColor(d.community)))
    .attr('stroke-width', 1.5)

  // Label
  nodeSelection.append('text')
    .text(d => d.label)
    .attr('dx', d => getNodeRadius(d) + 4)
    .attr('dy', 4)
    .attr('font-size', d => getFontSize(d))
    .attr('fill', '#333')
    .attr('pointer-events', 'none')

  // Hover – highlight neighbors
  nodeSelection
    .on('mouseenter', (event: MouseEvent, d: GraphNode) => {
      const rect = svgWrapRef.value!.getBoundingClientRect()
      tooltip.value = {
        visible: true,
        x: event.clientX - rect.left + 12,
        y: event.clientY - rect.top - 8,
        label: d.label,
        type: d.type,
        degree: d.degree,
        desc: d.description?.slice(0, 120) || ''
      }
      // Highlight node and its direct neighbors
      if (!selectedNode.value) {
        const sid = d.id
        const neighborIds = new Set<string>()
        workingEdges.forEach(e => {
          const srcId = (e.source as GraphNode).id
          const tgtId = (e.target as GraphNode).id
          if (srcId === sid) neighborIds.add(tgtId)
          if (tgtId === sid) neighborIds.add(srcId)
        })
        gRoot!.selectAll<SVGGElement, GraphNode>('.kg-node')
          .classed('kg-dimmed', n => n.id !== sid && !neighborIds.has(n.id))
        gRoot!.selectAll<SVGLineElement, GraphEdge>('.kg-edge')
          .classed('kg-dimmed', e => {
            const srcId = (e.source as GraphNode).id
            const tgtId = (e.target as GraphNode).id
            return srcId !== sid && tgtId !== sid
          })
      }
    })
    .on('mouseleave', () => {
      tooltip.value.visible = false
      // Remove hover highlight if no node is selected
      if (!selectedNode.value && gRoot) {
        gRoot.selectAll('.kg-node').classed('kg-dimmed', false)
        gRoot.selectAll('.kg-edge').classed('kg-dimmed', false)
      }
    })
    .on('click', (_event: MouseEvent, d: GraphNode) => {
      selectNode(d)
    })
    .on('dblclick', (_event: MouseEvent, d: GraphNode) => {
      // Double-click: navigate to document if source_docs exist
      if (d.source_docs && d.source_docs.length > 0) {
        router.push({ path: `/document/${d.source_docs[0]}` })
      }
    })

  // Click SVG background to deselect
  svgSelection.on('click', (event) => {
    if (event.target === svgRef.value) deselectNode()
  })

  // Simulation – optimized force parameters
  simulation = d3.forceSimulation<GraphNode>(workingNodes)
    .force('link', d3.forceLink<GraphNode, GraphEdge>(workingEdges).id(d => d.id).distance(80))
    .force('charge', d3.forceManyBody().strength(-300))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide<GraphNode>().radius(d => getNodeRadius(d) + 5))
    .on('tick', () => {
      edgeSelection
        .attr('x1', d => (d.source as GraphNode).x!)
        .attr('y1', d => (d.source as GraphNode).y!)
        .attr('x2', d => (d.target as GraphNode).x!)
        .attr('y2', d => (d.target as GraphNode).y!)
      nodeSelection
        .attr('transform', d => `translate(${d.x},${d.y})`)
    })
}

function destroyGraph() {
  if (simulation) { simulation.stop(); simulation = null }
  if (svgRef.value) d3.select(svgRef.value).selectAll('*').remove()
  gRoot = null
  svgSelection = null
}

// ── Resize ─────────────────────────────────────────────
function handleResize() {
  if (!svgWrapRef.value || !svgSelection || !simulation) return
  const w = svgWrapRef.value.clientWidth
  const h = svgWrapRef.value.clientHeight
  svgSelection.attr('width', w).attr('height', h)
  simulation.force('center', d3.forceCenter(w / 2, h / 2))
  simulation.alpha(0.3).restart()
}

// ── Lifecycle ──────────────────────────────────────────
onMounted(() => {
  if (props.graphData && props.graphData.nodes.length) {
    nextTick(() => buildGraph())
  }
  if (svgWrapRef.value) {
    resizeObserver = new ResizeObserver(() => handleResize())
    resizeObserver.observe(svgWrapRef.value)
  }
})

watch(() => props.graphData, (val) => {
  if (val && val.nodes.length) {
    nextTick(() => buildGraph())
  } else {
    destroyGraph()
  }
}, { deep: false })

onBeforeUnmount(() => {
  destroyGraph()
  if (resizeObserver) { resizeObserver.disconnect(); resizeObserver = null }
})
</script>

<style scoped>
.kg-graph-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  position: relative;
  background: #fafbfc;
  border-radius: 8px;
  overflow: hidden;
}

/* ── Loading / Empty ── */
.kg-loading,
.kg-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #909399;
}
.kg-loading-icon {
  animation: kg-spin 1s linear infinite;
}
@keyframes kg-spin {
  to { transform: rotate(360deg); }
}

/* ── Toolbar ── */
.kg-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
  flex-shrink: 0;
}
.kg-search {
  width: 200px;
}
.kg-stats {
  display: flex;
  gap: 6px;
}

/* ── Main ── */
.kg-main {
  flex: 1;
  display: flex;
  position: relative;
  overflow: hidden;
}
.kg-svg-wrap {
  flex: 1;
  position: relative;
  overflow: hidden;
}
.kg-svg {
  display: block;
  width: 100%;
  height: 100%;
  background: #fafbfc;
}

/* ── Tooltip ── */
.kg-tooltip {
  position: absolute;
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.8);
  color: #fff;
  border-radius: 6px;
  font-size: 12px;
  pointer-events: none;
  z-index: 10;
  max-width: 260px;
}
.kg-tooltip-label {
  font-weight: 600;
  margin-bottom: 2px;
}
.kg-tooltip-type {
  font-size: 11px;
  opacity: 0.8;
}
.kg-tooltip-degree {
  font-size: 11px;
  opacity: 0.8;
  margin-top: 2px;
}
.kg-tooltip-desc {
  margin-top: 4px;
  font-size: 11px;
  opacity: 0.7;
  line-height: 1.4;
}

/* ── Zoom Controls ── */
.graph-controls {
  position: absolute;
  bottom: 16px;
  right: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  z-index: 5;
}
.graph-controls button {
  width: 36px;
  height: 30px;
  border: 1px solid #dcdfe6;
  background: #fff;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: #606266;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.graph-controls button:hover {
  background: #f5f7fa;
  border-color: #409eff;
  color: #409eff;
}

/* ── Legend Overlay (top-left) ── */
.kg-legend-overlay {
  position: absolute;
  top: 12px;
  left: 12px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 11px;
  color: #606266;
  z-index: 5;
  max-width: 200px;
  backdrop-filter: blur(4px);
}
.kg-legend-title {
  font-weight: 600;
  font-size: 12px;
  margin-bottom: 6px;
  color: #303133;
}
.kg-legend-items {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 10px;
  margin-bottom: 6px;
}
.kg-legend-size {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
  padding-top: 4px;
  border-top: 1px solid #ebeef5;
}
.kg-legend-size-small {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #c0c4cc;
}
.kg-legend-size-large {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #c0c4cc;
}
.kg-legend-size-label {
  font-size: 10px;
  color: #909399;
}
.kg-legend-core {
  margin-top: 4px;
  padding-top: 4px;
  border-top: 1px solid #ebeef5;
  color: #e6a23c;
  display: flex;
  align-items: center;
  gap: 4px;
}

/* ── Detail Panel ── */
.kg-detail {
  width: 280px;
  flex-shrink: 0;
  border-left: 1px solid #ebeef5;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}
.kg-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid #ebeef5;
  font-weight: 600;
  font-size: 15px;
}
.kg-detail-close {
  cursor: pointer;
  color: #909399;
  transition: color 0.15s;
}
.kg-detail-close:hover {
  color: #f54a45;
}
.kg-detail-body {
  padding: 12px 14px;
}
.kg-detail-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 10px;
}
.kg-detail-key {
  font-size: 12px;
  color: #909399;
  min-width: 40px;
  flex-shrink: 0;
}
.kg-detail-val {
  font-size: 13px;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 4px;
}
.kg-detail-section {
  margin-top: 14px;
  padding-top: 10px;
  border-top: 1px solid #ebeef5;
}
.kg-neighbor-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 8px;
  font-size: 13px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.15s;
}
.kg-neighbor-item:hover {
  background: #f5f7fa;
}
.kg-source-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  font-size: 12px;
  color: #606266;
}

/* ── Legend ── */
.kg-legend {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 14px;
  border-top: 1px solid #ebeef5;
  background: #fff;
  flex-shrink: 0;
  flex-wrap: wrap;
  font-size: 12px;
  color: #606266;
}
.kg-legend-item {
  display: flex;
  align-items: center;
  gap: 4px;
}
.kg-legend-dot,
.kg-filter-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.kg-legend-sep {
  color: #dcdfe6;
}
.kg-legend-god {
  color: #e6a23c;
}

/* ── Filter ── */
.kg-filter-panel {
  max-height: 320px;
  overflow-y: auto;
}
.kg-filter-section {
  margin-bottom: 12px;
}
.kg-filter-section:last-child {
  margin-bottom: 0;
}
.kg-filter-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
}
.kg-filter-section .el-checkbox {
  display: flex;
  margin-right: 0;
}

/* ── D3 element states ── */
:deep(.kg-node.kg-dimmed) {
  opacity: 0.15;
}
:deep(.kg-node.kg-highlighted circle) {
  stroke-width: 3;
  filter: drop-shadow(0 0 4px rgba(51, 112, 255, 0.4));
}
:deep(.kg-node.kg-search-pulse circle) {
  stroke-width: 3;
  stroke: #409eff;
  animation: kg-pulse 1.2s ease-in-out infinite;
}
@keyframes kg-pulse {
  0%, 100% { filter: drop-shadow(0 0 3px rgba(64, 158, 255, 0.4)); }
  50% { filter: drop-shadow(0 0 10px rgba(64, 158, 255, 0.8)); transform: scale(1.08); }
}
:deep(.kg-edge.kg-dimmed) {
  opacity: 0.05;
}
:deep(.kg-edge.kg-edge-highlight) {
  stroke: #3370ff !important;
  stroke-width: 2 !important;
  stroke-opacity: 1 !important;
}

/* ── Slide transition ── */
.kg-slide-enter-active,
.kg-slide-leave-active {
  transition: width 0.25s ease, opacity 0.25s ease;
}
.kg-slide-enter-from,
.kg-slide-leave-to {
  width: 0;
  opacity: 0;
}
</style>
