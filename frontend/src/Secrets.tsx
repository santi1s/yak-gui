import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Button, 
  Space, 
  Typography, 
  Input, 
  Alert, 
  Spin, 
  Row, 
  Col,
  Select,
  Modal,
  Popconfirm,
  Form
} from 'antd';
import {
  SafetyOutlined,
  ReloadOutlined,
  KeyOutlined,
  EyeOutlined,
  FolderOutlined,
  EditOutlined,
  DeleteOutlined,
  PlusOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;
const { Option } = Select;

// Types matching the Go backend
interface SecretListItem {
  path: string;
  version: number;
  owner: string;
  usage: string;
  source: string;
  createdAt: string;
  updatedAt: string;
}

interface SecretData {
  path: string;
  version: number;
  data: Record<string, string>;
  metadata: SecretMetadata;
}

interface SecretMetadata {
  owner: string;
  usage: string;
  source: string;
  createdAt: string;
  updatedAt: string;
  version: number;
  destroyed: boolean;
}

interface SecretConfig {
  platform: string;
  environment: string;
}

interface JWTClientConfig {
  platform: string;
  environment: string;
  path: string;
  owner: string;
  localName: string;
  targetService: string;
  secret: string;
}

interface JWTServerConfig {
  platform: string;
  environment: string;
  path: string;
  owner: string;
  localName: string;
  serviceName: string;
  clientName: string;
  clientSecret: string;
}

const Secrets: React.FC = () => {
  const [secrets, setSecrets] = useState<SecretListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPath, setCurrentPath] = useState<string>('');
  const [config, setConfig] = useState<SecretConfig>({
    platform: 'dev',
    environment: ''
  });
  const [selectedSecret, setSelectedSecret] = useState<SecretListItem | null>(null);
  const [secretData, setSecretData] = useState<SecretData | null>(null);
  const [showSecretDialog, setShowSecretDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [editingSecretData, setEditingSecretData] = useState<Record<string, string>>({});
  const [breadcrumbs, setBreadcrumbs] = useState<string[]>([]);
  const [showJWTClientDialog, setShowJWTClientDialog] = useState(false);
  const [showJWTServerDialog, setShowJWTServerDialog] = useState(false);
  const [jwtClientForm, setJwtClientForm] = useState<JWTClientConfig>({
    platform: '',
    environment: '',
    path: '',
    owner: '',
    localName: '',
    targetService: '',
    secret: ''
  });
  const [jwtServerForm, setJwtServerForm] = useState<JWTServerConfig>({
    platform: '',
    environment: '',
    path: '',
    owner: '',
    localName: '',
    serviceName: '',
    clientName: '',
    clientSecret: ''
  });
  
  // Dynamic configuration state
  const [availablePlatforms, setAvailablePlatforms] = useState<string[]>([]);
  const [availableEnvironments, setAvailableEnvironments] = useState<string[]>([]);
  const [availablePaths, setAvailablePaths] = useState<string[]>([]);
  const [configLoading, setConfigLoading] = useState(false);
  const [pathsLoading, setPathsLoading] = useState(false);

  const loadSecrets = async () => {
    if (!config.platform) {
      setError('Platform is required');
      return;
    }

    if (!window.go || !window.go.main || !window.go.main.App) {
      setError('Wails bindings not available. Please wait for the app to fully load.');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const secretsData = await window.go.main.App.GetSecrets(config, currentPath);
      
      console.log('Secrets data received:', secretsData);
      console.log('Secrets data type:', typeof secretsData);
      console.log('Is array:', Array.isArray(secretsData));
      
      if (Array.isArray(secretsData)) {
        setSecrets(secretsData);
      } else {
        console.error('Received non-array data:', secretsData);
        setError(`Expected array, got ${typeof secretsData}`);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error('Failed to load secrets:', error);
      setError(`Failed to load secrets: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  // Load dynamic configuration
  const loadConfiguration = async () => {
    setConfigLoading(true);
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const platforms = await window.go.main.App.GetSecretConfigPlatforms();
        setAvailablePlatforms(platforms);
        
        // Set default platform if available
        if (platforms.length > 0 && !config.platform) {
          setConfig(prev => ({ ...prev, platform: platforms[0] }));
        }
      }
    } catch (error) {
      console.error('Failed to load secret configuration:', error);
      setError('Failed to load secret configuration');
    } finally {
      setConfigLoading(false);
    }
  };

  // Load environments when platform changes
  const loadEnvironments = async (platform: string) => {
    if (!platform) {
      setAvailableEnvironments([]);
      return;
    }
    
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const environments = await window.go.main.App.GetSecretConfigEnvironments(platform);
        setAvailableEnvironments(environments);
      }
    } catch (error) {
      console.error('Failed to load environments:', error);
      setAvailableEnvironments([]);
    }
  };

  // Load paths when platform or environment changes
  const loadPaths = async (platform: string, environment: string) => {
    if (!platform) {
      setAvailablePaths([]);
      return;
    }
    
    setPathsLoading(true);
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const paths = await window.go.main.App.GetSecretConfigPaths(platform, environment || '');
        setAvailablePaths(paths);
      }
    } catch (error) {
      console.error('Failed to load paths:', error);
      setAvailablePaths(['']); // Fallback to empty path
    } finally {
      setPathsLoading(false);
    }
  };

  useEffect(() => {
    loadConfiguration();
  }, []);

  useEffect(() => {
    loadEnvironments(config.platform);
    loadPaths(config.platform, config.environment);
  }, [config.platform]);

  useEffect(() => {
    loadPaths(config.platform, config.environment);
  }, [config.environment]);

  // Reload paths when navigating to different directories
  useEffect(() => {
    if (config.platform) {
      loadPaths(config.platform, config.environment);
    }
  }, [currentPath]);

  useEffect(() => {
    loadSecrets();
  }, [config, currentPath]);

  const handleNavigate = (path: string) => {
    // Append the new path to current path to build full path
    const newPath = currentPath ? `${currentPath.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}` : path;
    setCurrentPath(newPath);
    // Build breadcrumbs from the full path
    const pathParts = newPath.split('/').filter(part => part !== '');
    setBreadcrumbs(pathParts);
  };

  const handleBreadcrumbClick = (index: number) => {
    const newBreadcrumbs = breadcrumbs.slice(0, index + 1);
    const newPath = newBreadcrumbs.join('/') + '/';
    setCurrentPath(newPath);
    setBreadcrumbs(newBreadcrumbs);
  };

  const handleGoBack = () => {
    if (breadcrumbs.length > 0) {
      const newBreadcrumbs = breadcrumbs.slice(0, -1);
      const newPath = newBreadcrumbs.length > 0 ? newBreadcrumbs.join('/') + '/' : '';
      setCurrentPath(newPath);
      setBreadcrumbs(newBreadcrumbs);
    } else {
      setCurrentPath('');
      setBreadcrumbs([]);
    }
  };

  const handleViewSecret = async (secret: SecretListItem) => {
    setSelectedSecret(secret);
    try {
      // Construct the full path by combining current path with secret path
      // Remove any trailing slash from currentPath and ensure proper path construction
      let fullPath = secret.path;
      if (currentPath && currentPath.trim() !== '') {
        const cleanCurrentPath = currentPath.replace(/\/+$/, ''); // Remove trailing slashes
        const cleanSecretPath = secret.path.replace(/^\/+/, ''); // Remove leading slashes
        fullPath = `${cleanCurrentPath}/${cleanSecretPath}`;
      }
      
      console.log('Getting secret with currentPath:', currentPath);
      console.log('Getting secret with secret.path:', secret.path);
      console.log('Getting secret with constructed fullPath:', fullPath);
      
      const data = await window.go.main.App.GetSecretData(config, fullPath, secret.version);
      console.log('Secret data received:', data);
      
      // Validate the received data
      if (!data) {
        throw new Error('No data returned from secret get operation');
      }
      
      // Ensure data.data exists and is an object
      if (!data.data || typeof data.data !== 'object') {
        console.warn('Secret data.data is missing or invalid:', data);
        data.data = {}; // Initialize as empty object to prevent crashes
      }
      
      // Ensure metadata exists
      if (!data.metadata) {
        console.warn('Secret metadata is missing:', data);
        data.metadata = {
          owner: 'Unknown',
          usage: 'Unknown',
          source: 'Unknown',
          version: secret.version || 1,
          destroyed: false,
          createdAt: '',
          updatedAt: ''
        };
      }
      
      setSecretData(data);
      setShowSecretDialog(true);
    } catch (error) {
      console.error('Error loading secret data:', error);
      setError(`Failed to load secret data: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleEditSecret = async (secret: SecretListItem) => {
    try {
      // First get the current secret data
      let fullPath = secret.path;
      if (currentPath && currentPath.trim() !== '') {
        const cleanCurrentPath = currentPath.replace(/\/+$/, ''); 
        const cleanSecretPath = secret.path.replace(/^\/+/, ''); 
        fullPath = `${cleanCurrentPath}/${cleanSecretPath}`;
      }
      
      const data = await window.go.main.App.GetSecretData(config, fullPath, secret.version);
      
      if (data && data.data) {
        setSelectedSecret(secret);
        setEditingSecretData(data.data);
        setShowEditDialog(true);
      }
    } catch (error) {
      setError(`Failed to load secret for editing: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleUpdateSecret = async () => {
    if (!selectedSecret) return;
    
    try {
      let fullPath = selectedSecret.path;
      if (currentPath && currentPath.trim() !== '') {
        const cleanCurrentPath = currentPath.replace(/\/+$/, '');
        const cleanSecretPath = selectedSecret.path.replace(/^\/+/, '');
        fullPath = `${cleanCurrentPath}/${cleanSecretPath}`;
      }
      
      await window.go.main.App.UpdateSecret(config, fullPath, editingSecretData);
      setShowEditDialog(false);
      setEditingSecretData({});
      setSelectedSecret(null);
      await loadSecrets(); // Refresh the list
    } catch (error) {
      setError(`Failed to update secret: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleDeleteSecret = async (secret: SecretListItem) => {
    try {
      let fullPath = secret.path;
      if (currentPath && currentPath.trim() !== '') {
        const cleanCurrentPath = currentPath.replace(/\/+$/, '');
        const cleanSecretPath = secret.path.replace(/^\/+/, '');
        fullPath = `${cleanCurrentPath}/${cleanSecretPath}`;
      }
      
      await window.go.main.App.DeleteSecret(config, fullPath, secret.version);
      await loadSecrets(); // Refresh the list
    } catch (error) {
      setError(`Failed to delete secret: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleCreateSecret = async (path: string, owner: string, usage: string, source: string, data: Record<string, string>) => {
    try {
      let fullPath = path;
      if (currentPath && currentPath.trim() !== '') {
        const cleanCurrentPath = currentPath.replace(/\/+$/, '');
        const cleanPath = path.replace(/^\/+/, '');
        fullPath = `${cleanCurrentPath}/${cleanPath}`;
      }
      
      await window.go.main.App.CreateSecret(config, fullPath, owner, usage, source, data);
      setShowCreateDialog(false);
      await loadSecrets(); // Refresh the list
    } catch (error) {
      setError(`Failed to create secret: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  // Filter available paths based on current directory context
  const getFilteredPaths = () => {
    if (!currentPath) {
      // If at root, show all top-level paths
      return availablePaths.filter(path => {
        if (path === '') return true;
        const parts = path.replace(/\/+$/, '').split('/');
        return parts.length === 1; // Only show direct children of root
      });
    } else {
      // If in a subdirectory, show paths that are children of current path
      const currentClean = currentPath.replace(/\/+$/, '');
      return availablePaths.filter(path => {
        if (path === '') return true; // Always allow going back to root
        if (path.startsWith(currentClean + '/')) {
          const relativePath = path.substring(currentClean.length + 1);
          const parts = relativePath.replace(/\/+$/, '').split('/');
          return parts.length === 1; // Only show direct children
        }
        return false;
      });
    }
  };

  const handleCreateJWTClient = async () => {
    try {
      await window.go.main.App.CreateJWTClient(jwtClientForm);
      setShowJWTClientDialog(false);
      setJwtClientForm({
        platform: '',
        environment: '',
        path: '',
        owner: '',
        localName: '',
        targetService: '',
        secret: ''
      });
      await loadSecrets(); // Refresh the list
    } catch (error) {
      setError(`Failed to create JWT client: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleCreateJWTServer = async () => {
    try {
      await window.go.main.App.CreateJWTServer(jwtServerForm);
      setShowJWTServerDialog(false);
      setJwtServerForm({
        platform: '',
        environment: '',
        path: '',
        owner: '',
        localName: '',
        serviceName: '',
        clientName: '',
        clientSecret: ''
      });
      await loadSecrets(); // Refresh the list
    } catch (error) {
      setError(`Failed to create JWT server: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      <Row justify="space-between" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2}>
            <SafetyOutlined style={{ marginRight: '8px' }} />
            Secrets Management
          </Title>
          <Text type="secondary">Manage your secrets</Text>
        </Col>
        <Col>
          <Space>
            <Button 
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setShowCreateDialog(true)}
            >
              Create Secret
            </Button>
            <Button 
              icon={<KeyOutlined />}
              onClick={() => setShowJWTClientDialog(true)}
            >
              JWT Client
            </Button>
            <Button 
              icon={<KeyOutlined />}
              onClick={() => setShowJWTServerDialog(true)}
            >
              JWT Server
            </Button>
            <Button 
              icon={<ReloadOutlined />}
              onClick={loadSecrets}
              loading={loading}
            >
              Refresh
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

      <Card style={{ marginBottom: '16px' }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Row gutter={16}>
            <Col span={8}>
              <Text strong>Platform:</Text>
              <Select
                placeholder="Select platform..."
                value={config.platform}
                onChange={(value) => setConfig(prev => ({ ...prev, platform: value, environment: '' }))}
                style={{ width: '100%', marginLeft: '8px' }}
                loading={configLoading}
              >
                {availablePlatforms.map(platform => (
                  <Option key={platform} value={platform}>{platform}</Option>
                ))}
              </Select>
            </Col>
            <Col span={8}>
              <Text strong>Environment:</Text>
              <Select
                placeholder="Select environment..."
                value={config.environment}
                onChange={(value) => setConfig(prev => ({ ...prev, environment: value }))}
                style={{ width: '100%', marginLeft: '8px' }}
                allowClear
                disabled={!config.platform}
              >
                <Option value="">All environments</Option>
                {availableEnvironments.map(env => (
                  <Option key={env} value={env}>{env}</Option>
                ))}
              </Select>
            </Col>
            <Col span={8}>
              <Text strong>Path:</Text>
              <Select
                placeholder="Select path prefix..."
                value={currentPath}
                onChange={(value) => {
                  setCurrentPath(value || '');
                  // Update breadcrumbs when path is selected from dropdown
                  if (value) {
                    const pathParts = value.split('/').filter(part => part !== '');
                    setBreadcrumbs(pathParts);
                  } else {
                    setBreadcrumbs([]);
                  }
                }}
                style={{ width: '100%', marginLeft: '8px' }}
                allowClear
                showSearch
                loading={pathsLoading}
                disabled={!config.platform || pathsLoading}
                filterOption={(input, option) =>
                  option?.children?.toString().toLowerCase().includes(input.toLowerCase())
                }
              >
                {getFilteredPaths().map(path => (
                  <Option key={path} value={path}>
                    {path === '' ? 'All paths' : path}
                  </Option>
                ))}
              </Select>
            </Col>
          </Row>
        </Space>
      </Card>

      {/* Breadcrumb Navigation */}
      {(breadcrumbs.length > 0 || currentPath) && (
        <Card style={{ marginBottom: '16px' }}>
          <Space>
            <Button
              size="small"
              onClick={handleGoBack}
              icon={<span>←</span>}
            >
              Back
            </Button>
            
            <span style={{ color: '#999' }}>/</span>
            
            <Button
              type="link"
              size="small"
              onClick={() => {
                setCurrentPath('');
                setBreadcrumbs([]);
              }}
              style={{ padding: 0, height: 'auto' }}
            >
              root
            </Button>
            
            {breadcrumbs.map((crumb, index) => (
              <Space key={index} size={0}>
                <span style={{ color: '#999' }}>/</span>
                <Button
                  type="link"
                  size="small"
                  onClick={() => handleBreadcrumbClick(index)}
                  style={{ padding: 0, height: 'auto', color: '#1890ff' }}
                >
                  {crumb}
                </Button>
              </Space>
            ))}
          </Space>
        </Card>
      )}

      <div style={{ minHeight: '400px' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <Spin size="large" />
            <div style={{ marginTop: '16px' }}>Loading secrets...</div>
          </div>
        ) : secrets.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <KeyOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
            <div style={{ marginTop: '16px', color: '#999' }}>
              No secrets found
            </div>
          </div>
        ) : (
          <Row gutter={[16, 16]}>
            {secrets.map((secret, index) => {
              const isFolder = secret.path.endsWith('/');
              const displayName = isFolder ? secret.path.slice(0, -1) : secret.path;
              
              return (
                <Col xs={24} sm={12} md={8} lg={6} key={secret?.path || `secret-${index}`}>
                  <Card
                    title={
                      <Space>
                        {isFolder ? (
                          <FolderOutlined style={{ color: '#1890ff' }} />
                        ) : (
                          <KeyOutlined style={{ color: '#faad14' }} />
                        )}
                        <span>{displayName}</span>
                      </Space>
                    }
                    size="small"
                    style={{ 
                      marginBottom: 16,
                      cursor: isFolder ? 'pointer' : 'default'
                    }}
                    onClick={isFolder ? () => handleNavigate(secret.path) : undefined}
                    actions={
                      isFolder ? [
                        <Button
                          key="open"
                          type="text"
                          icon={<FolderOutlined />}
                          onClick={(e) => {
                            e.stopPropagation();
                            handleNavigate(secret.path);
                          }}
                          style={{ color: '#1890ff' }}
                        >
                          Open Folder
                        </Button>
                      ] : [
                        <Button
                          key="view"
                          type="text"
                          icon={<EyeOutlined />}
                          onClick={() => handleViewSecret(secret)}
                        >
                          View
                        </Button>,
                        <Button
                          key="edit"
                          type="text"
                          icon={<EditOutlined />}
                          onClick={() => handleEditSecret(secret)}
                        >
                          Edit
                        </Button>,
                        <Popconfirm
                          key="delete"
                          title="Are you sure you want to delete this secret?"
                          onConfirm={() => handleDeleteSecret(secret)}
                          okText="Yes"
                          cancelText="No"
                        >
                          <Button
                            type="text"
                            danger
                            icon={<DeleteOutlined />}
                          >
                            Delete
                          </Button>
                        </Popconfirm>
                      ]
                    }
                  >
                    <Space direction="vertical" style={{ width: '100%' }}>
                      <Row justify="space-between">
                        <Col><Text type="secondary">Owner:</Text></Col>
                        <Col><Text>{secret.owner}</Text></Col>
                      </Row>
                      <Row justify="space-between">
                        <Col><Text type="secondary">Usage:</Text></Col>
                        <Col><Text>{secret.usage}</Text></Col>
                      </Row>
                      {!isFolder && (
                        <Row justify="space-between">
                          <Col><Text type="secondary">Version:</Text></Col>
                          <Col><Text code>{secret.version}</Text></Col>
                        </Row>
                      )}
                    </Space>
                  </Card>
                </Col>
              );
            })}
          </Row>
        )}
      </div>

      {/* View Secret Modal */}
      {showSecretDialog && secretData && (
        <Modal
          title={`Secret: ${secretData.path || 'Unknown'}`}
          open={showSecretDialog}
          onCancel={() => setShowSecretDialog(false)}
          footer={[
            <Button key="close" onClick={() => setShowSecretDialog(false)}>
              Close
            </Button>
          ]}
          width={600}
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <Text strong>Metadata:</Text>
              <div style={{ marginTop: 8, fontSize: '12px' }}>
                <Text type="secondary">Owner: {secretData.metadata?.owner || 'Unknown'}</Text><br />
                <Text type="secondary">Usage: {secretData.metadata?.usage || 'Unknown'}</Text><br />
                <Text type="secondary">Source: {secretData.metadata?.source || 'Unknown'}</Text><br />
                <Text type="secondary">Version: {secretData.metadata?.version || 'Unknown'}</Text>
              </div>
            </div>
            <div>
              <Text strong>Data:</Text>
              <div style={{ marginTop: 8 }}>
                {secretData.data && typeof secretData.data === 'object' ? (
                  Object.entries(secretData.data).map(([key, value]) => (
                    <Row key={key} justify="space-between" style={{ marginBottom: 8 }}>
                      <Col><Text code>{key}:</Text></Col>
                      <Col><Input.Password value={String(value)} readOnly style={{ width: 200 }} /></Col>
                    </Row>
                  ))
                ) : (
                  <Text type="secondary">No secret data available</Text>
                )}
              </div>
            </div>
          </Space>
        </Modal>
      )}

      {/* Edit Secret Modal */}
      {showEditDialog && selectedSecret && (
        <Modal
          title={`Edit Secret: ${selectedSecret.path}`}
          open={showEditDialog}
          onCancel={() => {
            setShowEditDialog(false);
            setEditingSecretData({});
            setSelectedSecret(null);
          }}
          footer={[
            <Button key="cancel" onClick={() => {
              setShowEditDialog(false);
              setEditingSecretData({});
              setSelectedSecret(null);
            }}>
              Cancel
            </Button>,
            <Button key="save" type="primary" onClick={handleUpdateSecret}>
              Save Changes
            </Button>
          ]}
          width={700}
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text>Edit the key-value pairs for this secret:</Text>
            {Object.entries(editingSecretData).map(([key, value]) => (
              <Row key={key} gutter={16} style={{ marginBottom: 8 }}>
                <Col span={8}>
                  <Input
                    value={key}
                    disabled
                    addonBefore="Key"
                  />
                </Col>
                <Col span={16}>
                  <Input.Password
                    value={value}
                    onChange={(e) => setEditingSecretData(prev => ({
                      ...prev,
                      [key]: e.target.value
                    }))}
                    placeholder="Secret value"
                  />
                </Col>
              </Row>
            ))}
            <Button
              type="dashed"
              onClick={() => {
                const newKey = `new_key_${Date.now()}`;
                setEditingSecretData(prev => ({
                  ...prev,
                  [newKey]: ''
                }));
              }}
              style={{ width: '100%' }}
              icon={<PlusOutlined />}
            >
              Add New Key-Value Pair
            </Button>
          </Space>
        </Modal>
      )}

      {/* Create Secret Modal */}
      {showCreateDialog && (
        <CreateSecretModal
          visible={showCreateDialog}
          onCancel={() => setShowCreateDialog(false)}
          onSubmit={handleCreateSecret}
          currentPath={currentPath}
        />
      )}

      {/* JWT Client Modal */}
      {showJWTClientDialog && (
        <Modal
          title="Create JWT Client Secret"
          open={showJWTClientDialog}
          onCancel={() => setShowJWTClientDialog(false)}
          footer={[
            <Button key="cancel" onClick={() => setShowJWTClientDialog(false)}>
              Cancel
            </Button>,
            <Button 
              key="create" 
              type="primary" 
              onClick={handleCreateJWTClient}
              disabled={!jwtClientForm.path || !jwtClientForm.owner || !jwtClientForm.localName || !jwtClientForm.targetService || !jwtClientForm.secret}
            >
              Create JWT Client
            </Button>
          ]}
          width={600}
        >
          <Form layout="vertical">
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Platform">
                  <Input
                    value={jwtClientForm.platform}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, platform: e.target.value })}
                    placeholder="e.g., dev"
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Environment">
                  <Input
                    value={jwtClientForm.environment}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, environment: e.target.value })}
                    placeholder="e.g., staging"
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item label="Path" required>
              <Input
                value={jwtClientForm.path}
                onChange={(e) => setJwtClientForm({ ...jwtClientForm, path: e.target.value })}
                placeholder="e.g., personal-assistant/jwt"
              />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Owner" required>
                  <Input
                    value={jwtClientForm.owner}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, owner: e.target.value })}
                    placeholder="e.g., SRE Team"
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Local Name" required>
                  <Input
                    value={jwtClientForm.localName}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, localName: e.target.value })}
                    placeholder="e.g., personal-assistant"
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item label="Target Service" required>
              <Input
                value={jwtClientForm.targetService}
                onChange={(e) => setJwtClientForm({ ...jwtClientForm, targetService: e.target.value })}
                placeholder="e.g., api-service"
              />
            </Form.Item>
            <Form.Item label="Secret" required>
              <Input.Password
                value={jwtClientForm.secret}
                onChange={(e) => setJwtClientForm({ ...jwtClientForm, secret: e.target.value })}
                placeholder="Enter secret value"
              />
            </Form.Item>
          </Form>
        </Modal>
      )}

      {/* JWT Server Modal */}
      {showJWTServerDialog && (
        <Modal
          title="Create JWT Server Secret"
          open={showJWTServerDialog}
          onCancel={() => setShowJWTServerDialog(false)}
          footer={[
            <Button key="cancel" onClick={() => setShowJWTServerDialog(false)}>
              Cancel
            </Button>,
            <Button 
              key="create" 
              type="primary" 
              onClick={handleCreateJWTServer}
              disabled={!jwtServerForm.path || !jwtServerForm.owner || !jwtServerForm.localName || !jwtServerForm.serviceName || !jwtServerForm.clientName || !jwtServerForm.clientSecret}
            >
              Create JWT Server
            </Button>
          ]}
          width={600}
        >
          <Form layout="vertical">
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Platform">
                  <Input
                    value={jwtServerForm.platform}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, platform: e.target.value })}
                    placeholder="e.g., dev"
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Environment">
                  <Input
                    value={jwtServerForm.environment}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, environment: e.target.value })}
                    placeholder="e.g., staging"
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item label="Path" required>
              <Input
                value={jwtServerForm.path}
                onChange={(e) => setJwtServerForm({ ...jwtServerForm, path: e.target.value })}
                placeholder="e.g., api-service/jwt"
              />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Owner" required>
                  <Input
                    value={jwtServerForm.owner}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, owner: e.target.value })}
                    placeholder="e.g., Backend Team"
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Local Name" required>
                  <Input
                    value={jwtServerForm.localName}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, localName: e.target.value })}
                    placeholder="e.g., api-service"
                  />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Service Name" required>
                  <Input
                    value={jwtServerForm.serviceName}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, serviceName: e.target.value })}
                    placeholder="e.g., api-service"
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Client Name" required>
                  <Input
                    value={jwtServerForm.clientName}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, clientName: e.target.value })}
                    placeholder="e.g., personal-assistant"
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item label="Client Secret" required>
              <Input.Password
                value={jwtServerForm.clientSecret}
                onChange={(e) => setJwtServerForm({ ...jwtServerForm, clientSecret: e.target.value })}
                placeholder="Enter client secret value"
              />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </div>
  );
};

