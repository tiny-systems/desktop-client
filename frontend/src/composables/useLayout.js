import dagre from '@dagrejs/dagre'
import { Position, useVueFlow } from '@vue-flow/core'

export function useLayout() {
  const { findNode } = useVueFlow()

  /**
   * Apply dagre layout algorithm to nodes
   * @param nodes - Array of nodes to layout
   * @param edges - Array of edges connecting nodes
   * @param direction - Layout direction: 'LR' (left-right) or 'TB' (top-bottom)
   * @returns Nodes with updated positions
   */
  function layout(nodes, edges, direction = 'LR') {
    const dagreGraph = new dagre.graphlib.Graph()
    dagreGraph.setDefaultEdgeLabel(() => ({}))

    // Configure graph layout
    dagreGraph.setGraph({
      rankdir: direction,
      nodesep: 80,    // Horizontal spacing between nodes
      ranksep: 150,   // Vertical spacing between ranks (columns in LR)
      marginx: 50,
      marginy: 50
    })

    const isHorizontal = direction === 'LR'

    // Add nodes to dagre graph
    for (const node of nodes) {
      // Get actual node dimensions from Vue Flow
      const graphNode = findNode(node.id)
      const width = graphNode?.dimensions?.width || 180
      const height = graphNode?.dimensions?.height || 80

      dagreGraph.setNode(node.id, { width, height })
    }

    // Add edges to dagre graph
    for (const edge of edges) {
      dagreGraph.setEdge(edge.source, edge.target)
    }

    // Run the layout algorithm
    dagre.layout(dagreGraph)

    // Apply calculated positions to nodes
    return nodes.map((node) => {
      const nodeWithPosition = dagreGraph.node(node.id)
      if (!nodeWithPosition) {
        return node
      }

      // Dagre returns center position, adjust to top-left corner
      const graphNode = findNode(node.id)
      const width = graphNode?.dimensions?.width || 180
      const height = graphNode?.dimensions?.height || 80

      return {
        ...node,
        // Update handle positions based on layout direction
        targetPosition: isHorizontal ? Position.Left : Position.Top,
        sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
        position: {
          x: nodeWithPosition.x - width / 2,
          y: nodeWithPosition.y - height / 2
        },
      }
    })
  }

  return { layout }
}
