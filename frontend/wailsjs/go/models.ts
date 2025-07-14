export namespace main {
	
	export class ArgoApp {
	    AppName: string;
	    Health: string;
	    Sync: string;
	    Suspended: boolean;
	    SyncLoop: string;
	    Conditions: string[];
	
	    static createFrom(source: any = {}) {
	        return new ArgoApp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AppName = source["AppName"];
	        this.Health = source["Health"];
	        this.Sync = source["Sync"];
	        this.Suspended = source["Suspended"];
	        this.SyncLoop = source["SyncLoop"];
	        this.Conditions = source["Conditions"];
	    }
	}
	export class ArgoConfig {
	    server: string;
	    project: string;
	    username?: string;
	    password?: string;
	
	    static createFrom(source: any = {}) {
	        return new ArgoConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.project = source["project"];
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class EnvironmentProfile {
	    name: string;
	    aws_profile: string;
	    kubeconfig: string;
	    path: string;
	    tf_infra_repository_path: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvironmentProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.aws_profile = source["aws_profile"];
	        this.kubeconfig = source["kubeconfig"];
	        this.path = source["path"];
	        this.tf_infra_repository_path = source["tf_infra_repository_path"];
	        this.created_at = source["created_at"];
	    }
	}
	export class JWTClientConfig {
	    platform: string;
	    environment: string;
	    team: string;
	    path: string;
	    owner: string;
	    localName: string;
	    targetService: string;
	    secret: string;
	
	    static createFrom(source: any = {}) {
	        return new JWTClientConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.environment = source["environment"];
	        this.team = source["team"];
	        this.path = source["path"];
	        this.owner = source["owner"];
	        this.localName = source["localName"];
	        this.targetService = source["targetService"];
	        this.secret = source["secret"];
	    }
	}
	export class JWTServerConfig {
	    platform: string;
	    environment: string;
	    team: string;
	    path: string;
	    owner: string;
	    localName: string;
	    serviceName: string;
	    clientName: string;
	    clientSecret: string;
	
	    static createFrom(source: any = {}) {
	        return new JWTServerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.environment = source["environment"];
	        this.team = source["team"];
	        this.path = source["path"];
	        this.owner = source["owner"];
	        this.localName = source["localName"];
	        this.serviceName = source["serviceName"];
	        this.clientName = source["clientName"];
	        this.clientSecret = source["clientSecret"];
	    }
	}
	export class KubernetesConfig {
	    server: string;
	    namespace: string;
	
	    static createFrom(source: any = {}) {
	        return new KubernetesConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.namespace = source["namespace"];
	    }
	}
	export class RolloutListItem {
	    name: string;
	    namespace: string;
	    status: string;
	    replicas: string;
	    age: string;
	    strategy: string;
	    revision: string;
	    images: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new RolloutListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.namespace = source["namespace"];
	        this.status = source["status"];
	        this.replicas = source["replicas"];
	        this.age = source["age"];
	        this.strategy = source["strategy"];
	        this.revision = source["revision"];
	        this.images = source["images"];
	    }
	}
	export class RolloutStatus {
	    name: string;
	    namespace: string;
	    status: string;
	    replicas: string;
	    updated: string;
	    ready: string;
	    available: string;
	    strategy: string;
	    currentStep: string;
	    revision: string;
	    message: string;
	    analysis: string;
	    images: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new RolloutStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.namespace = source["namespace"];
	        this.status = source["status"];
	        this.replicas = source["replicas"];
	        this.updated = source["updated"];
	        this.ready = source["ready"];
	        this.available = source["available"];
	        this.strategy = source["strategy"];
	        this.currentStep = source["currentStep"];
	        this.revision = source["revision"];
	        this.message = source["message"];
	        this.analysis = source["analysis"];
	        this.images = source["images"];
	    }
	}
	export class SecretConfig {
	    platform: string;
	    environment: string;
	    team: string;
	
	    static createFrom(source: any = {}) {
	        return new SecretConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.environment = source["environment"];
	        this.team = source["team"];
	    }
	}
	export class SecretMetadata {
	    owner: string;
	    usage: string;
	    source: string;
	    createdAt: string;
	    updatedAt: string;
	    version: number;
	    destroyed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SecretMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.owner = source["owner"];
	        this.usage = source["usage"];
	        this.source = source["source"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.version = source["version"];
	        this.destroyed = source["destroyed"];
	    }
	}
	export class SecretData {
	    path: string;
	    version: number;
	    data: Record<string, string>;
	    metadata: SecretMetadata;
	
	    static createFrom(source: any = {}) {
	        return new SecretData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.version = source["version"];
	        this.data = source["data"];
	        this.metadata = this.convertValues(source["metadata"], SecretMetadata);
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
	export class SecretListItem {
	    path: string;
	    version: number;
	    owner: string;
	    usage: string;
	    source: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SecretListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.version = source["version"];
	        this.owner = source["owner"];
	        this.usage = source["usage"];
	        this.source = source["source"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

