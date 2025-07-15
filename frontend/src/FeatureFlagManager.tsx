import React from 'react';
import { Card, Switch, Typography, Space, Divider, Row, Col, Button, Alert } from 'antd';
import { ExperimentOutlined, ReloadOutlined, ClearOutlined } from '@ant-design/icons';
import { useFeatureFlags, defaultFeatureFlags, type FeatureFlags } from './featureFlags';

const { Title, Text } = Typography;

interface FeatureFlagManagerProps {
  onFlagsChange?: (flags: FeatureFlags) => void;
}

const FeatureFlagManager: React.FC<FeatureFlagManagerProps> = ({ onFlagsChange }) => {
  const [flags, updateFlags] = useFeatureFlags();

  const handleFlagChange = (flagName: keyof FeatureFlags, value: boolean) => {
    const updates = { [flagName]: value };
    updateFlags(updates);
    if (onFlagsChange) {
      onFlagsChange({ ...flags, ...updates });
    }
  };

  const handleResetToDefaults = () => {
    updateFlags(defaultFeatureFlags);
    if (onFlagsChange) {
      onFlagsChange(defaultFeatureFlags);
    }
  };

  const handleClearAll = () => {
    localStorage.removeItem('yakgui_feature_flags');
    window.location.reload();
  };

  const tabFlags = [
    { key: 'showEnvironmentTab', label: 'Environment Tab', description: 'Environment configuration and profiles' },
    { key: 'showArgoCDTab', label: 'ArgoCD Tab', description: 'ArgoCD applications management' },
    { key: 'showRolloutsTab', label: 'Rollouts Tab', description: 'Argo Rollouts management' },
    { key: 'showSecretsTab', label: 'Secrets Tab', description: 'Secrets management' },
    { key: 'showCertificatesTab', label: 'Certificates Tab', description: 'SSL certificate management' },
    { key: 'showTFETab', label: 'TFE Tab (Experimental)', description: 'Terraform Enterprise management' },
  ] as const;

  const featureFlags = [
    { key: 'enableAutoRefresh', label: 'Auto Refresh', description: 'Enable automatic refresh functionality' },
    { key: 'enableDarkMode', label: 'Dark Mode', description: 'Enable dark mode theme support' },
    { key: 'enableDetailedLogging', label: 'Detailed Logging', description: 'Enable detailed console logging' },
  ] as const;

  return (
    <Card 
      title={
        <Space>
          <ExperimentOutlined />
          Feature Flags Manager
        </Space>
      }
      extra={
        <Space>
          <Button 
            icon={<ReloadOutlined />}
            onClick={handleResetToDefaults}
            size="small"
          >
            Reset to Defaults
          </Button>
          <Button 
            icon={<ClearOutlined />}
            onClick={handleClearAll}
            size="small"
            danger
          >
            Clear All & Reload
          </Button>
        </Space>
      }
    >
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        <Alert
          message="Development Feature Flags"
          description="These flags control which tabs and features are visible in the application. Changes are saved to localStorage and persist between sessions."
          type="info"
          showIcon
        />

        <div>
          <Title level={4}>Tab Visibility</Title>
          <Space direction="vertical" style={{ width: '100%' }}>
            {tabFlags.map(({ key, label, description }) => (
              <Row key={key} justify="space-between" align="middle">
                <Col flex="auto">
                  <Space direction="vertical" size="small">
                    <Text strong>{label}</Text>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      {description}
                    </Text>
                  </Space>
                </Col>
                <Col>
                  <Switch
                    checked={flags[key]}
                    onChange={(value) => handleFlagChange(key, value)}
                    disabled={key === 'showEnvironmentTab' || key === 'showTFETab'} // Environment tab always enabled, TFE tab always disabled
                  />
                </Col>
              </Row>
            ))}
          </Space>
        </div>

        <Divider />

        <div>
          <Title level={4}>Feature Flags</Title>
          <Space direction="vertical" style={{ width: '100%' }}>
            {featureFlags.map(({ key, label, description }) => (
              <Row key={key} justify="space-between" align="middle">
                <Col flex="auto">
                  <Space direction="vertical" size="small">
                    <Text strong>{label}</Text>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      {description}
                    </Text>
                  </Space>
                </Col>
                <Col>
                  <Switch
                    checked={flags[key]}
                    onChange={(value) => handleFlagChange(key, value)}
                  />
                </Col>
              </Row>
            ))}
          </Space>
        </div>

        <Divider />

        <div>
          <Title level={5}>Current Configuration</Title>
          <pre style={{ 
            backgroundColor: '#f5f5f5', 
            padding: '12px', 
            borderRadius: '4px', 
            fontSize: '11px',
            overflow: 'auto',
            maxHeight: '200px'
          }}>
            {JSON.stringify(flags, null, 2)}
          </pre>
        </div>
      </Space>
    </Card>
  );
};

export default FeatureFlagManager;