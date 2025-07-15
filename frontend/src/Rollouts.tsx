import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Button, 
  Space, 
  Typography, 
  Tag, 
  Input, 
  Select, 
  Alert, 
  Spin, 
  Row, 
  Col, 
  Modal,
  Descriptions,
  Divider,
  message,
  Tooltip,
  Switch
} from 'antd';
import {
  ReloadOutlined,
  PauseCircleOutlined,
  StopOutlined,
  PlayCircleOutlined,
  UpOutlined,
  FastForwardOutlined,
  EyeOutlined,
  CopyOutlined,
  DatabaseOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;
const { Option } = Select;
const { Search } = Input;

// Types matching the Go backend
interface RolloutListItem {
  name: string;
  namespace: string;
  status: string;
  replicas: string;
  age: string;
  strategy: string;
  revision: string;
  images: Record<string, string>;
}

interface RolloutStatus {
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
}

interface KubernetesConfig {
  server: string;
  namespace: string;
}

const StatusTag: React.FC<{ status: string, type: 'status' | 'strategy' }> = ({ status, type }) => {
  const getStatusColor = () => {
    if (type === 'status') {
      switch (status.toLowerCase()) {
        case 'healthy': return 'success';
        case 'progressing': return 'processing';
        case 'degraded': return 'error';
        case 'paused': return 'warning';
        case 'error': return 'error';
        default: return 'default';
      }
    } else if (type === 'strategy') {
      switch (status.toLowerCase()) {
        case 'canary': return 'purple';
        case 'bluegreen': return 'cyan';
        default: return 'default';
      }
    }
    return 'default';
  };

  return <Tag color={getStatusColor()}>{status}</Tag>;
};

// Detailed Rollout Modal Component
const RolloutDetailModal: React.FC<{ 
  rollout: RolloutListItem; 
  config: KubernetesConfig;
  visible: boolean;
  onClose: () => void;
}> = ({ rollout, config, visible, onClose }) => {
  const [detailedStatus, setDetailedStatus] = useState<RolloutStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadDetailedStatus = async () => {
    if (!visible || !rollout?.name) return;
    
    setLoading(true);
    setError(null);
    try {
      // Use the rollout's specific namespace, not the global config namespace
      const rolloutConfig = {
        ...config,
        namespace: rollout.namespace || config.namespace
      };
      const status = await window.go.main.App.GetRolloutStatus(rolloutConfig, rollout.name);
      setDetailedStatus(status);
    } catch (error) {
      setError(`Failed to load detailed status: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible) {
      loadDetailedStatus();
    }
  }, [visible, rollout?.name]);

  const handleCopyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      message.success('Copied to clipboard!');
    } catch (err) {
      message.error('Failed to copy to clipboard');
    }
  };

  return (
    <Modal
      title={
        <Space>
          <DatabaseOutlined />
          <span>Rollout Details: {rollout.name}</span>
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
          <div style={{ marginTop: '16px' }}>Loading detailed status...</div>
        </div>
      ) : error ? (
        <Alert
          message="Error"
          description={error}
          type="error"
          showIcon
          action={
            <Button size="small" onClick={loadDetailedStatus}>
              Retry
            </Button>
          }
        />
      ) : detailedStatus ? (
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="Status">
              <StatusTag status={detailedStatus.status} type="status" />
            </Descriptions.Item>
            <Descriptions.Item label="Strategy">
              <StatusTag status={detailedStatus.strategy} type="strategy" />
            </Descriptions.Item>
            <Descriptions.Item label="Replicas">
              <Text code>{detailedStatus.replicas}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Updated">
              <Text code>{detailedStatus.updated}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Ready">
              <Text code>{detailedStatus.ready}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Available">
              <Text code>{detailedStatus.available}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Current Step">
              <Text code>{detailedStatus.currentStep || 'N/A'}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Revision">
              <Text code>{detailedStatus.revision}</Text>
            </Descriptions.Item>
          </Descriptions>

          {detailedStatus.message && (
            <Card type="inner" title="Message" size="small">
              <Text>{detailedStatus.message}</Text>
            </Card>
          )}

          {detailedStatus.analysis && (
            <Card type="inner" title="Analysis" size="small">
              <Text>{detailedStatus.analysis}</Text>
            </Card>
          )}

          <Card type="inner" title="Container Images" size="small">
            <Space direction="vertical" style={{ width: '100%' }}>
              {Object.entries(detailedStatus.images).map(([container, image]) => (
                <Row key={container} justify="space-between" align="middle">
                  <Col>
                    <Text strong>{container}:</Text>
                  </Col>
                  <Col flex="auto" style={{ marginLeft: '8px' }}>
                    <Text code style={{ wordBreak: 'break-all', fontSize: '12px' }}>
                      {image}
                    </Text>
                  </Col>
                  <Col>
                    <Button
                      type="text"
                      icon={<CopyOutlined />}
                      size="small"
                      onClick={() => handleCopyToClipboard(image)}
                    />
                  </Col>
                </Row>
              ))}
            </Space>
          </Card>
        </Space>
      ) : null}
    </Modal>
  );
};

// Rollout Card Component
const RolloutCard: React.FC<{
  rollout: RolloutListItem;
  config: KubernetesConfig;
  onAction: () => void;
}> = ({ rollout, config, onAction }) => {
  const [loading, setLoading] = useState(false);
  const [showDetails, setShowDetails] = useState(false);

  const handleAction = async (action: string) => {
    setLoading(true);
    try {
      // Use the rollout's specific namespace, not the global config namespace
      const rolloutConfig = {
        ...config,
        namespace: rollout.namespace || config.namespace
      };
      
      switch (action) {
        case 'promote':
          await window.go.main.App.PromoteRollout(rolloutConfig, rollout.name, false);
          break;
        case 'promote-full':
          await window.go.main.App.PromoteRollout(rolloutConfig, rollout.name, true);
          break;
        case 'pause':
          await window.go.main.App.PauseRollout(rolloutConfig, rollout.name);
          break;
        case 'abort':
          await window.go.main.App.AbortRollout(rolloutConfig, rollout.name);
          break;
        case 'restart':
          await window.go.main.App.RestartRollout(rolloutConfig, rollout.name);
          break;
      }
      onAction();
      message.success(`${action} completed successfully`);
    } catch (error) {
      message.error(`${action} failed: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Card
        title={
          <Space>
            <DatabaseOutlined />
            <span>{rollout.name}</span>
            <Tag>{rollout.namespace}</Tag>
          </Space>
        }
        extra={
          <Tooltip title="View details">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => setShowDetails(true)}
              size="small"
            />
          </Tooltip>
        }
        size="small"
        style={{ marginBottom: 16 }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Row justify="space-between">
            <Col>
              <Text type="secondary">Status:</Text>
            </Col>
            <Col>
              <StatusTag status={rollout.status} type="status" />
            </Col>
          </Row>
          
          <Row justify="space-between">
            <Col>
              <Text type="secondary">Strategy:</Text>
            </Col>
            <Col>
              <StatusTag status={rollout.strategy} type="strategy" />
            </Col>
          </Row>
          
          <Row justify="space-between">
            <Col>
              <Text type="secondary">Replicas:</Text>
            </Col>
            <Col>
              <Text code>{rollout.replicas}</Text>
            </Col>
          </Row>
          
          <Row justify="space-between">
            <Col>
              <Text type="secondary">Revision:</Text>
            </Col>
            <Col>
              <Text code>{rollout.revision}</Text>
            </Col>
          </Row>
          
          <Row justify="space-between">
            <Col>
              <Text type="secondary">Age:</Text>
            </Col>
            <Col>
              <Text>{rollout.age}</Text>
            </Col>
          </Row>
          
          <Divider style={{ margin: '12px 0' }} />
          
          <Space wrap>
            <Tooltip title="Promote rollout to next step">
              <Button
                type="primary"
                size="small"
                onClick={() => handleAction('promote')}
                loading={loading}
                icon={<UpOutlined />}
              >
                Promote
              </Button>
            </Tooltip>
            <Tooltip title="Promote rollout to completion">
              <Button
                size="small"
                onClick={() => handleAction('promote-full')}
                loading={loading}
                icon={<FastForwardOutlined />}
              >
                Full Promote
              </Button>
            </Tooltip>
            <Tooltip title="Pause rollout deployment">
              <Button
                size="small"
                onClick={() => handleAction('pause')}
                loading={loading}
                icon={<PauseCircleOutlined />}
              >
                Pause
              </Button>
            </Tooltip>
            <Tooltip title="Abort rollout and revert to stable version">
              <Button
                size="small"
                onClick={() => handleAction('abort')}
                loading={loading}
                icon={<StopOutlined />}
                danger
              >
                Abort
              </Button>
            </Tooltip>
            <Tooltip title="Restart rollout pods">
              <Button
                size="small"
                onClick={() => handleAction('restart')}
                loading={loading}
                icon={<ReloadOutlined />}
              >
                Restart
              </Button>
            </Tooltip>
          </Space>
        </Space>
      </Card>

      <RolloutDetailModal
        rollout={rollout}
        config={config}
        visible={showDetails}
        onClose={() => setShowDetails(false)}
      />
    </>
  );
};

// Main Rollouts Component
const Rollouts: React.FC = () => {
  const [rollouts, setRollouts] = useState<RolloutListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [config, setConfig] = useState<KubernetesConfig>({
    server: '',
    namespace: ''
  });

  const loadRollouts = async () => {
    // Check if Wails bindings are available
    if (!window.go || !window.go.main || !window.go.main.App) {
      setError('Wails bindings not available. Please wait for the app to fully load.');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const rolloutsData = await window.go.main.App.GetRollouts(config);
      
      if (Array.isArray(rolloutsData)) {
        setRollouts(rolloutsData);
      } else {
        setError(`Expected array, got ${typeof rolloutsData}`);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error('Failed to load rollouts:', error);
      setError(`Failed to load rollouts: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  // Load rollouts when config changes
  useEffect(() => {
    const checkWailsAndLoad = () => {
      if (window.go && window.go.main && window.go.main.App) {
        loadRollouts();
      } else {
        // If Wails isn't ready, wait a bit and try again
        setTimeout(checkWailsAndLoad, 100);
      }
    };
    
    checkWailsAndLoad();
  }, [config]);

  // Auto-refresh rollouts when enabled
  useEffect(() => {
    if (autoRefresh && config.server) {
      const interval = setInterval(loadRollouts, 30000); // Refresh every 30 seconds
      return () => clearInterval(interval);
    }
  }, [config.server, autoRefresh]);

  const filteredRollouts = rollouts.filter(rollout =>
    rollout.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    rollout.namespace.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div style={{ padding: '24px' }}>
      <Row justify="space-between" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2}>
            <DatabaseOutlined style={{ marginRight: '8px' }} />
            Argo Rollouts
          </Title>
          <Text type="secondary">Manage your Argo Rollouts</Text>
        </Col>
        <Col>
          <Space>
            <Switch
              checked={autoRefresh}
              onChange={setAutoRefresh}
              checkedChildren="Auto"
              unCheckedChildren="Manual"
            />
            <Button 
              icon={<ReloadOutlined />}
              onClick={loadRollouts}
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
        <Space style={{ width: '100%' }}>
          <Text>
            <Text strong>Namespace:</Text> {config.namespace || 'All namespaces'}
          </Text>
          <Search
            placeholder="Search rollouts..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{ width: '200px' }}
          />
        </Space>
      </Card>

      <div style={{ minHeight: '400px' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <Spin size="large" />
            <div style={{ marginTop: '16px' }}>Loading rollouts...</div>
          </div>
        ) : filteredRollouts.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <DatabaseOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
            <div style={{ marginTop: '16px', color: '#999' }}>
              No rollouts found
            </div>
          </div>
        ) : (
          <Row gutter={[16, 16]}>
            {filteredRollouts.map((rollout, index) => (
              <Col xs={24} sm={12} md={8} lg={6} key={rollout?.name || `rollout-${index}`}>
                <RolloutCard
                  rollout={rollout}
                  config={config}
                  onAction={loadRollouts}
                />
              </Col>
            ))}
          </Row>
        )}
      </div>
    </div>
  );
};

export default Rollouts;