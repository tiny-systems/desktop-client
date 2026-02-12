export namespace kubernetes {
	
	export class Clientset {
	    LegacyPrefix: string;
	    UseLegacyDiscovery: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Clientset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.LegacyPrefix = source["LegacyPrefix"];
	        this.UseLegacyDiscovery = source["UseLegacyDiscovery"];
	    }
	}

}

export namespace main {
	
	export class ApplyTraceToFlowResponse {
	    nodes: any[];
	    edges: any[];
	
	    static createFrom(source: any = {}) {
	        return new ApplyTraceToFlowResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodes = source["nodes"];
	        this.edges = source["edges"];
	    }
	}
	export class BuildInfo {
	    buildTime: string;
	    version: string;
	    sdkVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.buildTime = source["buildTime"];
	        this.version = source["version"];
	        this.sdkVersion = source["sdkVersion"];
	    }
	}
	export class ComponentInfo {
	    name: string;
	    module: string;
	    version: string;
	    description: string;
	    info: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new ComponentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.module = source["module"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.info = source["info"];
	        this.tags = source["tags"];
	    }
	}
	export class Flow {
	    name: string;
	    resourceName: string;
	    nodeCount: number;
	    graph?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Flow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.resourceName = source["resourceName"];
	        this.nodeCount = source["nodeCount"];
	        this.graph = source["graph"];
	    }
	}
	export class ProjectInfo {
	    name: string;
	    resourceName: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.resourceName = source["resourceName"];
	    }
	}
	export class FlowInfo {
	    name: string;
	    resourceName: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.resourceName = source["resourceName"];
	        this.description = source["description"];
	    }
	}
	export class FlowEditorData {
	    flow: FlowInfo;
	    project: ProjectInfo;
	    elements: any[];
	    meta: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new FlowEditorData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flow = this.convertValues(source["flow"], FlowInfo);
	        this.project = this.convertValues(source["project"], ProjectInfo);
	        this.elements = source["elements"];
	        this.meta = source["meta"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class KubeContext {
	    name: string;
	    cluster: string;
	    user: string;
	    current: boolean;
	
	    static createFrom(source: any = {}) {
	        return new KubeContext(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.cluster = source["cluster"];
	        this.user = source["user"];
	        this.current = source["current"];
	    }
	}
	export class ModuleComponent {
	    name: string;
	    description: string;
	    info: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new ModuleComponent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.info = source["info"];
	        this.tags = source["tags"];
	    }
	}
	export class Module {
	    name: string;
	    version: string;
	    sdkVersion: string;
	    components: ModuleComponent[];
	
	    static createFrom(source: any = {}) {
	        return new Module(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.sdkVersion = source["sdkVersion"];
	        this.components = this.convertValues(source["components"], ModuleComponent);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Preferences {
	    lastContext: string;
	    lastNamespace: string;
	
	    static createFrom(source: any = {}) {
	        return new Preferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastContext = source["lastContext"];
	        this.lastNamespace = source["lastNamespace"];
	    }
	}
	export class PreviewEdgeMappingResult {
	    result: string;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new PreviewEdgeMappingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result = source["result"];
	        this.errors = source["errors"];
	    }
	}
	export class Project {
	    name: string;
	    title: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
	        this.description = source["description"];
	    }
	}
	export class ProjectDetails {
	    name: string;
	    title: string;
	    description: string;
	    resourceName: string;
	    clusterName: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.resourceName = source["resourceName"];
	        this.clusterName = source["clusterName"];
	    }
	}
	
	export class ProjectStats {
	    widgetsCount: number;
	    flowsCount: number;
	    nodesCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.widgetsCount = source["widgetsCount"];
	        this.flowsCount = source["flowsCount"];
	        this.nodesCount = source["nodesCount"];
	    }
	}
	export class RunExpressionResult {
	    result: string;
	    validSchema: boolean;
	    validationError: string;
	
	    static createFrom(source: any = {}) {
	        return new RunExpressionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result = source["result"];
	        this.validSchema = source["validSchema"];
	        this.validationError = source["validationError"];
	    }
	}
	export class TraceDataResponse {
	    traceId: string;
	    spans: utils.Span[];
	
	    static createFrom(source: any = {}) {
	        return new TraceDataResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traceId = source["traceId"];
	        this.spans = this.convertValues(source["spans"], utils.Span);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TracesResponse {
	    traces: utils.TraceInfo[];
	    total: number;
	    offset: number;
	
	    static createFrom(source: any = {}) {
	        return new TracesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traces = this.convertValues(source["traces"], utils.TraceInfo);
	        this.total = source["total"];
	        this.offset = source["offset"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TransferNodesRequest {
	    fromFlowResourceName: string;
	    toFlowResourceName: string;
	    projectResourceName: string;
	    nodeIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new TransferNodesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fromFlowResourceName = source["fromFlowResourceName"];
	        this.toFlowResourceName = source["toFlowResourceName"];
	        this.projectResourceName = source["projectResourceName"];
	        this.nodeIds = source["nodeIds"];
	    }
	}
	export class Widget {
	    id: string;
	    title: string;
	    nodeName: string;
	    port: string;
	    defaultSchema: Record<string, any>;
	    schema?: Record<string, any>;
	    data: Record<string, any>;
	    gridX: number;
	    gridY: number;
	    gridW: number;
	    gridH: number;
	    pages?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Widget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.nodeName = source["nodeName"];
	        this.port = source["port"];
	        this.defaultSchema = source["defaultSchema"];
	        this.schema = source["schema"];
	        this.data = source["data"];
	        this.gridX = source["gridX"];
	        this.gridY = source["gridY"];
	        this.gridW = source["gridW"];
	        this.gridH = source["gridH"];
	        this.pages = source["pages"];
	    }
	}
	export class WidgetPage {
	    name: string;
	    title: string;
	    resourceName: string;
	    sortIdx: number;
	
	    static createFrom(source: any = {}) {
	        return new WidgetPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
	        this.resourceName = source["resourceName"];
	        this.sortIdx = source["sortIdx"];
	    }
	}

}

export namespace utils {
	
	export class SpanEvent {
	    name: string;
	    attributes: SpanAttribute[];
	
	    static createFrom(source: any = {}) {
	        return new SpanEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.attributes = this.convertValues(source["attributes"], SpanAttribute);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SpanAttribute {
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new SpanAttribute(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class Span {
	    trace_id: string;
	    span_id: string;
	    name: string;
	    start_time_unix_nano: number;
	    end_time_unix_nano: number;
	    attributes: SpanAttribute[];
	    events: SpanEvent[];
	
	    static createFrom(source: any = {}) {
	        return new Span(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trace_id = source["trace_id"];
	        this.span_id = source["span_id"];
	        this.name = source["name"];
	        this.start_time_unix_nano = source["start_time_unix_nano"];
	        this.end_time_unix_nano = source["end_time_unix_nano"];
	        this.attributes = this.convertValues(source["attributes"], SpanAttribute);
	        this.events = this.convertValues(source["events"], SpanEvent);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class TraceInfo {
	    id: string;
	    spans: number;
	    errors: number;
	    data: number;
	    length: number;
	    duration: number;
	    start: number;
	    end: number;
	
	    static createFrom(source: any = {}) {
	        return new TraceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.spans = source["spans"];
	        this.errors = source["errors"];
	        this.data = source["data"];
	        this.length = source["length"];
	        this.duration = source["duration"];
	        this.start = source["start"];
	        this.end = source["end"];
	    }
	}

}

