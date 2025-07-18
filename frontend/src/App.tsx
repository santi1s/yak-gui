import React, { useState, useEffect } from 'react';
import { 
  Layout, 
  Tabs, 
  Button, 
  Card, 
  Input, 
  Select, 
  Space, 
  Typography, 
  Alert, 
  Spin, 
  Switch, 
  ConfigProvider, 
  theme,
  Tag,
  Table,
  Modal,
  Form,
  Popconfirm,
  Divider,
  Row,
  Col,
  Descriptions,
  Tooltip
} from 'antd';
import {
  DesktopOutlined,
  CloudOutlined,
  DatabaseOutlined,
  SafetyOutlined,
  SettingOutlined,
  MoonOutlined,
  SunOutlined,
  ImportOutlined,
  SaveOutlined,
  LoadingOutlined,
  DeleteOutlined,
  EditOutlined,
  CopyOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  UserOutlined,
  FolderOutlined,
  ApiOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  FileProtectOutlined,
  UnorderedListOutlined,
  AppstoreOutlined,
  BarsOutlined
} from '@ant-design/icons';
import Rollouts from './Rollouts';
import Secrets from './Secrets';
import Certificates from './Certificates';
import TFE from './TFE';
import ArgoCD from './ArgoCD';
import FeatureFlagManager from './FeatureFlagManager';
import { useFeatureFlags } from './featureFlags';

const { Header, Content } = Layout;
const { Title, Text } = Typography;
const { Option } = Select;

// Types matching the Go backend

interface EnvironmentProfile {
  name: string;
  aws_profile: string;
  kubeconfig: string;
  path: string;
  tf_infra_repository_path: string;
  created_at: string;
}

