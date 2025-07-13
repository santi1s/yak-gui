package teleport

type ProfileStatus struct {
	Name           string   `json:"username"`
	Roles          []string `json:"roles"`
	ActiveRequests []string `json:"active_requests"`
	ValidUntil     string   `json:"valid_until"`
}

type ActiveProfileStatus struct {
	Active ProfileStatus `json:"active"`
}

type Resource struct {
	Kind     string           `json:"kind"`
	Version  string           `json:"version"`
	Metadata ResourceMetadata `json:"metadata"`
	Spec     ResourceSpec     `json:"spec"`
	Status   ResourceStatus   `json:"status"`
	Users    Users            `json:"users"`
}

type ResourceMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type ResourceSpec struct {
	URI string      `json:"uri"`
	Aws ResourceAws `json:"aws"`
}

type ResourceAws struct {
	Region          string      `json:"region"`
	Rds             ResourceRds `json:"rds"`
	AccountID       string      `json:"account_id"`
	IamPolicyStatus string      `json:"iam_policy_status"`
}

type ResourceRds struct {
	ClusterID  string `json:"cluster_id"`
	ResourceID string `json:"resource_id"`
	IamAuth    bool   `json:"iam_auth"`
}

type ResourceStatus struct {
	CaCert string      `json:"ca_cert"`
	Aws    ResourceAws `json:"aws"`
}

type Users struct {
	Allowed []string `json:"allowed"`
}

type TeleportDBConnexionConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Database string `json:"database"`
	CA       string `json:"ca"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
}

type ProxyConfig struct {
	EndpointURL     string `json:"endpoint_url"`
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
	AWSCaBundle     string `json:"aws_ca_bundle"`
	TargetProfiles  []string
}

type Config struct {
	Version  string             `yaml:"version"`
	Cluster  string             `yaml:"cluster"`
	Accounts map[string]Account `yaml:"accounts"`
}

// Account holds account-specific configurations
type Account struct {
	Name  string `yaml:"name"`
	Roles []Role `yaml:"roles"`
}

// Role defines the properties of a role within an account
type Role struct {
	Name                     string   `yaml:"name"`
	Type                     string   `yaml:"type"`
	MaxDuration              string   `yaml:"max_duration"`
	AWSConfigProfiles        []string `yaml:"aws_config_profiles"`
	AWSApplicationName       string   `yaml:"application_name"`
	Permission               string   `yaml:"permission"`
	BypassTeleport           bool     `yaml:"bypass_teleport"`
	Status                   string   `yaml:"status"`
	BackupSSORoleName        string   `yaml:"backup_sso_role_name"`
	AccountNo                string   `yaml:"account_no"`
	SlackReviewersUserGroups []string `yaml:"slack_user_groups"`
}

type AccessRequest struct {
	Metadata AccessRequestMetadata `json:"metadata"`
	Spec     AccessRequestSpec     `json:"spec"`
}

type AccessRequestMetadata struct {
	Name    string `json:"name"`
	Expires string `json:"expires"`
}

type AccessRequestSpec struct {
	User   string   `json:"user"`
	Roles  []string `json:"roles"`
	State  int      `json:"state"`
	Reason string   `json:"request_reason"`
}

type AssumeRoleResponse struct {
	Credentials AssumeRoleCredentials `json:"Credentials"`
}

type AssumeRoleCredentials struct {
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

type CredentialProcessResponse struct {
	Version         int    `json:"Version"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}
