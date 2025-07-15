import React, { useState, useEffect } from 'react';
import {
  Layout,
  Card,
  Button,
  Input,
  Select,
  Space,
  Typography,
  Alert,
  Spin,
  Table,
  Tag,
  Row,
  Col,
  Tabs,
  Form,
  Switch,
  Tooltip,
  Progress,
  Modal,
  List,
  Divider,
} from 'antd';
import {
  CloudServerOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  SettingOutlined,
  FileTextOutlined,
  DatabaseOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  LockOutlined,
  UnlockOutlined,
  DeleteOutlined,
  EditOutlined,
  SearchOutlined,
} from '@ant-design/icons';

const { Title, Text } = Typography;
const { Option } = Select;

// TFE interfaces based on the yak tfe command analysis
interface TFEWorkspace {
  id: string;
  name: string;
  description?: string;
  environment?: string;
  terraformVersion?: string;
  status?: 'active' | 'locked' | 'disabled';
  lastRun?: string;
  owner?: string;
  tags?: string[];
  organization?: string;
  createdAt?: string;
  updatedAt?: string;
  autoApply?: boolean;
  terraformWorking?: boolean;
  vcsRepo?: {
    identifier: string;
    branch: string;
    ingressSubmodules: boolean;
  };
  variables?: Record<string, string>;
}

interface TFERun {
  id: string;
  workspaceId: string;
  workspaceName: string;
  status: 'pending' | 'planning' | 'planned' | 'applying' | 'applied' | 'discarded' | 'errored' | 'canceled';
  createdAt: string;
  message?: string;
  source?: string;
  terraformVersion?: string;
  hasChanges?: boolean;
  isDestroy?: boolean;
  isConfirmable?: boolean;
  actions?: {
    isConfirmable: boolean;
    isCancelable: boolean;
    isDiscardable: boolean;
  };
  createdBy?: string;
  url?: string;
}

interface TFEConfig {
  endpoint: string;
  organization: string;
  token?: string;
}


