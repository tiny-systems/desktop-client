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
	    resourceName: string;
	    clusterName: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
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

