import { defineStore } from 'pinia'
import { isEdge, isNode } from '@vue-flow/core'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const GoApp = window.go?.main?.App

function clone(obj) {
  try {
    return JSON.parse(JSON.stringify(obj))
  } catch (error) {
    return obj
  }
}

export const useFlowStore = defineStore('flowStore', {
  state() {
    return {
      loading: false,
      loadingAlt: false,
      ready: false,
      flowResourceName: '',
      flowName: '',
      projectResourceName: '',
      projectName: '',
      contextName: '',
      namespace: '',
      mutate: 0,
      elements: [],
      meta: {},
      lastUpdate: null,
      animationCheckInterval: null,
      watching: false,
      trace: null // Selected trace ID for using real runtime data
    }
  },
  getters: {
    selectedEdges: (state) => state.elements.filter((a) => (a.selected || false) && isEdge(a)),
    selectedEdge: (state) => {
      const edges = state.elements.filter((a) => (a.selected || false) && isEdge(a))
      return edges.length > 0 ? edges[0] : undefined
    },
    selectedEdgeData() {
      return this.selectedEdge?.data
    },
    selectedEdgeTargetHandle() {
      if (!this.selectedEdge?.targetNode?.data?.handles) {
        return undefined
      }
      return this.selectedEdge.targetNode.data.handles.find(
        (a) => a.id === this.selectedEdge.targetHandle
      )
    },
    selectedConfiguration() {
      const edgeData = this.selectedEdgeData
      if (edgeData?.configuration) {
        return edgeData.configuration
      }
      let handle = this.selectedEdgeTargetHandle
      if (!handle) {
        handle = this.settingsHandle
      }
      if (!handle) {
        return '{}'
      }
      return handle.configuration || '{}'
    },
    selectedSchema() {
      const edgeData = this.selectedEdgeData
      if (edgeData?.schema) {
        return edgeData.schema
      }
      let handle = this.selectedEdgeTargetHandle
      if (!handle) {
        handle = this.settingsHandle
      }
      if (!handle) {
        return '{}'
      }
      return handle.schema || '{}'
    },
    selectedControl() {
      const handle = this.controlHandle
      if (!handle) {
        return undefined
      }
      return handle.configuration || '{}'
    },
    selectedNodes: (state) => state.elements.filter((a) => (a.selected || false) && isNode(a)),
    selectedNode() {
      return this.selectedNodes.length > 0 ? this.selectedNodes[0] : undefined
    },
    selectedNodeId() {
      return this.selectedNode?.id || ''
    },
    selectedNodeLabel() {
      return this.selectedNode?.data?.label || this.selectedNode?.id || ''
    },
    controlHandle() {
      if (!this.selectedNode?.data?.handles) {
        return undefined
      }
      return this.selectedNode.data.handles.find((a) => a.id === '_control')
    },
    controlHandleSchema() {
      return this.controlHandle?.schema || '{}'
    },
    settingsHandle() {
      if (!this.selectedNode?.data?.handles) {
        return undefined
      }
      return this.selectedNode.data.handles.find((a) => a.id === '_settings')
    },
    settingsHandleSchema() {
      return this.settingsHandle?.schema || '{}'
    },
    selectedHandle() {
      if (!this.selectedNode?.data?.handles?.length) {
        return undefined
      }
      let sel = this.selectedNode.selectedHandleId
      if (this.selectedNode.data.trace?.port && !sel) {
        sel = this.selectedNode.data.trace.port
      }
      let selected = this.selectedNode.data.handles.find((a) => a.id === sel)
      if (!selected) {
        selected = this.selectedNodeHandles[0]
      }
      return selected
    },
    selectedNodeHandles() {
      return this.selectedNode?.data?.handles || []
    },
    nodes: (state) => state.elements.filter((el) => isNode(el)),
    edges: (state) => state.elements.filter((el) => isEdge(el))
  },
  actions: {
    checkStaleAnimations() {
      const now = Date.now() / 1000
      for (let i = 0; i < this.elements.length; i++) {
        if (this.elements[i].data?.stats) {
          const busyTimestamp = parseInt(this.elements[i].data.stats['tiny_edge_busy'] || 0)
          const timeSinceActivity = now - busyTimestamp
          this.elements[i].animated = timeSinceActivity < 7
        }
      }
    },
    startAnimationCheck() {
      if (this.animationCheckInterval === null) {
        this.animationCheckInterval = setInterval(() => {
          this.checkStaleAnimations()
        }, 1000)
      }
    },
    stopAnimationCheck() {
      if (this.animationCheckInterval !== null) {
        clearInterval(this.animationCheckInterval)
        this.animationCheckInterval = null
      }
    },
    select(id) {
      for (let i = 0; i < this.elements.length; i++) {
        this.elements[i].selected = this.elements[i].id === id
      }
    },
    selectElement(id) {
      this.elements.forEach((a) => {
        a.selected = id && a.id === id
      })
    },
    addElement(element) {
      if (!element.graph) {
        return
      }
      // Check for duplicates - don't add if element already exists
      const existingIndex = this.elements.findIndex(el => el.id === element.graph.id)
      if (existingIndex !== -1) {
        // Update existing element instead of adding duplicate
        // Preserve existing stats before merging
        const existingStats = this.elements[existingIndex].data?.stats
        Object.assign(this.elements[existingIndex].data || {}, element.graph.data || {})
        // Merge stats: preserve existing stats and layer new stats on top
        if (existingStats || element.graph.data?.stats) {
          this.elements[existingIndex].data.stats = Object.assign({}, existingStats, element.graph.data?.stats)
        }
        if (element.graph.position) {
          this.elements[existingIndex].position = element.graph.position
        }
        return
      }

      // Ensure edges have valid: true and stats initialized by default
      if (isEdge(element.graph)) {
        if (!element.graph.data) {
          element.graph.data = {}
        }
        if (element.graph.data.valid === undefined) {
          element.graph.data.valid = true
        }
        if (!element.graph.data.stats) {
          element.graph.data.stats = {}
        }
      }

      if (element.graph.data?.blocked || element.graph.data?.disabled) {
        element.graph.draggable = false
        if (isEdge(element.graph)) {
          element.graph.selectable = false
        }
      }
      this.elements.push(element.graph)
    },
    updateElement(event) {
      for (let i = 0; i < this.elements.length; i++) {
        if (this.elements[i].id === event.id) {
          // Preserve existing stats before updating data
          const existingStats = this.elements[i].data?.stats
          this.elements[i].data = { ...event.graph.data }
          // Merge stats: preserve existing stats and layer new stats on top
          if (existingStats || event.graph.data?.stats) {
            this.elements[i].data.stats = Object.assign({}, existingStats, event.graph.data?.stats)
          }
          // Update position if provided
          if (event.graph.position) {
            this.elements[i].position = event.graph.position
          }
          return true
        }
      }
      return false
    },
    update(element) {
      for (let i = 0; i < this.elements.length; i++) {
        if (this.elements[i].id === element.id) {
          this.elements[i] = element
        }
      }
    },
    deleteElementSilent(id) {
      this.elements = this.elements.filter((a) => a.id !== id)
    },
    deleteElementHidden() {
      this.elements = this.elements.filter((a) => !a.hidden)
    },
    getElement(id) {
      return this.elements.find((a) => a.id === id)
    },
    deleteSelected() {
      this.$patch((state) => {
        state.elements = this.elements.filter((a) => !a.selected)
      })
    },
    applyStats(event) {
      for (let i = 0; i < this.elements.length; i++) {
        for (let key in event.graph) {
          if (this.elements[i].id !== key) continue
          if (!Object.hasOwn(event.graph, key)) continue
          const newStats = event.graph[key]
          if (!this.elements[i].data) {
            this.elements[i].data = {}
          }
          if (!this.elements[i].data.stats) {
            this.elements[i].data.stats = {}
          }
          Object.assign(this.elements[i].data.stats, newStats)
        }
      }
      this.checkStaleAnimations()
    },
    processNodeEvent(event) {
      if (!event) return

      const type = event.type || ''
      const id = event.id || ''
      const graphData = event.graph

      this.lastUpdate = { id, type, graph: graphData }

      switch (type) {
        case 'ADDED':
          this.addElement({ id, graph: graphData })
          break
        case 'MODIFIED':
          if (!this.updateElement({ id, graph: graphData })) {
            // If element doesn't exist, add it
            this.addElement({ id, graph: graphData })
          }
          break
        case 'DELETED':
          this.deleteElementSilent(id)
          break
        case 'STATS':
          this.applyStats({ graph: graphData })
          break
        default:
          if (graphData) {
            this.addElement({ id, graph: graphData })
          }
      }
    },
    async load(contextName, namespace, projectName, flowResourceName) {
      if (!GoApp) {
        throw new Error('Wails runtime not available')
      }

      this.loading = true
      this.contextName = contextName
      this.namespace = namespace
      this.projectResourceName = projectName
      this.flowResourceName = flowResourceName

      try {
        const data = await GoApp.GetFlowForEditor(contextName, namespace, projectName, flowResourceName)

        this.flowName = data.flow.name
        this.flowResourceName = data.flow.resourceName
        this.projectName = data.project.name || projectName
        this.projectResourceName = data.project.resourceName || projectName

        // Load metadata (viewport position)
        if (data.meta) {
          this.meta = data.meta
        }

        // Process elements
        this.elements = []
        if (data.elements) {
          data.elements.forEach((el) => {
            this.addElement({ id: el.id, graph: el })
          })
        }

        this.ready = true
        return data
      } finally {
        this.loading = false
      }
    },
    async startWatching() {
      if (!GoApp || this.watching) return

      try {
        await GoApp.WatchFlowNodes(
          this.contextName,
          this.namespace,
          this.projectResourceName,
          this.flowResourceName
        )

        EventsOn('flowNodeUpdate', (event) => {
          this.processNodeEvent(event)
        })

        this.watching = true
        this.startAnimationCheck()
      } catch (err) {
        console.error('Failed to start watching:', err)
      }
    },
    async stopWatching() {
      if (!GoApp || !this.watching) return

      try {
        EventsOff('flowNodeUpdate')
        await GoApp.StopWatchFlowNodes()
        this.watching = false
        this.stopAnimationCheck()
      } catch (err) {
        console.error('Failed to stop watching:', err)
      }
    },
    clean() {
      this.elements = []
      this.trace = null
      this.stopAnimationCheck()
      this.stopWatching()
      this.ready = false
    },
    setMeta(meta) {
      this.meta = meta
    },
    async saveMeta() {
      if (!GoApp || !this.flowResourceName || !this.contextName) return

      if (this.meta?.x === undefined || this.meta?.y === undefined || this.meta?.zoom === undefined) {
        return
      }

      try {
        await GoApp.SaveFlowMeta(
          this.contextName,
          this.namespace,
          this.flowResourceName,
          this.meta.x,
          this.meta.y,
          this.meta.zoom
        )
      } catch (e) {
        console.error('Failed to save meta:', e)
      }
    },
    async addNode(componentName, moduleName, moduleVersion, posX, posY) {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        const nodeElement = await GoApp.AddNode(
          this.contextName,
          this.namespace,
          this.projectResourceName,
          this.flowResourceName,
          componentName,
          moduleName,
          moduleVersion,
          posX,
          posY
        )

        this.addElement({ id: nodeElement.id, graph: nodeElement })
        return nodeElement
      } finally {
        this.loadingAlt = false
      }
    },
    async deleteNode(nodeId) {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        await GoApp.DeleteNode(this.contextName, this.namespace, nodeId)
        this.deleteElementSilent(nodeId)

        // Also delete edges connected to this node
        const edgesToDelete = this.elements.filter(
          (el) => isEdge(el) && (el.source === nodeId || el.target === nodeId)
        )
        edgesToDelete.forEach((edge) => {
          this.deleteElementSilent(edge.id)
        })
      } finally {
        this.loadingAlt = false
      }
    },
    async updateNodePosition(nodeId, x, y) {
      if (!GoApp) return

      try {
        await GoApp.UpdateNodePosition(this.contextName, this.namespace, nodeId, x, y)
      } catch (e) {
        console.error('Failed to update node position:', e)
      }
    },
    async batchUpdateNodePositions(positions) {
      if (!GoApp) return

      try {
        await GoApp.BatchUpdateNodePositions(this.contextName, this.namespace, positions)
      } catch (e) {
        console.error('Failed to batch update positions:', e)
      }
    },
    async updateNodeLabel(nodeId, label) {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        await GoApp.UpdateNodeLabel(this.contextName, this.namespace, nodeId, label)

        // Update local state
        const node = this.getElement(nodeId)
        if (node?.data) {
          node.data.label = label
        }
      } finally {
        this.loadingAlt = false
      }
    },
    async updateNodeComment(nodeId, comment) {
      if (!GoApp) throw new Error('Wails runtime not available')

      try {
        await GoApp.UpdateNodeComment(this.contextName, this.namespace, nodeId, comment)

        const node = this.getElement(nodeId)
        if (node?.data) {
          node.data.comment = comment
        }
      } catch (e) {
        console.error('Failed to update node comment:', e)
      }
    },
    async rotateNode(nodeId) {
      if (!GoApp) throw new Error('Wails runtime not available')

      try {
        await GoApp.RotateNode(this.contextName, this.namespace, nodeId)

        // Update local state
        const node = this.getElement(nodeId)
        if (node?.data) {
          node.data.spin = ((node.data.spin || 0) + 1) % 4
          if (node.data.handles) {
            node.data.handles.forEach((h) => {
              h.rotated_position = (h.position + node.data.spin) % 4
            })
          }
        }
      } catch (e) {
        console.error('Failed to rotate node:', e)
      }
    },
    async rotate() {
      for (const node of this.selectedNodes) {
        await this.rotateNode(node.id)
      }
    },
    async toggleNodeDashboard(nodeId, enabled) {
      if (!GoApp) throw new Error('Wails runtime not available')

      try {
        await GoApp.ToggleNodeDashboard(this.contextName, this.namespace, nodeId, enabled)

        const node = this.getElement(nodeId)
        if (node?.data) {
          // Use string "true" for consistency with platform
          node.data.dashboard = enabled ? 'true' : ''
        }
      } catch (e) {
        console.error('Failed to toggle dashboard:', e)
      }
    },
    async updateNodeConfiguration(nodeId, port, configuration, schema) {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        const configStr = typeof configuration === 'string' ? configuration : JSON.stringify(configuration)
        const schemaStr = schema ? (typeof schema === 'string' ? schema : JSON.stringify(schema)) : ''

        await GoApp.UpdateNodeConfiguration(this.contextName, this.namespace, nodeId, port, configStr, schemaStr)
      } finally {
        this.loadingAlt = false
      }
    },
    async connectNodes(sourceNode, sourcePort, targetNode, targetPort, configuration = '') {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        const configStr = typeof configuration === 'string' ? configuration : JSON.stringify(configuration || {})

        await GoApp.ConnectNodes(
          this.contextName,
          this.namespace,
          this.flowResourceName,
          sourceNode,
          sourcePort,
          targetNode,
          targetPort,
          configStr
        )

        // Edge will be added via the watcher
      } finally {
        this.loadingAlt = false
      }
    },
    async disconnectNodes(sourceNode, edgeId) {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        await GoApp.DisconnectNodes(this.contextName, this.namespace, sourceNode, edgeId)
        this.deleteElementSilent(edgeId)
      } finally {
        this.loadingAlt = false
      }
    },
    async updateEdgeConfiguration(sourceNode, sourcePort, targetTo, configuration, flowId) {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        const configStr = typeof configuration === 'string' ? configuration : JSON.stringify(configuration)

        await GoApp.UpdateEdgeConfiguration(
          this.contextName,
          this.namespace,
          sourceNode,
          sourcePort,
          targetTo,
          configStr,
          flowId || this.flowResourceName
        )
      } finally {
        this.loadingAlt = false
      }
    },
    async runAction(nodeId, port, data) {
      if (!GoApp) throw new Error('Wails runtime not available')

      this.loadingAlt = true
      try {
        const dataStr = typeof data === 'string' ? data : JSON.stringify(data)
        return await GoApp.RunNodeAction(this.contextName, this.namespace, nodeId, port, dataStr)
      } finally {
        this.loadingAlt = false
      }
    },
    async inspectNodePort(nodeId, port, traceID = null) {
      if (!GoApp) throw new Error('Wails runtime not available')

      // Use stored trace if no specific trace ID provided
      const effectiveTraceID = traceID ?? this.trace ?? ''
      return await GoApp.InspectNodePort(this.contextName, this.namespace, this.projectResourceName, nodeId, port, effectiveTraceID)
    },
    async getNodeHandles(nodeId) {
      if (!GoApp) throw new Error('Wails runtime not available')

      return await GoApp.GetNodeHandles(this.contextName, this.namespace, nodeId)
    },
    async getAvailableComponents() {
      if (!GoApp) throw new Error('Wails runtime not available')

      return await GoApp.GetAvailableComponents(this.contextName, this.namespace)
    },
    export() {
      let copy = clone(this.elements)
      Object.keys(copy).forEach((key) => {
        delete copy[key]['events']
        delete copy[key]['sourceNode']
        delete copy[key]['targetNode']
        delete copy[key]['isParent']
        delete copy[key]['dragging']
        delete copy[key]['initialized']
        delete copy[key]['selected']
        delete copy[key]['resizing']
        delete copy[key]['computedPosition']
        delete copy[key]['labelBgStyle']
        delete copy[key]['handleBounds']
        if (isEdge(copy[key])) {
          delete copy[key]['data']?.['error']
          delete copy[key]['animated']
          delete copy[key]['sourceX']
          delete copy[key]['sourceY']
          delete copy[key]['targetX']
          delete copy[key]['targetY']
        } else if (isNode(copy[key])) {
          delete copy[key]['data']?.['stats']
          delete copy[key]['data']?.['emit']
          delete copy[key]['data']?.['blocked']
          delete copy[key]['data']?.['status']
          delete copy[key]['data']?.['error']
          delete copy[key]['data']?.['emitting']
          delete copy[key]['data']?.['disabled']
          delete copy[key]['data']?.['last_status_update']
          if (copy[key]['data']?.['handles']) {
            Object.keys(copy[key]['data']['handles']).forEach((keyh) => {
              delete copy[key]['data']['handles'][keyh]['style']
              delete copy[key]['data']['handles'][keyh]['class']
            })
          }
        }
      })
      return copy
    },
    async import(elements) {
      this.loading = true
      elements.forEach((el) => {
        el.hidden = true
        this.addElement({ id: el.id, graph: el })
      })
      try {
        await this.save()
      } catch (e) {
        console.error('Import error:', e)
      }
      this.loading = false
    },
    up() {
      this.$patch((state) => {
        state.mutate++
      })
    },
    highlightTrace(traceId) {
      // Deselect all elements first
      this.elements.forEach((el) => {
        el.selected = false
      })

      if (!traceId) return

      // Highlight edges that match the trace ID
      this.elements.forEach((el) => {
        if (isEdge(el) && el.data?.trace?.traceID === traceId) {
          el.selected = true
        }
      })
    },
    async setTrace(traceId) {
      this.trace = traceId

      if (!traceId || !GoApp) {
        this.highlightTrace(traceId)
        return
      }

      try {
        // Fetch graph elements with trace stats applied
        const result = await GoApp.ApplyTraceToFlow(
          this.contextName,
          this.namespace,
          this.projectResourceName,
          this.flowResourceName,
          traceId
        )

        // Apply trace data to nodes
        if (result.nodes) {
          for (const nodeData of result.nodes) {
            const existing = this.elements.find(el => el.id === nodeData.id)
            if (existing && existing.data) {
              existing.data.trace = nodeData.data?.trace
            }
          }
        }

        // Apply trace data and styles to edges
        if (result.edges) {
          for (const edgeData of result.edges) {
            const existing = this.elements.find(el => el.id === edgeData.id)
            if (existing) {
              if (existing.data) {
                existing.data.trace = edgeData.data?.trace
              }
              // Apply edge styling from trace
              if (edgeData.style) {
                existing.style = { ...existing.style, ...edgeData.style }
              }
            }
          }
        }
      } catch (err) {
        console.error('Failed to apply trace to flow:', err)
      }

      this.highlightTrace(traceId)
    },
    clearTrace() {
      // Clear trace data from all elements
      for (const el of this.elements) {
        if (el.data?.trace) {
          delete el.data.trace
        }
        // Reset edge styles
        if (isEdge(el) && el.style) {
          delete el.style.stroke
          delete el.style.strokeWidth
        }
      }
      this.trace = null
      this.highlightTrace(null)
    },
    async runExpression(expression, data, schema) {
      if (!GoApp) throw new Error('Wails runtime not available')

      const dataStr = typeof data === 'string' ? data : JSON.stringify(data)
      const schemaStr = typeof schema === 'string' ? schema : JSON.stringify(schema || {})

      return await GoApp.RunExpression(expression, dataStr, schemaStr)
    },
    // Get source data for an edge (data from the source node's output port)
    getEdgeSourceData(edge) {
      if (!edge) return null

      const sourceNode = this.getElement(edge.source)
      if (!sourceNode?.data?.handles) return null

      const sourceHandle = sourceNode.data.handles.find(h => h.id === edge.sourceHandle)
      if (!sourceHandle) return null

      // Try to get the sample data from the handle
      // This might be available from port inspection or default data
      try {
        if (sourceHandle.data) {
          return typeof sourceHandle.data === 'string' ? JSON.parse(sourceHandle.data) : sourceHandle.data
        }
        // Fallback to default from schema
        if (sourceHandle.schema) {
          const schema = typeof sourceHandle.schema === 'string' ? JSON.parse(sourceHandle.schema) : sourceHandle.schema
          return schema.default || {}
        }
      } catch {
        return {}
      }
      return {}
    }
  }
})