// Declare global functions for Wails - consolidated interface for all components
declare global {
  interface Window {
    go: {
      main: {
        App: {
          // ArgoCD functions
          GetArgoApps: (config: ArgoConfig) => Promise<ArgoApp[]>;
          SyncArgoApp: (config: ArgoConfig, appName: string, prune: boolean, dryRun: boolean) => Promise<void>;
          RefreshArgoApp: (config: ArgoConfig, appName: string) => Promise<void>;
          SuspendArgoApp: (config: ArgoConfig, appName: string) => Promise<void>;
          UnsuspendArgoApp: (config: ArgoConfig, appName: string) => Promise<void>;
          GetArgoCDServerFromProfile: () => Promise<string>;
          GetArgoAppDetail: (config: ArgoConfig, appName: string) => Promise<ArgoAppDetail>;
          GetCurrentAWSProfile: () => Promise<string>;
          SetAWSProfile: (profile: string) => Promise<void>;
          GetKubeconfig: () => Promise<string>;
          SetKubeconfig: (path: string) => Promise<void>;
          SetPATH: (path: string) => Promise<void>;
          SetTfInfraRepositoryPath: (path: string) => Promise<void>;
          GetAWSProfiles: () => Promise<string[]>;
          GetShellPATH: () => Promise<string>;
          GetShellEnvironment: () => Promise<Record<string, string>>;
          ImportShellEnvironment: () => Promise<void>;
          GetEnvironmentVariables: () => Promise<Record<string, string>>;
          SaveEnvironmentProfile: (name: string) => Promise<void>;
          GetEnvironmentProfiles: () => Promise<EnvironmentProfile[]>;
          LoadEnvironmentProfile: (name: string) => Promise<void>;
          DeleteEnvironmentProfile: (name: string) => Promise<void>;
          GetAppVersion: () => Promise<Record<string, string>>;
          TestSimpleArray: () => Promise<string[]>;
          TestSimpleApps: () => Promise<ArgoApp[]>;
          LoginToArgoCD: (config: ArgoConfig) => Promise<void>;
          // Rollouts functions
          GetRollouts: (config: any) => Promise<any[]>;
          GetRolloutStatus: (config: any, rolloutName: string) => Promise<any>;
          PromoteRollout: (config: any, rolloutName: string, full: boolean) => Promise<void>;
          PauseRollout: (config: any, rolloutName: string) => Promise<void>;
          AbortRollout: (config: any, rolloutName: string) => Promise<void>;
          RestartRollout: (config: any, rolloutName: string) => Promise<void>;
          SetRolloutImage: (config: any, rolloutName: string, image: string, container: string) => Promise<void>;
          // Secret functions
          GetSecrets: (config: any, path: string) => Promise<any[]>;
          GetSecretData: (config: any, path: string, version: number) => Promise<any>;
          CreateSecret: (config: any, path: string, owner: string, usage: string, source: string, data: Record<string, string>) => Promise<void>;
          UpdateSecret: (config: any, path: string, data: Record<string, string>) => Promise<void>;
          DeleteSecret: (config: any, path: string, version: number) => Promise<void>;
          // Secret config functions
          GetSecretConfigPlatforms: () => Promise<string[]>;
          GetSecretConfigEnvironments: (platform: string) => Promise<string[]>;
          GetSecretConfigPaths: (platform: string, environment: string) => Promise<string[]>;
          // JWT functions
          CreateJWTClient: (config: any) => Promise<void>;
          CreateJWTServer: (config: any) => Promise<void>;
          // Certificate functions
          CheckGandiToken: () => Promise<any>;
          ListCertificates: () => Promise<string[]>;
          RenewCertificate: (certificateName: string, jiraTicket: string) => Promise<any>;
          RefreshCertificateSecret: (certificateName: string, jiraTicket: string) => Promise<any>;
          DescribeCertificateSecret: (certificateName: string, version: number, diffVersion: number) => Promise<any>;
          SendCertificateNotification: (certificateName: string, operationDate: string, operation: string) => Promise<any>;
          // TFE functions
          GetTFEConfig: () => Promise<any>;
          SetTFEConfig: (config: any) => Promise<void>;
          GetTFEWorkspaces: (config: any) => Promise<any[]>;
          GetTFEWorkspacesByTag: (config: any, tag: string, not: boolean) => Promise<any[]>;
          ExecuteTFEPlan: (config: any, execution: any) => Promise<any[]>;
          GetTFERuns: (config: any, workspaceID: string) => Promise<any[]>;
          GetTFERunLogs: (config: any, runID: string) => Promise<string>;
          LockTFEWorkspace: (config: any, workspaceNames: string[], checkStatus: boolean) => Promise<void>;
          UnlockTFEWorkspace: (config: any, workspaceNames: string[], force: boolean) => Promise<void>;
          SetTFEWorkspaceVersion: (config: any, workspaceNames: string[], version: string) => Promise<void>;
          DiscardTFERuns: (config: any, ageHours: number, discardPending: boolean, dryRun: boolean, allWorkspaces: boolean) => Promise<void>;
          GetTFEVersions: (config: any) => Promise<any[]>;
          CheckTFEDeprecatedVersions: (config: any, versionFile: string, teamsFile: string, sendEmail: boolean) => Promise<any>;
          // TFE Variable functions
          GetTFEWorkspaceVariables: (config: any, workspaceId: string, includeSets: boolean) => Promise<any[]>;
          GetTFEVariableSetVariables: (config: any, variableSetName: string) => Promise<any[]>;
          GetTFEVariableSets: (config: any) => Promise<any[]>;
          GetTFEWorkspaceDetails: (config: any, workspaceName: string) => Promise<any>;
          GetTFEVariableSetDetails: (config: any, variableSetName: string) => Promise<any>;
          // Window control functions
          MaximizeWindow: () => void;
          UnmaximizeWindow: () => void;
          IsWindowMaximized: () => boolean;
        };
      };
    };
  }
}





