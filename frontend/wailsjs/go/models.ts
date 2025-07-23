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
	export class CloudflareAPIConfig {
	    token: string;
	    accountName: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareAPIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.accountName = source["accountName"];
	    }
	}
	export class CloudflareAccount {
	    id: string;
	    name: string;
	    created_on: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareAccount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_on = source["created_on"];
	    }
	}
	
	export class CloudflareDNSRecord {
	    id: string;
	    type: string;
	    name: string;
	    content: string;
	    ttl: number;
	    proxied?: boolean;
	    zone_id: string;
	    zone_name: string;
	    created_on?: string;
	    modified_on?: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareDNSRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.name = source["name"];
	        this.content = source["content"];
	        this.ttl = source["ttl"];
	        this.proxied = source["proxied"];
	        this.zone_id = source["zone_id"];
	        this.zone_name = source["zone_name"];
	        this.created_on = source["created_on"];
	        this.modified_on = source["modified_on"];
	    }
	}
	export class CloudflareHost {
	    Name: string;
	    Website: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareHost(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Website = source["Website"];
	    }
	}
	export class CloudflareLoadBalancer {
	    id: string;
	    name: string;
	    description?: string;
	    enabled: boolean;
	    ttl: number;
	    proxied: boolean;
	    default_pools: string[];
	    fallback_pool?: string;
	    session_affinity: string;
	    steering_policy: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareLoadBalancer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	        this.ttl = source["ttl"];
	        this.proxied = source["proxied"];
	        this.default_pools = source["default_pools"];
	        this.fallback_pool = source["fallback_pool"];
	        this.session_affinity = source["session_affinity"];
	        this.steering_policy = source["steering_policy"];
	    }
	}
	export class CloudflareMeta {
	    page_rule_quota: number;
	    wildcard_proxiable: boolean;
	    phishing_detected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page_rule_quota = source["page_rule_quota"];
	        this.wildcard_proxiable = source["wildcard_proxiable"];
	        this.phishing_detected = source["phishing_detected"];
	    }
	}
	export class CloudflareOrigin {
	    name: string;
	    address: string;
	    enabled: boolean;
	    weight?: number;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareOrigin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.address = source["address"];
	        this.enabled = source["enabled"];
	        this.weight = source["weight"];
	    }
	}
	export class CloudflareOwner {
	    id: string;
	    email: string;
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareOwner(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.email = source["email"];
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class CloudflarePlan {
	    id: string;
	    name: string;
	    currency: string;
	    legacy_id: string;
	    is_subscribed: boolean;
	    can_subscribe: boolean;
	    legacy_discount: boolean;
	    externally_managed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CloudflarePlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.currency = source["currency"];
	        this.legacy_id = source["legacy_id"];
	        this.is_subscribed = source["is_subscribed"];
	        this.can_subscribe = source["can_subscribe"];
	        this.legacy_discount = source["legacy_discount"];
	        this.externally_managed = source["externally_managed"];
	    }
	}
	export class CloudflarePool {
	    id: string;
	    name: string;
	    description?: string;
	    enabled: boolean;
	    healthy: boolean;
	    minimum_origins: number;
	    origins: CloudflareOrigin[];
	    monitor?: string;
	    notification_email?: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflarePool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	        this.healthy = source["healthy"];
	        this.minimum_origins = source["minimum_origins"];
	        this.origins = this.convertValues(source["origins"], CloudflareOrigin);
	        this.monitor = source["monitor"];
	        this.notification_email = source["notification_email"];
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
	export class CloudflareWAFRule {
	    id: string;
	    description: string;
	    expression: string;
	    action: string;
	    enabled: boolean;
	    ref?: string;
	    version?: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareWAFRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.description = source["description"];
	        this.expression = source["expression"];
	        this.action = source["action"];
	        this.enabled = source["enabled"];
	        this.ref = source["ref"];
	        this.version = source["version"];
	    }
	}
	export class CloudflareZone {
	    id: string;
	    name: string;
	    development_mode: number;
	    original_name_servers?: string[];
	    original_registrar?: string;
	    original_dnshost?: string;
	    created_on?: string;
	    modified_on?: string;
	    name_servers?: string[];
	    owner?: CloudflareOwner;
	    permissions?: string[];
	    plan?: CloudflarePlan;
	    plan_pending?: CloudflarePlan;
	    status: string;
	    paused: boolean;
	    type: string;
	    host?: CloudflareHost;
	    vanity_name_servers?: string[];
	    betas?: string[];
	    deactivation_reason?: string;
	    meta?: CloudflareMeta;
	    account?: CloudflareAccount;
	    verification_key?: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudflareZone(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.development_mode = source["development_mode"];
	        this.original_name_servers = source["original_name_servers"];
	        this.original_registrar = source["original_registrar"];
	        this.original_dnshost = source["original_dnshost"];
	        this.created_on = source["created_on"];
	        this.modified_on = source["modified_on"];
	        this.name_servers = source["name_servers"];
	        this.owner = this.convertValues(source["owner"], CloudflareOwner);
	        this.permissions = source["permissions"];
	        this.plan = this.convertValues(source["plan"], CloudflarePlan);
	        this.plan_pending = this.convertValues(source["plan_pending"], CloudflarePlan);
	        this.status = source["status"];
	        this.paused = source["paused"];
	        this.type = source["type"];
	        this.host = this.convertValues(source["host"], CloudflareHost);
	        this.vanity_name_servers = source["vanity_name_servers"];
	        this.betas = source["betas"];
	        this.deactivation_reason = source["deactivation_reason"];
	        this.meta = this.convertValues(source["meta"], CloudflareMeta);
	        this.account = this.convertValues(source["account"], CloudflareAccount);
	        this.verification_key = source["verification_key"];
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
	    status: string;
	    created_at: string;
	    message?: string;
	    source: string;
	    url?: string;
	
	    static createFrom(source: any = {}) {
	        return new TFERun(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status = source["status"];
	        this.created_at = source["created_at"];
	        this.message = source["message"];
	        this.source = source["source"];
	        this.url = source["url"];
	    }
	}
	export class TFEVCSConnection {
	    repository: string;
	    branch: string;
	    working_directory: string;
	    webhook_url: string;
	
	    static createFrom(source: any = {}) {
	        return new TFEVCSConnection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repository = source["repository"];
	        this.branch = source["branch"];
	        this.working_directory = source["working_directory"];
	        this.webhook_url = source["webhook_url"];
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
	export class TFEVariable {
	    key: string;
	    value: string;
	    category: string;
	    hcl: boolean;
	    sensitive: boolean;
	    source: string;
	    description: string;
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new TFEVariable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	        this.category = source["category"];
	        this.hcl = source["hcl"];
	        this.sensitive = source["sensitive"];
	        this.source = source["source"];
	        this.description = source["description"];
	        this.id = source["id"];
	    }
	}
	export class TFEVariableSet {
	    id: string;
	    name: string;
	    description?: string;
	    global: boolean;
	    organization: string;
	    workspace_count: number;
	
	    static createFrom(source: any = {}) {
	        return new TFEVariableSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.global = source["global"];
	        this.organization = source["organization"];
	        this.workspace_count = source["workspace_count"];
	    }
	}
	export class TFEVariableSetWorkspace {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new TFEVariableSetWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class TFEVariableSetDetails {
	    id: string;
	    name: string;
	    description?: string;
	    global: boolean;
	    variables: TFEVariable[];
	    workspaces: TFEVariableSetWorkspace[];
	
	    static createFrom(source: any = {}) {
	        return new TFEVariableSetDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.global = source["global"];
	        this.variables = this.convertValues(source["variables"], TFEVariable);
	        this.workspaces = this.convertValues(source["workspaces"], TFEVariableSetWorkspace);
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
	    terraform_version?: string;
	    status: string;
	    lastRun?: string;
	    owner?: string;
	    tag_names?: string[];
	    organization: string;
	    created_at?: string;
	    updated_at?: string;
	    auto_apply: boolean;
	    locked: boolean;
	    working_directory?: string;
	    terraformWorking: boolean;
	    vcsRepo?: TFEVCSRepo;
	    variables?: Record<string, string>;
	    execution_mode?: string;
	    workspace_type?: string;
	    vcs_connection?: TFEVCSConnection;
	
	    static createFrom(source: any = {}) {
	        return new TFEWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.environment = source["environment"];
	        this.terraform_version = source["terraform_version"];
	        this.status = source["status"];
	        this.lastRun = source["lastRun"];
	        this.owner = source["owner"];
	        this.tag_names = source["tag_names"];
	        this.organization = source["organization"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.auto_apply = source["auto_apply"];
	        this.locked = source["locked"];
	        this.working_directory = source["working_directory"];
	        this.terraformWorking = source["terraformWorking"];
	        this.vcsRepo = this.convertValues(source["vcsRepo"], TFEVCSRepo);
	        this.variables = source["variables"];
	        this.execution_mode = source["execution_mode"];
	        this.workspace_type = source["workspace_type"];
	        this.vcs_connection = this.convertValues(source["vcs_connection"], TFEVCSConnection);
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

