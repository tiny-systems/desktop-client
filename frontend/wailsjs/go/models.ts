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

}

