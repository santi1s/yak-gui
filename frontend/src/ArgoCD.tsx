import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Button, 
  Space, 
  Typography, 
  Alert, 
  Spin, 
  Switch, 
  Row,
  Col,
  Descriptions,
  Tooltip,
  Modal,
  Table,
  Tag,
  Input,
  Divider
} from 'antd';
import {
  ApiOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  SettingOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  EyeOutlined,
  CopyOutlined,
  AppstoreOutlined,
  BarsOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;

// Types matching the Go backend
interface ArgoApp {
  AppName: string;
  Health: string;
  Sync: string;
  Suspended: boolean;
  SyncLoop: string;
  Conditions: string[];
}

interface ArgoConfig {
  server: string;
  project: string;
  username?: string;
  password?: string;
}

interface ArgoAppDetail {
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
}

// Status Badge Component
const StatusBadge: React.FC<{ status: string, type: 'health' | 'sync' | 'syncLoop' }> = ({ status, type }) => {
  const getStatusColor = () => {
    if (type === 'health') {
      switch (status.toLowerCase()) {
        case 'healthy': return 'success';
        case 'progressing': return 'processing';
        case 'degraded': return 'error';
        case 'suspended': return 'warning';
        case 'missing': return 'default';
        default: return 'default';
      }
    } else if (type === 'sync') {
      switch (status.toLowerCase()) {
        case 'synced': return 'success';
        case 'outofsync': return 'error';
        default: return 'default';
      }
    } else if (type === 'syncLoop') {
      switch (status.toLowerCase()) {
        case 'enabled': return 'success';
        case 'disabled': return 'default';
        default: return 'default';
      }
    }
    return 'default';
  };

  return <Tag color={getStatusColor()}>{status}</Tag>;
};

// ArgoCD App Detail Modal Component
const ArgoAppDetailModal: React.FC<{
  app: ArgoApp;
  config: ArgoConfig;
  visible: boolean;
  onClose: () => void;
}> = ({ app, config, visible, onClose }) => {
  const [detailedApp, setDetailedApp] = useState<ArgoAppDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadDetailedApp = async () => {
    if (!visible || !app?.AppName) return;
    
    setLoading(true);
    setError(null);
    try {
      const detail = await window.go.main.App.GetArgoAppDetail(config, app.AppName);
      setDetailedApp(detail);
    } catch (error) {
      setError(`Failed to load app details: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible) {
      loadDetailedApp();
    }
  }, [visible, app?.AppName]);

  const handleCopyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      // You can add a message here if needed
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
    }
  };

  return (
    <Modal
      title={
        <Space>
          <ApiOutlined />
          <span>App Details: {app.AppName}</span>
        </Space>
      }
      open={visible}
      onCancel={onClose}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>
      ]}
      width={800}
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Spin size="large" />
          <div style={{ marginTop: '16px' }}>Loading app details...</div>
        </div>
      ) : error ? (
        <Alert
          message="Error"
          description={error}
          type="error"
          showIcon
          action={
            <Button size="small" onClick={loadDetailedApp}>
              Retry
            </Button>
          }
        />
      ) : detailedApp ? (
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="Health">
              <StatusBadge status={detailedApp.Health} type="health" />
            </Descriptions.Item>
            <Descriptions.Item label="Sync">
              <StatusBadge status={detailedApp.Sync} type="sync" />
            </Descriptions.Item>
            <Descriptions.Item label="Namespace">
              <Text code>{detailedApp.namespace}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Project">
              <Text code>{detailedApp.project}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Server">
              <Text code>{detailedApp.server}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Cluster">
              <Text code>{detailedApp.cluster}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Suspended">
              <Tag color={detailedApp.Suspended ? 'red' : 'green'}>
                {detailedApp.Suspended ? 'Yes' : 'No'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Sync Loop">
              <StatusBadge status={detailedApp.SyncLoop} type="syncLoop" />
            </Descriptions.Item>
            <Descriptions.Item label="Created At">
              <Text>{detailedApp.createdAt || 'N/A'}</Text>
            </Descriptions.Item>
          </Descriptions>

          <Card type="inner" title="Git Repository" size="small">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Row justify="space-between" align="middle">
                <Col>
                  <Text strong>Repository URL:</Text>
                </Col>
                <Col flex="auto" style={{ marginLeft: '8px' }}>
                  <Text code style={{ wordBreak: 'break-all', fontSize: '12px' }}>
                    {detailedApp.repoUrl}
                  </Text>
                </Col>
                <Col>
                  <Button
                    type="text"
                    icon={<CopyOutlined />}
                    size="small"
                    onClick={() => handleCopyToClipboard(detailedApp.repoUrl)}
                  />
                </Col>
              </Row>
              <Row justify="space-between">
                <Col>
                  <Text strong>Path:</Text>
                </Col>
                <Col>
                  <Text code>{detailedApp.path}</Text>
                </Col>
              </Row>
              <Row justify="space-between">
                <Col>
                  <Text strong>Target Revision:</Text>
                </Col>
                <Col>
                  <Text code>{detailedApp.targetRev}</Text>
                </Col>
              </Row>
            </Space>
          </Card>

          {detailedApp.Conditions && detailedApp.Conditions.length > 0 && (
            <Card type="inner" title="Conditions" size="small">
              <Space wrap>
                {detailedApp.Conditions.map((condition, index) => (
                  <Tag key={index} color="blue">
                    {condition}
                  </Tag>
                ))}
              </Space>
            </Card>
          )}

          {detailedApp.labels && Object.keys(detailedApp.labels).length > 0 && (
            <Card type="inner" title="Labels" size="small">
              <Space direction="vertical" style={{ width: '100%' }}>
                {Object.entries(detailedApp.labels).map(([key, value]) => (
                  <Row key={key} justify="space-between">
                    <Col>
                      <Text strong>{key}:</Text>
                    </Col>
                    <Col>
                      <Text code>{value}</Text>
                    </Col>
                  </Row>
                ))}
              </Space>
            </Card>
          )}

          {detailedApp.annotations && Object.keys(detailedApp.annotations).length > 0 && (
            <Card type="inner" title="Annotations" size="small">
              <Space direction="vertical" style={{ width: '100%' }}>
                {Object.entries(detailedApp.annotations).map(([key, value]) => (
                  <Row key={key} justify="space-between" style={{ marginBottom: 8 }}>
                    <Col>
                      <Text strong>{key}:</Text>
                    </Col>
                    <Col flex="auto" style={{ marginLeft: '8px' }}>
                      <Text code style={{ wordBreak: 'break-all', fontSize: '12px' }}>
                        {value}
                      </Text>
                    </Col>
                    <Col>
                      <Button
                        type="text"
                        icon={<CopyOutlined />}
                        size="small"
                        onClick={() => handleCopyToClipboard(value)}
                      />
                    </Col>
                  </Row>
                ))}
              </Space>
            </Card>
          )}
        </Space>
      ) : null}
    </Modal>
  );
};

// App List Item Component (for table view)
const AppListItem: React.FC<{
  app: ArgoApp;
  config: ArgoConfig;
  onAction: () => void;
  onOperationFeedback: (message: string, type: 'success' | 'error') => void;
}> = ({ app, config, onAction, onOperationFeedback }) => {
  const [loading, setLoading] = useState(false);
  const [showDetails, setShowDetails] = useState(false);

  const handleSync = async (prune: boolean = false, dryRun: boolean = false) => {
    setLoading(true);
    try {
      await window.go.main.App.SyncArgoApp(config, app.AppName, prune, dryRun);
      const operation = dryRun ? 'Dry run' : prune ? 'Sync + Prune' : 'Sync';
      onOperationFeedback(`${operation} completed successfully for ${app.AppName}`, 'success');
      onAction();
    } catch (error) {
      const operation = dryRun ? 'Dry run' : prune ? 'Sync + Prune' : 'Sync';
      onOperationFeedback(`${operation} failed for ${app.AppName}: ${error}`, 'error');
      console.error('Sync failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setLoading(true);
    try {
      await window.go.main.App.RefreshArgoApp(config, app.AppName);
      onOperationFeedback(`Refresh completed successfully for ${app.AppName}`, 'success');
      onAction();
    } catch (error) {
      onOperationFeedback(`Refresh failed for ${app.AppName}: ${error}`, 'error');
      console.error('Refresh failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSuspend = async () => {
    setLoading(true);
    try {
      if (app.Suspended) {
        await window.go.main.App.UnsuspendArgoApp(config, app.AppName);
        onOperationFeedback(`${app.AppName} resumed successfully`, 'success');
      } else {
        await window.go.main.App.SuspendArgoApp(config, app.AppName);
        onOperationFeedback(`${app.AppName} suspended successfully`, 'success');
      }
      onAction();
    } catch (error) {
      const operation = app.Suspended ? 'Resume' : 'Suspend';
      onOperationFeedback(`${operation} failed for ${app.AppName}: ${error}`, 'error');
      console.error('Suspend/Unsuspend failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Space>
        <Tooltip title="View details">
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => setShowDetails(true)}
            size="small"
          />
        </Tooltip>
        <Tooltip title="Refresh application">
          <Button
            type="text"
            icon={<ReloadOutlined />}
            onClick={handleRefresh}
            loading={loading}
            size="small"
          />
        </Tooltip>
        <Tooltip title={app.Suspended ? "Resume application" : "Suspend application"}>
          <Button
            type="text"
            icon={app.Suspended ? <PlayCircleOutlined /> : <PauseCircleOutlined />}
            onClick={handleSuspend}
            loading={loading}
            size="small"
          />
        </Tooltip>
        <Tooltip title="Synchronize application resources">
          <Button
            type="primary"
            size="small"
            onClick={() => handleSync(false, false)}
            loading={loading}
          >
            Sync
          </Button>
        </Tooltip>
        <Tooltip title="Synchronize and remove resources not in Git">
          <Button
            size="small"
            onClick={() => handleSync(true, false)}
            loading={loading}
          >
            Sync + Prune
          </Button>
        </Tooltip>
        <Tooltip title="Preview synchronization without applying changes">
          <Button
            size="small"
            onClick={() => handleSync(false, true)}
            loading={loading}
          >
            Dry Run
          </Button>
        </Tooltip>
      </Space>
      <ArgoAppDetailModal
        app={app}
        config={config}
        visible={showDetails}
        onClose={() => setShowDetails(false)}
      />
    </>
  );
};

// App Card Component
const AppCard: React.FC<{
  app: ArgoApp;
  config: ArgoConfig;
  onAction: () => void;
  onOperationFeedback: (message: string, type: 'success' | 'error') => void;
}> = ({ app, config, onAction, onOperationFeedback }) => {
  const [loading, setLoading] = useState(false);
  const [showDetails, setShowDetails] = useState(false);

  const handleSync = async (prune: boolean = false, dryRun: boolean = false) => {
    setLoading(true);
    try {
      await window.go.main.App.SyncArgoApp(config, app.AppName, prune, dryRun);
      const operation = dryRun ? 'Dry run' : prune ? 'Sync + Prune' : 'Sync';
      onOperationFeedback(`${operation} completed successfully for ${app.AppName}`, 'success');
      onAction();
    } catch (error) {
      const operation = dryRun ? 'Dry run' : prune ? 'Sync + Prune' : 'Sync';
      onOperationFeedback(`${operation} failed for ${app.AppName}: ${error}`, 'error');
      console.error('Sync failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setLoading(true);
    try {
      await window.go.main.App.RefreshArgoApp(config, app.AppName);
      onOperationFeedback(`Refresh completed successfully for ${app.AppName}`, 'success');
      onAction();
    } catch (error) {
      onOperationFeedback(`Refresh failed for ${app.AppName}: ${error}`, 'error');
      console.error('Refresh failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSuspend = async () => {
    setLoading(true);
    try {
      if (app.Suspended) {
        await window.go.main.App.UnsuspendArgoApp(config, app.AppName);
        onOperationFeedback(`${app.AppName} resumed successfully`, 'success');
      } else {
        await window.go.main.App.SuspendArgoApp(config, app.AppName);
        onOperationFeedback(`${app.AppName} suspended successfully`, 'success');
      }
      onAction();
    } catch (error) {
      const operation = app.Suspended ? 'Resume' : 'Suspend';
      onOperationFeedback(`${operation} failed for ${app.AppName}: ${error}`, 'error');
      console.error('Suspend/Unsuspend failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card
      title={
        <Space>
          <ApiOutlined />
          {app.AppName}
        </Space>
      }
      extra={
        <Space>
          <Tooltip title="View details">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => setShowDetails(true)}
              size="small"
            />
          </Tooltip>
          <Tooltip title="Refresh application">
            <Button
              type="text"
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
              size="small"
            />
          </Tooltip>
          <Tooltip title={app.Suspended ? "Resume application" : "Suspend application"}>
            <Button
              type="text"
              icon={app.Suspended ? <PlayCircleOutlined /> : <PauseCircleOutlined />}
              onClick={handleSuspend}
              loading={loading}
              size="small"
            />
          </Tooltip>
        </Space>
      }
      size="small"
      style={{ marginBottom: 16 }}
    >
      <Space direction="vertical" style={{ width: '100%' }}>
        <Row justify="space-between">
          <Col>
            <Text type="secondary">Health:</Text>
          </Col>
          <Col>
            <StatusBadge status={app.Health} type="health" />
          </Col>
        </Row>
        
        <Row justify="space-between">
          <Col>
            <Text type="secondary">Sync:</Text>
          </Col>
          <Col>
            <StatusBadge status={app.Sync} type="sync" />
          </Col>
        </Row>
        
        <Row justify="space-between">
          <Col>
            <Text type="secondary">Sync Loop:</Text>
          </Col>
          <Col>
            <StatusBadge status={app.SyncLoop} type="syncLoop" />
          </Col>
        </Row>
        
        {app.Conditions && app.Conditions.length > 0 && (
          <div>
            <Text type="secondary">Conditions:</Text>
            <div style={{ marginTop: 4 }}>
              {app.Conditions.map((condition, index) => (
                <Tag key={index} size="small" style={{ marginBottom: 2 }}>
                  {condition}
                </Tag>
              ))}
            </div>
          </div>
        )}
        
        <Space style={{ marginTop: 8 }}>
          <Tooltip title="Synchronize application resources">
            <Button
              type="primary"
              size="small"
              onClick={() => handleSync(false, false)}
              loading={loading}
            >
              Sync
            </Button>
          </Tooltip>
          <Tooltip title="Synchronize and remove resources not in Git">
            <Button
              size="small"
              onClick={() => handleSync(true, false)}
              loading={loading}
            >
              Sync + Prune
            </Button>
          </Tooltip>
          <Tooltip title="Preview synchronization without applying changes">
            <Button
              size="small"
              onClick={() => handleSync(false, true)}
              loading={loading}
            >
              Dry Run
            </Button>
          </Tooltip>
        </Space>
      </Space>

      <ArgoAppDetailModal
        app={app}
        config={config}
        visible={showDetails}
        onClose={() => setShowDetails(false)}
      />
    </Card>
  );
};

// Main ArgoCD Component
const ArgoCD: React.FC<{ profileChangeCounter?: number }> = ({ profileChangeCounter = 0 }) => {
  const [apps, setApps] = useState<ArgoApp[]>([]);
  const [loading, setLoading] = useState(false);
  const [config, setConfig] = useState<ArgoConfig>({
    server: '',
    project: 'main',
    username: '',
    password: ''
  });
  const [error, setError] = useState<string | null>(null);
  const [showConfig, setShowConfig] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [awsProfile, setAwsProfile] = useState<string>('');
  const [isLoggingIn, setIsLoggingIn] = useState(false);
  const [viewMode, setViewMode] = useState<'cards' | 'list'>('cards');
  const [operationFeedback, setOperationFeedback] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
  const [pageSize, setPageSize] = useState(10);

  const loadAWSProfile = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const profile = await window.go.main.App.GetCurrentAWSProfile();
        setAwsProfile(profile);
        
        if (profile) {
          try {
            const server = await window.go.main.App.GetArgoCDServerFromProfile();
            setConfig(prev => ({ ...prev, server }));
          } catch (error) {
            console.warn('Failed to get ArgoCD server from profile:', error);
          }
        }
      }
    } catch (error) {
      console.error('Failed to load AWS profile:', error);
    }
  };

  const updateArgoCDServer = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const profile = await window.go.main.App.GetCurrentAWSProfile();
        setAwsProfile(profile);
        if (profile) {
          const server = await window.go.main.App.GetArgoCDServerFromProfile();
          setConfig(prev => ({ ...prev, server }));
          
          // Auto-load apps when server configuration changes
          if (server) {
            try {
              const updatedConfig = { ...config, server };
              const apps = await window.go.main.App.GetArgoApps(updatedConfig);
              setApps(apps);
              setError(null); // Clear any previous errors
            } catch (error) {
              console.warn('Failed to auto-load apps after server update:', error);
              setError(`Failed to load applications from ${server}`);
            }
          }
        } else {
          // No profile set, clear server configuration
          setConfig(prev => ({ ...prev, server: '' }));
          setApps([]);
          setError(null);
        }
      }
    } catch (error) {
      console.warn('Failed to update ArgoCD server:', error);
    }
  };

  const loadApps = async () => {
    if (!config.server) return;
    
    setLoading(true);
    setError(null);
    try {
      const apps = await window.go.main.App.GetArgoApps(config);
      setApps(apps);
    } catch (error) {
      setError(`Failed to load applications: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const handleLogin = async () => {
    setIsLoggingIn(true);
    setError(null);
    try {
      await window.go.main.App.LoginToArgoCD(config);
      await loadApps();
    } catch (error) {
      setError(`Login failed: ${error}`);
    } finally {
      setIsLoggingIn(false);
    }
  };

  const handleRefresh = async () => {
    await loadApps();
  };

  const handleOperationFeedback = (message: string, type: 'success' | 'error') => {
    setOperationFeedback({ message, type });
    // Auto-hide feedback after 5 seconds
    setTimeout(() => setOperationFeedback(null), 5000);
  };

  useEffect(() => {
    updateArgoCDServer();
  }, []);

  // Listen for profile changes from the parent component
  useEffect(() => {
    if (profileChangeCounter > 0) {
      updateArgoCDServer();
    }
  }, [profileChangeCounter]);

  useEffect(() => {
    if (config.server && autoRefresh) {
      const interval = setInterval(loadApps, 30000);
      return () => clearInterval(interval);
    }
  }, [config.server, autoRefresh]);

  const filteredApps = apps
    .filter(app => app?.AppName?.toLowerCase().includes(''))
    .sort((a, b) => (a?.AppName || '').localeCompare(b?.AppName || ''));

  return (
    <div style={{ padding: '24px' }}>
      <Row justify="space-between" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2}>
            <ApiOutlined style={{ marginRight: '8px' }} />
            ArgoCD Applications
          </Title>
          <Text type="secondary">Manage your ArgoCD applications</Text>
        </Col>
        <Col>
          <Space>
            <Button.Group>
              <Button 
                type={viewMode === 'cards' ? 'primary' : 'default'}
                icon={<AppstoreOutlined />}
                onClick={() => setViewMode('cards')}
              >
                Cards
              </Button>
              <Button 
                type={viewMode === 'list' ? 'primary' : 'default'}
                icon={<BarsOutlined />}
                onClick={() => setViewMode('list')}
              >
                List
              </Button>
            </Button.Group>
            <Divider type="vertical" />
            <Switch
              checked={autoRefresh}
              onChange={setAutoRefresh}
              checkedChildren="Auto"
              unCheckedChildren="Manual"
            />
            <Button 
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
            >
              Refresh
            </Button>
            <Button 
              icon={<SettingOutlined />}
              onClick={() => setShowConfig(!showConfig)}
            >
              Config
            </Button>
          </Space>
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

      {operationFeedback && (
        <Alert
          message={operationFeedback.message}
          type={operationFeedback.type}
          showIcon
          closable
          onClose={() => setOperationFeedback(null)}
          style={{ marginBottom: '16px' }}
        />
      )}

      {showConfig && (
        <Card style={{ marginBottom: '16px' }}>
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text strong>ArgoCD Configuration</Text>
            <Space style={{ width: '100%' }}>
              <Input
                placeholder="ArgoCD Server"
                value={config.server}
                onChange={(e) => setConfig(prev => ({ ...prev, server: e.target.value }))}
                style={{ flex: 1 }}
                prefix={<CloudServerOutlined />}
              />
              <Input
                placeholder="Project"
                value={config.project}
                onChange={(e) => setConfig(prev => ({ ...prev, project: e.target.value }))}
                style={{ width: '200px' }}
              />
            </Space>
            <Space>
              <Button
                type="primary"
                onClick={handleLogin}
                loading={isLoggingIn}
                disabled={!config.server}
              >
                {isLoggingIn ? 'Logging in...' : 'Login to ArgoCD'}
              </Button>
              <Text type="secondary">
                {awsProfile ? `AWS Profile: ${awsProfile}` : 'No AWS profile set'}
              </Text>
            </Space>
          </Space>
        </Card>
      )}

      <div style={{ minHeight: '400px' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <Spin size="large" />
            <div style={{ marginTop: '16px' }}>Loading applications...</div>
          </div>
        ) : filteredApps.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <CloudServerOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
            <div style={{ marginTop: '16px', color: '#999' }}>
              {config.server ? 'No applications found' : 'Configure ArgoCD server to get started'}
            </div>
          </div>
        ) : viewMode === 'cards' ? (
          <Row gutter={[16, 16]}>
            {filteredApps.map((app, index) => (
              <Col xs={24} sm={12} md={8} lg={6} key={app?.AppName || `app-${index}`}>
                <AppCard
                  app={app}
                  config={config}
                  onAction={handleRefresh}
                  onOperationFeedback={handleOperationFeedback}
                />
              </Col>
            ))}
          </Row>
        ) : (
          <Table
            dataSource={filteredApps.map(app => ({
              key: app.AppName,
              app: app,
            }))}
            columns={[
              {
                title: 'Application',
                dataIndex: 'app',
                key: 'appName',
                sorter: (a, b) => a.app.AppName.localeCompare(b.app.AppName),
                sortDirections: ['ascend', 'descend'],
                defaultSortOrder: 'ascend',
                render: (app: ArgoApp) => (
                  <Space>
                    <ApiOutlined />
                    <Text strong>{app.AppName}</Text>
                  </Space>
                ),
              },
              {
                title: 'Health',
                dataIndex: 'app',
                key: 'health',
                width: 100,
                filters: [
                  { text: 'Healthy', value: 'healthy' },
                  { text: 'Progressing', value: 'progressing' },
                  { text: 'Degraded', value: 'degraded' },
                  { text: 'Suspended', value: 'suspended' },
                ],
                onFilter: (value, record) => record.app.Health.toLowerCase().includes(value.toLowerCase()),
                render: (app: ArgoApp) => <StatusBadge status={app.Health} type="health" />,
              },
              {
                title: 'Sync',
                dataIndex: 'app',
                key: 'sync',
                width: 100,
                filters: [
                  { text: 'Synced', value: 'synced' },
                  { text: 'OutOfSync', value: 'outofsync' },
                ],
                onFilter: (value, record) => record.app.Sync.toLowerCase().includes(value.toLowerCase()),
                render: (app: ArgoApp) => <StatusBadge status={app.Sync} type="sync" />,
              },
              {
                title: 'Sync Loop',
                dataIndex: 'app',
                key: 'syncLoop',
                width: 100,
                render: (app: ArgoApp) => <StatusBadge status={app.SyncLoop} type="syncLoop" />,
              },
              {
                title: 'Suspended',
                dataIndex: 'app',
                key: 'suspended',
                width: 100,
                filters: [
                  { text: 'Yes', value: true },
                  { text: 'No', value: false },
                ],
                onFilter: (value, record) => record.app.Suspended === value,
                render: (app: ArgoApp) => (
                  <Tag color={app.Suspended ? 'red' : 'green'}>
                    {app.Suspended ? 'Yes' : 'No'}
                  </Tag>
                ),
              },
              {
                title: 'Conditions',
                dataIndex: 'app',
                key: 'conditions',
                width: 200,
                render: (app: ArgoApp) => app.Conditions && app.Conditions.length > 0 ? (
                  <Space wrap>
                    {app.Conditions.map((condition, index) => (
                      <Tag key={index} size="small">
                        {condition}
                      </Tag>
                    ))}
                  </Space>
                ) : (
                  <Text type="secondary">None</Text>
                ),
              },
              {
                title: 'Actions',
                dataIndex: 'app',
                key: 'actions',
                width: 400,
                render: (app: ArgoApp) => (
                  <AppListItem
                    app={app}
                    config={config}
                    onAction={handleRefresh}
                    onOperationFeedback={handleOperationFeedback}
                  />
                ),
              },
            ]}
            pagination={{ 
              pageSize: pageSize,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} applications`,
              pageSizeOptions: ['10', '20', '50', '100'],
              onShowSizeChange: (current, size) => setPageSize(size)
            }}
            size="small"
            loading={loading}
            scroll={{ x: 1200 }}
          />
        )}
      </div>
    </div>
  );
};

export default ArgoCD;