// Create Secret Modal Component
const CreateSecretModal: React.FC<{
  visible: boolean;
  onCancel: () => void;
  onSubmit: (path: string, owner: string, usage: string, source: string, data: Record<string, string>) => void;
  currentPath: string;
}> = ({ visible, onCancel, onSubmit, currentPath }) => {
  const [form] = Form.useForm();
  const [secretData, setSecretData] = useState<Record<string, string>>({});

  const handleSubmit = () => {
    form.validateFields().then((values) => {
      onSubmit(values.path, values.owner, values.usage, values.source, secretData);
      form.resetFields();
      setSecretData({});
    });
  };

  return (
    <Modal
      title="Create New Secret"
      open={visible}
      onCancel={() => {
        onCancel();
        form.resetFields();
        setSecretData({});
      }}
      footer={[
        <Button key="cancel" onClick={() => {
          onCancel();
          form.resetFields();
          setSecretData({});
        }}>
          Cancel
        </Button>,
        <Button key="create" type="primary" onClick={handleSubmit}>
          Create Secret
        </Button>
      ]}
      width={800}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="path"
          label="Secret Path"
          rules={[{ required: true, message: 'Please enter the secret path' }]}
          initialValue=""
        >
          <Input 
            placeholder="e.g., myapp/database" 
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="off"
            spellCheck={false}
          />
        </Form.Item>
        
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item
              name="owner"
              label="Owner"
              rules={[{ required: true, message: 'Please enter the owner' }]}
            >
              <Input 
                placeholder="e.g., SRE, Backend Team" 
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="usage"
              label="Usage"
              rules={[{ required: true, message: 'Please enter the usage' }]}
            >
              <Input 
                placeholder="e.g., used by myapp" 
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="source"
              label="Source"
              rules={[{ required: true, message: 'Please enter the source' }]}
            >
              <Input 
                placeholder="e.g., manual, terraform" 
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
              />
            </Form.Item>
          </Col>
        </Row>

        <div style={{ marginTop: 16 }}>
          <Text strong>Secret Data:</Text>
          <div style={{ marginTop: 8 }}>
            {Object.entries(secretData).map(([key, value]) => (
              <Row key={key} gutter={16} style={{ marginBottom: 8 }}>
                <Col span={8}>
                  <Input
                    value={key}
                    onChange={(e) => {
                      const newKey = e.target.value;
                      const newData = { ...secretData };
                      delete newData[key];
                      newData[newKey] = value;
                      setSecretData(newData);
                    }}
                    placeholder="Key name"
                    autoComplete="off"
                    autoCorrect="off"
                    autoCapitalize="off"
                    spellCheck={false}
                  />
                </Col>
                <Col span={12}>
                  <Input.Password
                    value={value}
                    onChange={(e) => setSecretData(prev => ({
                      ...prev,
                      [key]: e.target.value
                    }))}
                    placeholder="Secret value"
                  />
                </Col>
                <Col span={4}>
                  <Button
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => {
                      const newData = { ...secretData };
                      delete newData[key];
                      setSecretData(newData);
                    }}
                  />
                </Col>
              </Row>
            ))}
            <Button
              type="dashed"
              onClick={() => {
                const newKey = `key_${Date.now()}`;
                setSecretData(prev => ({
                  ...prev,
                  [newKey]: ''
                }));
              }}
              style={{ width: '100%' }}
              icon={<PlusOutlined />}
            >
              Add Key-Value Pair
            </Button>
          </div>
        </div>
      </Form>
    </Modal>
  );
};

export default Secrets;