// Environment Configuration Component
const EnvironmentConfig: React.FC<{ 
  onAWSProfileChange?: () => void;
  onShellLoadingChange?: (isLoading: boolean) => void;
}> = ({ onAWSProfileChange, onShellLoadingChange }) => {
  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [awsProfile, setAwsProfile] = useState('');
  const [awsProfiles, setAwsProfiles] = useState<string[]>([]);
  const [kubeconfig, setKubeconfig] = useState('');
  const [pathVar, setPathVar] = useState('');
  const [tfInfraPath, setTfInfraPath] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [autoImported, setAutoImported] = useState(false);
  const [isAutoImporting, setIsAutoImporting] = useState(true);
  
  // Profile management state
  const [profiles, setProfiles] = useState<EnvironmentProfile[]>([]);
  const [newProfileName, setNewProfileName] = useState('');
  const [selectedProfile, setSelectedProfile] = useState('');
  
  // Version info state
  const [versionInfo, setVersionInfo] = useState<Record<string, string>>({});

  const loadEnvironmentVariables = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const vars = await window.go.main.App.GetEnvironmentVariables();
        setEnvVars(vars);
        setAwsProfile(vars.AWS_PROFILE || '');
        setKubeconfig(vars.KUBECONFIG || '');
        setPathVar(vars.PATH || '');
        setTfInfraPath(vars.TFINFRA_REPOSITORY_PATH || '');
        
        if (vars.TFINFRA_REPOSITORY_PATH && vars.AWS_PROFILE && !vars.KUBECONFIG) {
          const expectedKubeconfig = `${vars.TFINFRA_REPOSITORY_PATH}/setup/k8senv/${vars.AWS_PROFILE}/config`;
          setKubeconfig(expectedKubeconfig);
        }
        
        try {
          const profiles = await window.go.main.App.GetAWSProfiles();
          setAwsProfiles(profiles);
        } catch (error) {
          console.warn('Failed to load AWS profiles:', error);
        }
        
        if (!vars.PATH || vars.PATH.split(':').length < 4) {
          try {
            const shellPath = await window.go.main.App.GetShellPATH();
            if (shellPath && shellPath !== vars.PATH) {
              setPathVar(shellPath);
            }
          } catch (error) {
            console.warn('Failed to detect shell PATH:', error);
          }
        }
      }
    } catch (error) {
      console.error('Failed to load environment variables:', error);
      setError('Failed to load environment variables');
    }
  };

  const loadProfiles = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const envProfiles = await window.go.main.App.GetEnvironmentProfiles();
        setProfiles(envProfiles);
      }
    } catch (error) {
      console.error('Failed to load profiles:', error);
    }
  };

  const loadVersionInfo = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const version = await window.go.main.App.GetAppVersion();
        setVersionInfo(version);
      }
    } catch (error) {
      console.error('Failed to load version info:', error);
    }
  };

  useEffect(() => {
    const initializeEnvironment = async () => {
      setIsAutoImporting(true);
      if (onShellLoadingChange) {
        onShellLoadingChange(true);
      }
      
      // First, auto-import shell environment
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.ImportShellEnvironment();
          console.log('Shell environment auto-imported at startup');
          setAutoImported(true);
          setSuccess('Shell environment was automatically imported at startup');
        }
      } catch (error) {
        console.warn('Failed to auto-import shell environment:', error);
        setError('Failed to auto-import shell environment. You may need to manually import it.');
      } finally {
        setIsAutoImporting(false);
        if (onShellLoadingChange) {
          onShellLoadingChange(false);
        }
      }
      
      // Then load everything else
      await loadEnvironmentVariables();
      await loadProfiles();
      await loadVersionInfo();
      
      // After everything is loaded, trigger ArgoCD server update
      // Use a small delay to ensure all environment variables are properly set
      setTimeout(() => {
        if (onAWSProfileChange) {
          onAWSProfileChange();
        }
      }, 100);
    };
    
    initializeEnvironment();
  }, [onShellLoadingChange]);

  const handleSetAWSProfile = async () => {
    if (!awsProfile.trim()) {
      setError('AWS Profile cannot be empty');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await window.go.main.App.SetAWSProfile(awsProfile.trim());
      setSuccess('AWS Profile set successfully');
      await loadEnvironmentVariables();
      if (onAWSProfileChange) {
        onAWSProfileChange();
      }
    } catch (error) {
      setError(`Failed to set AWS Profile: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const handleImportShellEnvironment = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await window.go.main.App.ImportShellEnvironment();
      setSuccess('Shell environment imported successfully');
      await loadEnvironmentVariables();
      if (onAWSProfileChange) {
        onAWSProfileChange();
      }
    } catch (error) {
      setError(`Failed to import shell environment: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveProfile = async () => {
    if (!newProfileName.trim()) {
      setError('Profile name cannot be empty');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await window.go.main.App.SaveEnvironmentProfile(newProfileName.trim());
      setSuccess(`Profile '${newProfileName.trim()}' saved successfully`);
      setNewProfileName('');
      await loadProfiles();
    } catch (error) {
      setError(`Failed to save profile: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const handleLoadProfile = async () => {
    if (!selectedProfile) {
      setError('Please select a profile to load');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await window.go.main.App.LoadEnvironmentProfile(selectedProfile);
      setSuccess(`Profile '${selectedProfile}' loaded successfully`);
      await loadEnvironmentVariables();
      if (onAWSProfileChange) {
        onAWSProfileChange();
      }
    } catch (error) {
      setError(`Failed to load profile: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteProfile = async () => {
    if (!selectedProfile) {
      setError('Please select a profile to delete');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await window.go.main.App.DeleteEnvironmentProfile(selectedProfile);
      setSuccess(`Profile '${selectedProfile}' deleted successfully`);
      setSelectedProfile('');
      await loadProfiles();
    } catch (error) {
      setError(`Failed to delete profile: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      <Row justify="space-between" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2}>
            <DesktopOutlined style={{ marginRight: '8px' }} />
            Environment Configuration
          </Title>
          <Text type="secondary">Set environment variables for AWS and Kubernetes</Text>
        </Col>
        <Col>
          {versionInfo.version && (
            <div style={{ textAlign: 'right' }}>
              <Text strong>{versionInfo.name || 'Yak GUI'}</Text>
              <br />
              <Text type="secondary">v{versionInfo.version}</Text>
            </div>
          )}
        </Col>
      </Row>

      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: '16px' }}
        />
      )}

      {success && (
        <Alert
          message={success}
          type="success"
          showIcon
          closable
          onClose={() => setSuccess(null)}
          style={{ marginBottom: '16px' }}
        />
      )}

      <Space direction="vertical" style={{ width: '100%' }} size="large">
        {/* Import Shell Environment */}
        <Card 
          title={
            <Space>
              <ImportOutlined />
              Shell Environment
              {isAutoImporting && <Spin size="small" />}
            </Space>
          }
          style={{ 
            border: isAutoImporting 
              ? '2px solid #1890ff' 
              : autoImported 
                ? '2px solid #52c41a' 
                : '2px solid #1890ff' 
          }}
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            {isAutoImporting ? (
              <div style={{ textAlign: 'center', padding: '20px 0' }}>
                <Spin size="large" />
                <div style={{ marginTop: '16px' }}>
                  <Text>Importing shell environment at startup...</Text>
                </div>
              </div>
            ) : autoImported ? (
              <Text>
                <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
                Shell environment was automatically imported at startup. 
                If you need to re-import (after changing your shell configuration), click the button below.
              </Text>
            ) : (
              <Text>
                <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: '8px' }} />
                If you launched this app from Finder and don't see your AWS profiles or environment variables, 
                click this button to import your shell environment (PATH, AWS_PROFILE, TFINFRA_REPOSITORY_PATH, etc.).
              </Text>
            )}
            {!isAutoImporting && (
              <Button
                type={autoImported ? "default" : "primary"}
                icon={<ImportOutlined />}
                onClick={handleImportShellEnvironment}
                loading={loading}
                size="large"
                style={{ width: '100%' }}
              >
                {loading ? 'Importing...' : autoImported ? 'Re-import Shell Environment' : 'Import Shell Environment'}
              </Button>
            )}
          </Space>
        </Card>

        {/* Profile Management */}
        <Card 
          title={
            <Space>
              <SaveOutlined />
              Environment Profiles
            </Space>
          }
          style={{ border: '2px solid #1890ff' }}
        >
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Text>Save and load different environment configurations for quick switching between setups.</Text>
            
            {/* Save Profile */}
            <Card type="inner" title="Save Current Configuration" size="small">
              <Space style={{ width: '100%' }}>
                <Input
                  placeholder="Profile name (e.g., 'staging', 'production')"
                  value={newProfileName}
                  onChange={(e) => setNewProfileName(e.target.value)}
                  style={{ flex: 1 }}
                  autoCorrect="off"
                  autoCapitalize="off"
                  spellCheck={false}
                />
                <Button
                  type="primary"
                  onClick={handleSaveProfile}
                  disabled={loading || !newProfileName.trim()}
                  loading={loading}
                >
                  Save Profile
                </Button>
              </Space>
            </Card>

            {/* Load Profile */}
            {profiles.length > 0 && (
              <Card type="inner" title="Load Saved Profile" size="small">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Select
                    placeholder="Select a profile..."
                    value={selectedProfile}
                    onChange={setSelectedProfile}
                    style={{ width: '100%' }}
                  >
                    {profiles.map((profile) => (
                      <Option key={profile.name} value={profile.name}>
                        {profile.name} (AWS: {profile.aws_profile || 'none'})
                      </Option>
                    ))}
                  </Select>
                  <Space style={{ width: '100%' }}>
                    <Button
                      type="primary"
                      onClick={handleLoadProfile}
                      disabled={loading || !selectedProfile}
                      loading={loading}
                      style={{ flex: 1 }}
                    >
                      Load Profile
                    </Button>
                    <Popconfirm
                      title="Are you sure you want to delete this profile?"
                      onConfirm={handleDeleteProfile}
                      disabled={loading || !selectedProfile}
                    >
                      <Button
                        danger
                        disabled={loading || !selectedProfile}
                        loading={loading}
                      >
                        Delete
                      </Button>
                    </Popconfirm>
                  </Space>
                </Space>
              </Card>
            )}

            {profiles.length === 0 && (
              <Card type="inner">
                <Text type="secondary">No saved profiles yet. Save your current configuration above to get started.</Text>
              </Card>
            )}
          </Space>
        </Card>

        {/* AWS Profile Configuration */}
        <Card 
          title={
            <Space>
              <CloudOutlined />
              AWS Profile Configuration
            </Space>
          }
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text>
              <Text strong>Current AWS Profile:</Text> {envVars.AWS_PROFILE || 'Not set'}
            </Text>
            <Space style={{ width: '100%' }}>
              <Select
                placeholder="Select AWS Profile..."
                value={awsProfile}
                onChange={setAwsProfile}
                style={{ flex: 1 }}
              >
                {awsProfiles.map((profile) => (
                  <Option key={profile} value={profile}>
                    {profile}
                  </Option>
                ))}
              </Select>
              <Button
                type="primary"
                onClick={handleSetAWSProfile}
                disabled={loading || !awsProfile}
                loading={loading}
              >
                Set Profile
              </Button>
            </Space>
            <Space direction="vertical" size="small">
              {awsProfiles.length === 0 && (
                <Text type="secondary">💡 No AWS profiles found in ~/.aws/config</Text>
              )}
              {envVars.TFINFRA_REPOSITORY_PATH && (
                <Text type="success">✨ Auto-configures KUBECONFIG and kubectl context when profile is selected</Text>
              )}
              {!envVars.TFINFRA_REPOSITORY_PATH && (
                <Text type="warning">⚠️ TFINFRA_REPOSITORY_PATH not set - KUBECONFIG won't be auto-configured</Text>
              )}
              {awsProfile && (
                <Text type="success">📡 ArgoCD Server: argocd-{awsProfile}.doctolib.net</Text>
              )}
            </Space>
          </Space>
        </Card>

        {/* Environment Variables Display */}
        <Card 
          title={
            <Space>
              <SettingOutlined />
              Current Environment
            </Space>
          }
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            {Object.entries(envVars).map(([key, value]) => (
              <Row key={key} justify="space-between" style={{ marginBottom: '8px' }}>
                <Col flex="0 0 auto" style={{ minWidth: '220px', paddingRight: '16px' }}>
                  <Text strong>{key}:</Text>
                </Col>
                <Col flex="auto">
                  <Text code copyable style={{ wordBreak: 'break-all' }}>
                    {value || 'Not set'}
                  </Text>
                </Col>
              </Row>
            ))}
          </Space>
        </Card>
      </Space>
    </div>
  );
};

// Main App Component
const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'environment' | 'argocd' | 'rollouts' | 'secrets' | 'certificates' | 'tfe'>('environment');
  const [isDarkMode, setIsDarkMode] = useState(false);
  const [isShellLoading, setIsShellLoading] = useState(true);
  const [featureFlags] = useFeatureFlags();

  const [profileChangeCounter, setProfileChangeCounter] = useState(0);
  
  const updateArgoCDServer = async () => {
    // Trigger a re-render in the ArgoCD component to update server configuration
    setProfileChangeCounter(prev => prev + 1);
  };



  // Build tab items based on feature flags
  const buildTabItems = () => {
    const items = [];
    
    if (featureFlags.showEnvironmentTab) {
      items.push({
        key: 'environment',
        label: (
          <span>
            <DesktopOutlined />
            Environment
          </span>
        ),
        children: (
          <div>
            <EnvironmentConfig onAWSProfileChange={updateArgoCDServer} onShellLoadingChange={setIsShellLoading} />
            <div style={{ marginTop: '24px' }}>
              <FeatureFlagManager />
            </div>
          </div>
        ),
      });
    }
    
    if (featureFlags.showArgoCDTab) {
      items.push({
        key: 'argocd',
        label: (
          <span>
            <CloudOutlined />
            ArgoCD Applications
            {isShellLoading && <LoadingOutlined style={{ marginLeft: '8px' }} />}
          </span>
        ),
        children: <ArgoCD profileChangeCounter={profileChangeCounter} />,
        disabled: isShellLoading,
      });
    }
    
    if (featureFlags.showRolloutsTab) {
      items.push({
        key: 'rollouts',
        label: (
          <span>
            <DatabaseOutlined />
            Argo Rollouts
            {isShellLoading && <LoadingOutlined style={{ marginLeft: '8px' }} />}
          </span>
        ),
        children: <Rollouts />,
        disabled: isShellLoading,
      });
    }
    
    if (featureFlags.showSecretsTab) {
      items.push({
        key: 'secrets',
        label: (
          <span>
            <SafetyOutlined />
            Secrets
          </span>
        ),
        children: <Secrets />,
      });
    }
    
    if (featureFlags.showCertificatesTab) {
      items.push({
        key: 'certificates',
        label: (
          <span>
            <FileProtectOutlined />
            Certificates
          </span>
        ),
        children: <Certificates />,
      });
    }
    
    if (featureFlags.showTFETab) {
      items.push({
        key: 'tfe',
        label: (
          <span>
            <CloudServerOutlined />
            TFE
            <Tag color="orange" size="small" style={{ marginLeft: '8px' }}>BETA</Tag>
          </span>
        ),
        children: <TFE />,
      });
    }
    
    return items;
  };

  const tabItems = buildTabItems();

  return (
    <ConfigProvider
      theme={{
        algorithm: isDarkMode ? theme.darkAlgorithm : theme.defaultAlgorithm,
      }}
    >
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Title level={3} style={{ color: 'white', margin: 0 }}>
            Yak GUI
          </Title>
          <Space>
            <Switch
              checked={isDarkMode}
              onChange={setIsDarkMode}
              checkedChildren={<MoonOutlined />}
              unCheckedChildren={<SunOutlined />}
            />
          </Space>
        </Header>
        <Content>
          <Tabs
            activeKey={activeTab}
            onChange={(key) => setActiveTab(key as any)}
            items={tabItems}
            style={{ height: '100%' }}
          />
        </Content>
      </Layout>
    </ConfigProvider>
  );
};

export default App;