const TFE: React.FC = () => {
  const [config, setConfig] = useState<TFEConfig>({
    endpoint: 'app.terraform.io',
    organization: 'doctolib',
  });
  const [workspaces, setWorkspaces] = useState<TFEWorkspace[]>([]);
  const [filteredWorkspaces, setFilteredWorkspaces] = useState<TFEWorkspace[]>([]);
  const [runs, setRuns] = useState<TFERun[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedWorkspace, setSelectedWorkspace] = useState<string | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [activeTab, setActiveTab] = useState<'workspaces' | 'runs' | 'versions'>('workspaces');
  const [showConfigModal, setShowConfigModal] = useState(false);
  const [nameFilter, setNameFilter] = useState('');
  const [environmentFilter, setEnvironmentFilter] = useState('');
  const [pageSize, setPageSize] = useState(10);
  const [configForm] = Form.useForm();

  // Load workspaces
  const loadWorkspaces = async () => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEWorkspaces) {
        const workspaces = await window.go.main.App.GetTFEWorkspaces(config);
        setWorkspaces(workspaces);
        setFilteredWorkspaces(workspaces);
      } else {
        throw new Error('TFE backend not available');
      }
    } catch (error) {
      setError(`Failed to load workspaces: ${error}`);
      setWorkspaces([]);
      setFilteredWorkspaces([]);
    } finally {
      setLoading(false);
    }
  };

  // Load workspaces with real TFE tag filtering
  const loadWorkspacesByTag = async (tag: string, not: boolean = false) => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEWorkspacesByTag) {
        const workspaces = await window.go.main.App.GetTFEWorkspacesByTag(config, tag, not);
        setWorkspaces(workspaces);
        setFilteredWorkspaces(workspaces);
      } else {
        throw new Error('TFE tag filtering backend not available');
      }
    } catch (error) {
      setError(`Failed to load workspaces by tag: ${error}`);
      setWorkspaces([]);
      setFilteredWorkspaces([]);
    } finally {
      setLoading(false);
    }
  };

  // Load runs
  const loadRuns = async () => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFERuns) {
        const runs = await window.go.main.App.GetTFERuns(config, selectedWorkspace || '');
        setRuns(runs);
      } else {
        throw new Error('TFE runs backend not available');
      }
    } catch (error) {
      setError(`Failed to load runs: ${error}`);
      setRuns([]);
    } finally {
      setLoading(false);
    }
  };

  // Execute plan on workspace
  const executePlan = async (workspaceId: string, terraformVersion: string) => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.ExecuteTFEPlan) {
        // Find workspace name from workspaceId
        const workspace = workspaces.find(w => w.id === workspaceId);
        if (workspace) {
          const execution = {
            workspaceNames: [workspace.name],
            terraformVersion,
            wait: false,
          };
          await window.go.main.App.ExecuteTFEPlan(config, execution);
        } else {
          throw new Error('Workspace not found');
        }
      } else {
        throw new Error('TFE plan execution backend not available');
      }
      await loadRuns(); // Refresh runs after plan execution
    } catch (error) {
      setError(`Failed to execute plan: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  // Lock/unlock workspace
  const toggleWorkspaceLock = async (workspaceId: string, lock: boolean) => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.LockTFEWorkspace && window.go?.main?.App?.UnlockTFEWorkspace) {
        // Find workspace name from workspaceId
        const workspace = workspaces.find(w => w.id === workspaceId);
        if (workspace) {
          const workspaceNames = [workspace.name];
          if (lock) {
            await window.go.main.App.LockTFEWorkspace(config, workspaceNames, true);
          } else {
            await window.go.main.App.UnlockTFEWorkspace(config, workspaceNames, false);
          }
        } else {
          throw new Error('Workspace not found');
        }
      } else {
        throw new Error('TFE workspace lock/unlock backend not available');
      }
      await loadWorkspaces(); // Refresh workspaces after lock change
    } catch (error) {
      setError(`Failed to ${lock ? 'lock' : 'unlock'} workspace: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  // Load TFE configuration
  const loadTFEConfig = async () => {
    try {
      if (window.go?.main?.App?.GetTFEConfig) {
        const tfeConfig = await window.go.main.App.GetTFEConfig();
        setConfig(tfeConfig);
        configForm.setFieldsValue(tfeConfig);
      }
    } catch (error) {
      console.warn('Failed to load TFE config:', error);
    }
  };

  // Filter workspaces based on name and environment filters
  const filterWorkspaces = () => {
    let filtered = [...workspaces];

    // Filter by name
    if (nameFilter) {
      filtered = filtered.filter(workspace =>
        workspace.name.toLowerCase().includes(nameFilter.toLowerCase())
      );
    }

    // Filter by environment
    if (environmentFilter) {
      filtered = filtered.filter(workspace =>
        workspace.environment === environmentFilter
      );
    }

    setFilteredWorkspaces(filtered);
  };

  // Get unique environments from workspaces
  const getUniqueEnvironments = () => {
    const environments = workspaces
      .map(workspace => workspace.environment)
      .filter((env, index, self) => env && env !== 'unknown' && self.indexOf(env) === index)
      .sort();
    return environments;
  };

  // Handle configuration save
  const handleConfigSave = async (values: TFEConfig) => {
    try {
      if (window.go?.main?.App?.SetTFEConfig) {
        await window.go.main.App.SetTFEConfig(values);
      }
      setConfig(values);
      setShowConfigModal(false);
      // Reload workspaces with new config
      await loadWorkspaces();
    } catch (error) {
      setError(`Failed to save configuration: ${error}`);
    }
  };

  useEffect(() => {
    loadTFEConfig().then(() => {
      loadWorkspaces();
      loadRuns();
    });
  }, []);

  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(() => {
        loadWorkspaces();
        loadRuns();
      }, 30000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  // Filter workspaces when filters change
  useEffect(() => {
    filterWorkspaces();
  }, [nameFilter, environmentFilter, workspaces]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'locked': return 'orange';
      case 'disabled': return 'red';
      case 'applied': return 'green';
      case 'planned': return 'blue';
      case 'planning': return 'processing';
      case 'applying': return 'processing';
      case 'errored': return 'red';
      case 'canceled': return 'default';
      default: return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return <CheckCircleOutlined />;
      case 'locked': return <LockOutlined />;
      case 'disabled': return <StopOutlined />;
      case 'applied': return <CheckCircleOutlined />;
      case 'planned': return <InfoCircleOutlined />;
      case 'errored': return <ExclamationCircleOutlined />;
      default: return <InfoCircleOutlined />;
    }
  };

  const workspaceColumns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: TFEWorkspace) => (
        <Space>
          <Text strong>{text}</Text>
          {record.tags && record.tags.length > 0 && record.tags.map(tag => (
            <Tag key={tag} size="small">{tag}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Environment',
      dataIndex: 'environment',
      key: 'environment',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (text: string) => (
        <Tag color={getStatusColor(text)} icon={getStatusIcon(text)}>
          {text}
        </Tag>
      ),
    },
    {
      title: 'Terraform Version',
      dataIndex: 'terraformVersion',
      key: 'terraformVersion',
      render: (text: string) => <Text code>{text}</Text>,
    },
    {
      title: 'Owner',
      dataIndex: 'owner',
      key: 'owner',
    },
    {
      title: 'Last Run',
      dataIndex: 'lastRun',
      key: 'lastRun',
      render: (text: string) => text ? new Date(text).toLocaleString() : 'Never',
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (text: any, record: TFEWorkspace) => (
        <Space>
          <Tooltip title="Execute Plan">
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              size="small"
              onClick={() => executePlan(record.id, record.terraformVersion || '1.5.0')}
              loading={loading}
            />
          </Tooltip>
          <Tooltip title={record.status === 'locked' ? 'Unlock' : 'Lock'}>
            <Button
              icon={record.status === 'locked' ? <UnlockOutlined /> : <LockOutlined />}
              size="small"
              onClick={() => toggleWorkspaceLock(record.id, record.status !== 'locked')}
              loading={loading}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  const runColumns = [
    {
      title: 'Workspace',
      dataIndex: 'workspaceName',
      key: 'workspaceName',
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (text: string) => (
        <Tag color={getStatusColor(text)} icon={getStatusIcon(text)}>
          {text}
        </Tag>
      ),
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: 'Source',
      dataIndex: 'source',
      key: 'source',
      render: (text: string) => <Tag>{text}</Tag>,
    },
    {
      title: 'Terraform Version',
      dataIndex: 'terraformVersion',
      key: 'terraformVersion',
      render: (text: string) => <Text code>{text}</Text>,
    },
    {
      title: 'Created',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
      title: 'Changes',
      dataIndex: 'hasChanges',
      key: 'hasChanges',
      render: (hasChanges: boolean) => (
        <Tag color={hasChanges ? 'orange' : 'green'}>
          {hasChanges ? 'Has Changes' : 'No Changes'}
        </Tag>
      ),
    },
  ];

  const tabItems = [
    {
      key: 'workspaces',
      label: (
        <span>
          <DatabaseOutlined />
          Workspaces
        </span>
      ),
      children: (
        <div>
          <Row justify="space-between" style={{ marginBottom: '16px' }}>
            <Col>
              <Space>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadWorkspaces}
                  loading={loading}
                >
                  Refresh
                </Button>
                <Switch
                  checked={autoRefresh}
                  onChange={setAutoRefresh}
                  checkedChildren="Auto"
                  unCheckedChildren="Manual"
                />
              </Space>
            </Col>
            <Col>
              <Text type="secondary">
                {filteredWorkspaces.length} of {workspaces.length} workspaces
              </Text>
            </Col>
          </Row>
          
          {/* Filter Controls */}
          <Card size="small" style={{ marginBottom: '16px' }}>
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12} md={8}>
                <Input
                  placeholder="Filter by workspace name..."
                  value={nameFilter}
                  onChange={(e) => setNameFilter(e.target.value)}
                  prefix={<SearchOutlined />}
                  allowClear
                  autoCapitalize="none"
                  autoCorrect="off"
                  spellCheck={false}
                />
              </Col>
              <Col xs={24} sm={12} md={8}>
                <Select
                  placeholder="Filter by environment..."
                  value={environmentFilter}
                  onChange={setEnvironmentFilter}
                  allowClear
                  style={{ width: '100%' }}
                >
                  <Option value="">All Environments</Option>
                  {getUniqueEnvironments().map(env => (
                    <Option key={env} value={env}>
                      {env}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col xs={24} sm={24} md={8}>
                <Space>
                  <Text>Page size:</Text>
                  <Select
                    value={pageSize}
                    onChange={setPageSize}
                    style={{ width: 80 }}
                  >
                    <Option value={10}>10</Option>
                    <Option value={20}>20</Option>
                    <Option value={50}>50</Option>
                    <Option value={100}>100</Option>
                  </Select>
                </Space>
              </Col>
            </Row>
          </Card>

          <Table
            dataSource={filteredWorkspaces}
            columns={workspaceColumns}
            rowKey="id"
            loading={loading}
            pagination={{ 
              pageSize: pageSize,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} workspaces`,
              pageSizeOptions: ['10', '20', '50', '100'],
              onShowSizeChange: (current, size) => setPageSize(size)
            }}
            size="small"
          />
        </div>
      ),
    },
    {
      key: 'runs',
      label: (
        <span>
          <PlayCircleOutlined />
          Runs
        </span>
      ),
      children: (
        <div>
          <Row justify="space-between" style={{ marginBottom: '16px' }}>
            <Col>
              <Space>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadRuns}
                  loading={loading}
                >
                  Refresh
                </Button>
              </Space>
            </Col>
            <Col>
              <Text type="secondary">
                {runs.length} runs
              </Text>
            </Col>
          </Row>
          <Table
            dataSource={runs}
            columns={runColumns}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 10 }}
            size="small"
          />
        </div>
      ),
    },
    {
      key: 'versions',
      label: (
        <span>
          <FileTextOutlined />
          Versions
        </span>
      ),
      children: (
        <div style={{ padding: '24px' }}>
          <Alert
            message="Version Management"
            description="This section will provide Terraform version management capabilities including checking deprecated versions and updating workspace versions."
            type="info"
            showIcon
          />
        </div>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Row justify="space-between" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2}>
            <CloudServerOutlined style={{ marginRight: '8px' }} />
            Terraform Enterprise
            <Tag color="orange" style={{ marginLeft: '8px' }}>BETA</Tag>
          </Title>
          <Text type="secondary">Manage TFE workspaces, runs, and versions</Text>
        </Col>
        <Col>
          <Space>
            <Button
              icon={<SettingOutlined />}
              onClick={() => setShowConfigModal(true)}
            >
              Configure
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

      <Card>
        <Space direction="vertical" style={{ width: '100%', marginBottom: '16px' }}>
          <Row justify="space-between">
            <Col>
              <Text strong>Endpoint:</Text> <Text code>{config.endpoint}</Text>
            </Col>
            <Col>
              <Text strong>Organization:</Text> <Text code>{config.organization}</Text>
            </Col>
          </Row>
        </Space>

        <Tabs
          activeKey={activeTab}
          onChange={(key) => setActiveTab(key as any)}
          items={tabItems}
        />
      </Card>

      {/* Configuration Modal */}
      <Modal
        title="TFE Configuration"
        open={showConfigModal}
        onCancel={() => setShowConfigModal(false)}
        footer={null}
        width={600}
      >
        <Form
          form={configForm}
          layout="vertical"
          onFinish={handleConfigSave}
          initialValues={config}
        >
          <Alert
            message="TFE Configuration"
            description="Configure your Terraform Enterprise endpoint and organization. Changes will be applied immediately."
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          
          <Form.Item
            name="endpoint"
            label="TFE Endpoint"
            rules={[{ required: true, message: 'Please enter the TFE endpoint' }]}
          >
            <Input
              placeholder="e.g., tfe.doctolib.net or app.terraform.io"
              prefix={<CloudServerOutlined />}
            />
          </Form.Item>
          
          <Form.Item
            name="organization"
            label="Organization"
            rules={[{ required: true, message: 'Please enter the organization name' }]}
          >
            <Input
              placeholder="e.g., doctolib"
              prefix={<DatabaseOutlined />}
            />
          </Form.Item>
          
          <Form.Item
            name="token"
            label="TFE Token (Optional)"
            help="Token can also be set via TFE_TOKEN environment variable"
          >
            <Input.Password
              placeholder="Enter TFE token (optional)"
              prefix={<LockOutlined />}
            />
          </Form.Item>
          
          <Form.Item style={{ marginBottom: 0 }}>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                Save Configuration
              </Button>
              <Button onClick={() => setShowConfigModal(false)}>
                Cancel
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TFE;