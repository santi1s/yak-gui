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
	export class ArgoAppDetail {
	    AppName: string;
	    Health: string;
	    Sync: string;
	    Suspended: boolean;
	    SyncLoop: string;
	    Conditions: string[];
	    namespace: string;
	    project: string;
	    repoUrl: string;
	    path: string;
	    targetRev: string;
	    labels: Record<string, string>;
	    annotations: Record<string, string>;
	    createdAt: string;
	    server: string;
	    cluster: string;
	
	    static createFrom(source: any = {}) {
	        return new ArgoAppDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AppName = source["AppName"];
	        this.Health = source["Health"];
	        this.Sync = source["Sync"];
	        this.Suspended = source["Suspended"];
	        this.SyncLoop = source["SyncLoop"];
	        this.Conditions = source["Conditions"];
	        this.namespace = source["namespace"];
	        this.project = source["project"];
	        this.repoUrl = source["repoUrl"];
	        this.path = source["path"];
	        this.targetRev = source["targetRev"];
	        this.labels = source["labels"];
	        this.annotations = source["annotations"];
	        this.createdAt = source["createdAt"];
	        this.server = source["server"];
	        this.cluster = source["cluster"];
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
	export class SecretPath {
	    platform: string;
	    env: string;
	    path: string;
	    keys: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new SecretPath(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.env = source["env"];
	        this.path = source["path"];
	        this.keys = source["keys"];
	    }
	}
	export class CloudflareConfig {
	    path: string;
	    zone: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.zone = source["zone"];
	    }
	}
	export class Certificate {
	    name: string;
	    conf: string;
	    issuer: string;
	    tags: string[];
	    cloudflare: CloudflareConfig;
	    secret: SecretPath;
	
	    static createFrom(source: any = {}) {
	        return new Certificate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.conf = source["conf"];
	        this.issuer = source["issuer"];
	        this.tags = source["tags"];
	        this.cloudflare = this.convertValues(source["cloudflare"], CloudflareConfig);
	        this.secret = this.convertValues(source["secret"], SecretPath);
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
	export class CertificateOperation {
	    success: boolean;
	    message: string;
	    output: string;
	
	    static createFrom(source: any = {}) {
	        return new CertificateOperation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.output = source["output"];
	    }
	}
	
	export class ClusterConfig {
	    Endpoint: string;
	
	    static createFrom(source: any = {}) {
	        return new ClusterConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Endpoint = source["Endpoint"];
	    }
	}
	export class EnvironmentProfile {
	    name: string;
	    aws_profile: string;
	    kubeconfig: string;
	    path: string;
	    tf_infra_repository_path: string;
	    gandi_token: string;
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
	        this.gandi_token = source["gandi_token"];
	        this.created_at = source["created_at"];
	    }
	}
	export class JWTClientConfig {
	    platform: string;
	    environment: string;
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
	export class PlatformConfig {
	    Clusters: string[];
	    AwsProfile: string;
	    AwsRegion: string;
	    VaultRole: string;
	    Environments: Record<string, string>;
	    VaultParentNamespace: string;
	
	    static createFrom(source: any = {}) {
	        return new PlatformConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Clusters = source["Clusters"];
	        this.AwsProfile = source["AwsProfile"];
	        this.AwsRegion = source["AwsRegion"];
	        this.VaultRole = source["VaultRole"];
	        this.Environments = source["Environments"];
	        this.VaultParentNamespace = source["VaultParentNamespace"];
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
	
	    static createFrom(source: any = {}) {
	        return new SecretConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.environment = source["environment"];
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
	
	
	export class TFEConfig {
	    endpoint: string;
	    organization: string;
	    token?: string;
	
	    static createFrom(source: any = {}) {
	        return new TFEConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpoint = source["endpoint"];
	        this.organization = source["organization"];
	        this.token = source["token"];
	    }
	}
	export class TFEPlanExecution {
	    workspaceNames?: string[];
	    owner?: string;
	    terraformVersion: string;
	    message?: string;
	    wait: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TFEPlanExecution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceNames = source["workspaceNames"];
	        this.owner = source["owner"];
	        this.terraformVersion = source["terraformVersion"];
	        this.message = source["message"];
	        this.wait = source["wait"];
	    }
	}
	export class TFEPlanResult {
	    workspaceName: string;
	    runId: string;
	    status: string;
	    hasChanges: boolean;
	    message?: string;
	    error?: string;
	    url?: string;
	    duration?: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new TFEPlanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceName = source["workspaceName"];
	        this.runId = source["runId"];
	        this.status = source["status"];
	        this.hasChanges = source["hasChanges"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.url = source["url"];
	        this.duration = source["duration"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class TFERun {
	    id: string;
	    workspaceId: string;
	    workspaceName: string;
	    status: string;
	    createdAt: string;
	    message?: string;
	    source: string;
	    terraformVersion?: string;
	    hasChanges: boolean;
	    isDestroy: boolean;
	    isConfirmable: boolean;
	    // Go type: struct { IsConfirmable bool "json:\"isConfirmable\""; IsCancelable bool "json:\"isCancelable\""; IsDiscardable bool "json:\"isDiscardable\"" }
	    actions: any;
	    createdBy?: string;
	    url?: string;
	
	    static createFrom(source: any = {}) {
	        return new TFERun(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.workspaceName = source["workspaceName"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.message = source["message"];
	        this.source = source["source"];
	        this.terraformVersion = source["terraformVersion"];
	        this.hasChanges = source["hasChanges"];
	        this.isDestroy = source["isDestroy"];
	        this.isConfirmable = source["isConfirmable"];
	        this.actions = this.convertValues(source["actions"], Object);
	        this.createdBy = source["createdBy"];
	        this.url = source["url"];
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
	export class TFEVCSRepo {
	    identifier: string;
	    branch: string;
	    ingressSubmodules: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TFEVCSRepo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.identifier = source["identifier"];
	        this.branch = source["branch"];
	        this.ingressSubmodules = source["ingressSubmodules"];
	    }
	}
	export class TFEVersionInfo {
	    version: string;
	    status: string;
	    isDefault: boolean;
	    isSupported: boolean;
	    beta: boolean;
	    usage: number;
	
	    static createFrom(source: any = {}) {
	        return new TFEVersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.status = source["status"];
	        this.isDefault = source["isDefault"];
	        this.isSupported = source["isSupported"];
	        this.beta = source["beta"];
	        this.usage = source["usage"];
	    }
	}
	export class TFEWorkspace {
	    id: string;
	    name: string;
	    description?: string;
	    environment?: string;
	    terraformVersion?: string;
	    status: string;
	    lastRun?: string;
	    owner?: string;
	    tags?: string[];
	    organization: string;
	    createdAt?: string;
	    updatedAt?: string;
	    autoApply: boolean;
	    terraformWorking: boolean;
	    vcsRepo?: TFEVCSRepo;
	    variables?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new TFEWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.environment = source["environment"];
	        this.terraformVersion = source["terraformVersion"];
	        this.status = source["status"];
	        this.lastRun = source["lastRun"];
	        this.owner = source["owner"];
	        this.tags = source["tags"];
	        this.organization = source["organization"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.autoApply = source["autoApply"];
	        this.terraformWorking = source["terraformWorking"];
	        this.vcsRepo = this.convertValues(source["vcsRepo"], TFEVCSRepo);
	        this.variables = source["variables"];
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
	export class YakSecretConfig {
	    Clusters: Record<string, ClusterConfig>;
	    Platforms: Record<string, PlatformConfig>;
	
	    static createFrom(source: any = {}) {
	        return new YakSecretConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Clusters = this.convertValues(source["Clusters"], ClusterConfig, true);
	        this.Platforms = this.convertValues(source["Platforms"], PlatformConfig, true);
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

}

